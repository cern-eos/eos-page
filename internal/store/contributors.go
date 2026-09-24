package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type GitContributor struct {
	Sort     int    `json:"sort"`
	Commits  int    `json:"commits"`
	Assisted int    `json:"assisted,omitempty"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
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
	file.Contributors = normalizeSeedContributors(file.Contributors)
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
		if _, err := tx.Exec(`INSERT INTO git_contributors(id, sort, commits, assisted, name, email) VALUES(?,?,?,?,?,?)`,
			id, sort, p.Commits, p.Assisted, p.Name, p.Email); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListGitContributors() ([]GitContributor, error) {
	rows, err := s.db.Query(`SELECT sort, commits, assisted, name, email FROM git_contributors ORDER BY sort, commits DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GitContributor
	for rows.Next() {
		var p GitContributor
		if err := rows.Scan(&p.Sort, &p.Commits, &p.Assisted, &p.Name, &p.Email); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func normalizeSeedContributors(people []GitContributor) []GitContributor {
	type bucket struct {
		name, email string
		commits     int
		assisted    int
	}
	merged := map[string]*bucket{}
	order := []string{}
	for _, p := range people {
		if seedAutomation(p.Name, p.Email) {
			continue
		}
		name, email := seedCanonical(p.Name, p.Email)
		if name == "" {
			continue
		}
		key := strings.ToLower(email)
		if key == "" {
			key = "name:" + strings.ToLower(name)
		}
		if b, ok := merged[key]; ok {
			b.commits += p.Commits
			b.assisted += p.Assisted
			continue
		}
		merged[key] = &bucket{name: name, email: email, commits: p.Commits, assisted: p.Assisted}
		order = append(order, key)
	}
	out := make([]GitContributor, 0, len(order))
	for _, key := range order {
		b := merged[key]
		out = append(out, GitContributor{Name: b.name, Email: b.email, Commits: b.commits, Assisted: b.assisted})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Commits != out[j].Commits {
			return out[i].Commits > out[j].Commits
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	for i := range out {
		out[i].Sort = i + 1
	}
	return out
}

var seedAliases = map[string][2]string{
	"elvin.alin.sindrilaru@cern.ch": {"Elvin Alin Sindrilaru", "elvin.alin.sindrilaru@cern.ch"},
	"esindrl@cern.ch":               {"Elvin Alin Sindrilaru", "elvin.alin.sindrilaru@cern.ch"},
	"elvin.sindrilaru@gmail.com":    {"Elvin Alin Sindrilaru", "elvin.alin.sindrilaru@cern.ch"},
	"david.smith@cern.ch":           {"David Smith", "david.smith@cern.ch"},
	"luis.antonio.obis@gmail.com":   {"Luis Antonio Obis Aparicio", "luis.obis@cern.ch"},
	"luis.obis@cern.ch":             {"Luis Antonio Obis Aparicio", "luis.obis@cern.ch"},
	"amadio@cern.ch":                {"Guilherme Amadio", "amadio@cern.ch"},
	"guilherme@amadio.org":          {"Guilherme Amadio", "amadio@cern.ch"},
	"niels.alexander.bugel@cern.ch": {"Niels Alexander Buegel", "niels.alexander.bugel@cern.ch"},
	"bugel.niels@gmail.com":         {"Niels Alexander Buegel", "niels.alexander.bugel@cern.ch"},
	"rptaylor@uvic.ca":              {"Ryan Taylor", "rptaylor@uvic.ca"},
	"jgeens@cern.ch":                {"Jesse Geens", "jgeens@cern.ch"},
	"jesse.geens@gmail.com":         {"Jesse Geens", "jgeens@cern.ch"},
	"pablo.oliver.cortes@cern.ch":   {"Pablo Oliver Cortés", "pablo.oliver.cortes@cern.ch"},
}

func seedCanonical(name, email string) (string, string) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if pair, ok := seedAliases[strings.ToLower(email)]; ok {
		return pair[0], pair[1]
	}
	switch strings.ToLower(name) {
	case "kaehatah":
		name = "Karl Ehataht"
	case "okilicki":
		name = "Ozlem Kilickiran"
	}
	return name, email
}

func seedAutomation(name, email string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	e := strings.ToLower(strings.TrimSpace(email))
	local, _, _ := strings.Cut(e, "@")
	if n == "root" || n == "unknown" || local == "root" {
		return true
	}
	if local == "jenkins" || n == "mr jenkins" {
		return true
	}
	if strings.Contains(n, "ci/cd") || strings.Contains(e, "service_account") {
		return true
	}
	return false
}

func (s *Store) GitContributorCount() int {
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM git_contributors`).Scan(&n)
	return n
}
