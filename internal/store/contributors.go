package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GitContributor struct {
	Sort    int    `json:"sort"`
	Commits int    `json:"commits"`
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
}

func (s *Store) ensureGitContributors() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM git_contributors`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	people, err := loadSeedContributors()
	if err != nil {
		return err
	}
	return s.ReplaceGitContributors(people)
}

func loadSeedContributors() ([]GitContributor, error) {
	raw, err := seedFS.ReadFile("seeddata/contributors.json")
	if err != nil {
		return nil, err
	}
	var file struct {
		Contributors []GitContributor `json:"contributors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, err
	}
	kept := file.Contributors[:0]
	for _, p := range file.Contributors {
		if isRootGitContributor(p.Name, p.Email) {
			continue
		}
		kept = append(kept, p)
	}
	file.Contributors = kept
	for i := range file.Contributors {
		if file.Contributors[i].Sort == 0 {
			file.Contributors[i].Sort = i + 1
		}
	}
	if len(file.Contributors) == 0 {
		return nil, fmt.Errorf("empty contributors seed")
	}
	return file.Contributors, nil
}

func (s *Store) ReplaceGitContributors(people []GitContributor) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM git_contributors`); err != nil {
		return err
	}
	for i, p := range people {
		sort := p.Sort
		if sort == 0 {
			sort = i + 1
		}
		id := fmt.Sprintf("c-%d", sort)
		if _, err := tx.Exec(`INSERT INTO git_contributors(id, sort, commits, name, email) VALUES(?,?,?,?,?)`,
			id, sort, p.Commits, p.Name, p.Email); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListGitContributors() ([]GitContributor, error) {
	rows, err := s.db.Query(`SELECT sort, commits, name, email FROM git_contributors ORDER BY sort, commits DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GitContributor
	for rows.Next() {
		var p GitContributor
		if err := rows.Scan(&p.Sort, &p.Commits, &p.Name, &p.Email); err != nil {
			return nil, err
		}
		if isRootGitContributor(p.Name, p.Email) {
			continue
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func isRootGitContributor(name, email string) bool {
	if strings.EqualFold(strings.TrimSpace(name), "root") {
		return true
	}
	local, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	return local == "root"
}

func (s *Store) GitContributorCount() int {
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM git_contributors`).Scan(&n)
	return n
}
