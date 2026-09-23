package ingest

import (
	"sort"
	"strings"

	"github.com/apeters/eospage/internal/store"
)

type identity struct {
	Name  string
	Email string
}

// Alias emails map to one canonical name+email. Only listed aliases or the
// same address (any case) are merged.
var contributorAliases = map[string]identity{
	"elvin.alin.sindrilaru@cern.ch": {Name: "Elvin Alin Sindrilaru", Email: "elvin.alin.sindrilaru@cern.ch"},
	"esindrl@cern.ch":               {Name: "Elvin Alin Sindrilaru", Email: "elvin.alin.sindrilaru@cern.ch"},
	"elvin.sindrilaru@gmail.com":    {Name: "Elvin Alin Sindrilaru", Email: "elvin.alin.sindrilaru@cern.ch"},
	"david.smith@cern.ch":           {Name: "David Smith", Email: "david.smith@cern.ch"},
	"luis.antonio.obis@gmail.com":   {Name: "Luis Antonio Obis Aparicio", Email: "luis.obis@cern.ch"},
	"luis.obis@cern.ch":             {Name: "Luis Antonio Obis Aparicio", Email: "luis.obis@cern.ch"},
	"amadio@cern.ch":                {Name: "Guilherme Amadio", Email: "amadio@cern.ch"},
	"guilherme@amadio.org":          {Name: "Guilherme Amadio", Email: "amadio@cern.ch"},
	"niels.alexander.bugel@cern.ch": {Name: "Niels Alexander Buegel", Email: "niels.alexander.bugel@cern.ch"},
	"bugel.niels@gmail.com":         {Name: "Niels Alexander Buegel", Email: "niels.alexander.bugel@cern.ch"},
	"rptaylor@uvic.ca":              {Name: "Ryan Taylor", Email: "rptaylor@uvic.ca"},
	"jgeens@cern.ch":                {Name: "Jesse Geens", Email: "jgeens@cern.ch"},
	"jesse.geens@gmail.com":         {Name: "Jesse Geens", Email: "jgeens@cern.ch"},
	"pablo.oliver.cortes@cern.ch":   {Name: "Pablo Oliver Cortés", Email: "pablo.oliver.cortes@cern.ch"},
}

// Usernames verified against CERN GitLab / public CERN identity.
var verifiedNames = map[string]string{
	"kaehatah": "Karl Ehataht",
	"okilicki": "Ozlem Kilickiran",
}

func NormalizeGitContributors(people []store.GitContributor) []store.GitContributor {
	type bucket struct {
		name    string
		email   string
		commits int
	}
	merged := map[string]*bucket{}
	order := []string{}
	for _, p := range people {
		if isAutomationIdentity(p.Name, p.Email) {
			continue
		}
		name, email := canonicalIdentity(p.Name, p.Email)
		if name == "" {
			continue
		}
		key := strings.ToLower(email)
		if key == "" {
			key = "name:" + strings.ToLower(name)
		}
		if b, ok := merged[key]; ok {
			b.commits += p.Commits
			continue
		}
		merged[key] = &bucket{name: name, email: email, commits: p.Commits}
		order = append(order, key)
	}
	out := make([]store.GitContributor, 0, len(order))
	for _, key := range order {
		b := merged[key]
		out = append(out, store.GitContributor{Name: b.name, Email: b.email, Commits: b.commits})
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

func canonicalIdentity(name, email string) (string, string) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if canon, ok := contributorAliases[strings.ToLower(email)]; ok {
		return canon.Name, canon.Email
	}
	if pretty, ok := verifiedNames[strings.ToLower(name)]; ok {
		name = pretty
	}
	return name, email
}

func isAutomationIdentity(name, email string) bool {
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
