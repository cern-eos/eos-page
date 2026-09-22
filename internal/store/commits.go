package store

import (
	"strings"
)

type Commit struct {
	SHA    string `json:"sha"`
	Short  string `json:"short"`
	Date   string `json:"date"`
	Year   int    `json:"year"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
}

func (s *Store) CommitCount() int {
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM commits`).Scan(&n)
	return n
}

func (s *Store) UpsertCommits(commits []Commit) error {
	if len(commits) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ins, err := tx.Prepare(`
INSERT INTO commits(sha, short, date, year, author, title, body, url)
VALUES(?,?,?,?,?,?,?,?)
ON CONFLICT(sha) DO UPDATE SET
  short=excluded.short, date=excluded.date, year=excluded.year, author=excluded.author,
  title=excluded.title, body=excluded.body, url=excluded.url`)
	if err != nil {
		return err
	}
	defer ins.Close()
	delFTS, err := tx.Prepare(`DELETE FROM commits_fts WHERE sha=?`)
	if err != nil {
		return err
	}
	defer delFTS.Close()
	fts, err := tx.Prepare(`INSERT INTO commits_fts(sha, title, body, author) VALUES(?,?,?,?)`)
	if err != nil {
		return err
	}
	defer fts.Close()
	for _, c := range commits {
		if c.SHA == "" {
			continue
		}
		if c.Short == "" {
			c.Short = shortSHA(c.SHA)
		}
		if _, err := ins.Exec(c.SHA, c.Short, c.Date, c.Year, c.Author, c.Title, c.Body, c.URL); err != nil {
			return err
		}
		if _, err := delFTS.Exec(c.SHA); err != nil {
			return err
		}
		if _, err := fts.Exec(c.SHA, c.Title, c.Body, c.Author); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListCommits(limit int) ([]Commit, error) {
	if limit <= 0 || limit > 400 {
		limit = 200
	}
	rows, err := s.db.Query(`
SELECT sha, short, date, year, author, title, body, url
FROM commits
ORDER BY date DESC, sha
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommits(rows)
}

func (s *Store) SearchCommits(q string, limit int) ([]Commit, error) {
	if limit <= 0 || limit > 400 {
		limit = 200
	}
	match := ftsQuery(q)
	if match == "" {
		return s.ListCommits(limit)
	}
	rows, err := s.db.Query(`
SELECT c.sha, c.short, c.date, c.year, c.author, c.title, c.body, c.url
FROM commits_fts
JOIN commits c ON c.sha = commits_fts.sha
WHERE commits_fts MATCH ?
ORDER BY c.date DESC, c.sha
LIMIT ?`, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommits(rows)
}

func scanCommits(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]Commit, error) {
	out := []Commit{}
	for rows.Next() {
		var c Commit
		if err := rows.Scan(&c.SHA, &c.Short, &c.Date, &c.Year, &c.Author, &c.Title, &c.Body, &c.URL); err != nil {
			return nil, err
		}
		if len(c.Body) > 600 {
			c.Body = strings.TrimSpace(c.Body[:600]) + "…"
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
