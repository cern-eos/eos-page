package store

import (
	"fmt"
	"regexp"
	"strings"
)

type TalkHit struct {
	Talk
	WorkshopTitle string  `json:"workshopTitle"`
	Rank          float64 `json:"rank"`
}

type DocHit struct {
	Doc
	Rank float64 `json:"rank"`
}

type SearchResult struct {
	Query string    `json:"query"`
	Kind  string    `json:"kind"`
	Talks []TalkHit `json:"talks"`
	Docs  []DocHit  `json:"docs"`
}

func normalizeSearchKind(kind string) string {
	switch strings.TrimSpace(kind) {
	case "docs":
		return "docs"
	case "workshop", "talks":
		return "workshop"
	case "workshop-docs", "workshop_docs":
		return "workshop-docs"
	case "external", "conference":
		return "external"
	case "all":
		return "workshop-docs"
	default:
		return "presentations"
	}
}

func talkScope(kind string) string {
	switch kind {
	case "workshop", "workshop-docs":
		return "workshop"
	case "external":
		return "external"
	default:
		return ""
	}
}

func talkScopeSQL(scope string) string {
	switch scope {
	case "workshop":
		return `(w.kind IS NULL OR w.kind='' OR w.kind!='conference')`
	case "external":
		return `w.kind='conference'`
	default:
		return ""
	}
}

var tokenRe = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9._+-]{0,40}`)

func ftsQuery(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	toks := tokenRe.FindAllString(raw, 12)
	if len(toks) == 0 {
		return ""
	}
	parts := make([]string, 0, len(toks))
	for _, t := range toks {
		t = strings.ReplaceAll(t, `"`, "")
		if t == "" {
			continue
		}
		if len(t) >= 3 {
			parts = append(parts, t+"*")
		} else {
			parts = append(parts, `"`+t+`"`)
		}
	}
	return strings.Join(parts, " OR ")
}

func (s *Store) ListTalks(year int, scope string) ([]TalkHit, error) {
	query := `
SELECT t.id, t.event_id, t.year, t.title, t.abstract, t.speakers, t.session, t.url, t.slides, t.recording, t.start_date,
       COALESCE(w.title,'')
FROM talks t
LEFT JOIN workshops w ON w.id = t.event_id`
	args := []any{}
	where := []string{}
	if year > 0 {
		where = append(where, `t.year=?`)
		args = append(args, year)
	}
	if extra := talkScopeSQL(scope); extra != "" {
		where = append(where, extra)
	}
	if len(where) > 0 {
		query += ` WHERE ` + strings.Join(where, " AND ")
	}
	query += ` ORDER BY t.year DESC, t.start_date DESC, t.title COLLATE NOCASE`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TalkHit
	for rows.Next() {
		var h TalkHit
		if err := rows.Scan(&h.ID, &h.EventID, &h.Year, &h.Title, &h.Abstract, &h.Speakers, &h.Session, &h.URL, &h.Slides, &h.Recording, &h.Start, &h.WorkshopTitle); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) Search(q, kind string, year int, limit int) (SearchResult, error) {
	if limit <= 0 || limit > 80 {
		limit = 40
	}
	kind = normalizeSearchKind(kind)
	out := SearchResult{Query: q, Kind: kind, Talks: []TalkHit{}, Docs: []DocHit{}}
	match := ftsQuery(q)
	if match == "" {
		if kind != "docs" {
			talks, err := s.ListTalks(year, talkScope(kind))
			if err != nil {
				return out, err
			}
			out.Talks = talks
		}
		return out, nil
	}

	if kind != "docs" {
		query := `
SELECT t.id, t.event_id, t.year, t.title, t.abstract, t.speakers, t.session, t.url, t.slides, t.recording, t.start_date,
       COALESCE(w.title,''), bm25(talks_fts)
FROM talks_fts
JOIN talks t ON t.id = talks_fts.id
LEFT JOIN workshops w ON w.id = t.event_id
WHERE talks_fts MATCH ?`
		args := []any{match}
		if extra := talkScopeSQL(talkScope(kind)); extra != "" {
			query += ` AND ` + extra
		}
		if year > 0 {
			query += ` AND t.year=?`
			args = append(args, year)
		}
		query += ` ORDER BY bm25(talks_fts) LIMIT ?`
		args = append(args, limit)
		rows, err := s.db.Query(query, args...)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var h TalkHit
			if err := rows.Scan(&h.ID, &h.EventID, &h.Year, &h.Title, &h.Abstract, &h.Speakers, &h.Session, &h.URL, &h.Slides, &h.Recording, &h.Start, &h.WorkshopTitle, &h.Rank); err != nil {
				rows.Close()
				return out, err
			}
			out.Talks = append(out.Talks, h)
		}
		rows.Close()
	}

	if kind == "docs" || kind == "workshop-docs" {
		query := `
SELECT d.id, d.page_id, d.title, d.section, d.url, d.summary, d.body, d.sort, bm25(docs_fts)
FROM docs_fts
JOIN docs d ON d.id = docs_fts.id
WHERE docs_fts MATCH ?
ORDER BY bm25(docs_fts)
LIMIT ?`
		rows, err := s.db.Query(query, match, limit)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var h DocHit
			if err := rows.Scan(&h.ID, &h.PageID, &h.Title, &h.Section, &h.URL, &h.Summary, &h.Body, &h.Sort, &h.Rank); err != nil {
				rows.Close()
				return out, err
			}
			if len(h.Body) > 600 {
				h.Body = h.Body[:600] + "…"
			}
			out.Docs = append(out.Docs, h)
		}
		rows.Close()
	}

	if kind == "docs" {
		out.Talks = []TalkHit{}
	}
	return out, nil
}

func (s *Store) WorkshopYears() ([]int, error) {
	rows, err := s.db.Query(`SELECT DISTINCT year FROM workshops WHERE visible=1 ORDER BY year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		out = append(out, y)
	}
	return out, rows.Err()
}

func (s *Store) IndexStats() map[string]any {
	talks, slides, recs, _ := s.TalkCounts()
	docs, _ := s.DocCount()
	return map[string]any{
		"talks":      talks,
		"slides":     slides,
		"recordings": recs,
		"docs":       docs,
		"queryHelp":  fmt.Sprintf("%d talks · %d slides · %d recordings · %d doc sections", talks, slides, recs, docs),
	}
}
