package ingest

import (
	"testing"

	"github.com/apeters/eospage/internal/store"
)

func TestParseGitShortlog(t *testing.T) {
	raw := `
  8760	Elvin Alin Sindrilaru <elvin.alin.sindrilaru@cern.ch>
  9238	Andreas Joachim Peters <andreas.joachim.peters@cern.ch>
    35	Elvin Sindrilaru <esindrl@cern.ch>
     6	Elvin Sindrilaru <elvin.sindrilaru@gmail.com>
    25	Unknown <root@vmeos03.cern.ch>
     2	Duo Fix CI/CD Pipeline <service_account_group_629_0e695a98b99b4dfe646c0d85e8b7c991@noreply.gitlab.cern.ch>
     1	Mr Jenkins <jenkins@pb-d-128-141-194-219.cern.ch>
     1	okilicki <ozlem.kilickiran@cern.ch>
     1	kaehatah <karl.ehataht@cern.ch>
`
	got := ParseGitShortlog(raw)
	if len(got) != 4 {
		t.Fatalf("len %d: %+v", len(got), got)
	}
	if got[0].Name != "Andreas Joachim Peters" || got[0].Commits != 9238 {
		t.Fatalf("first %+v", got[0])
	}
	if got[1].Name != "Elvin Alin Sindrilaru" || got[1].Email != "elvin.alin.sindrilaru@cern.ch" || got[1].Commits != 8801 {
		t.Fatalf("elvin %+v", got[1])
	}
	if got[2].Name != "Karl Ehataht" || got[2].Email != "karl.ehataht@cern.ch" {
		t.Fatalf("karl %+v", got[2])
	}
	if got[3].Name != "Ozlem Kilickiran" {
		t.Fatalf("ozlem %+v", got[3])
	}
}

func TestNormalizeMergesAndDropsAutomation(t *testing.T) {
	got := NormalizeGitContributors([]store.GitContributor{
		c("Elvin Alin Sindrilaru", "elvin.alin.sindrilaru@cern.ch", 8760),
		c("Elvin Sindrilaru", "esindrl@cern.ch", 35),
		c("Elvin Sindrilaru", "elvin.sindrilaru@gmail.com", 6),
		c("David Smith", "david.smith@cern.ch", 105),
		c("David Smith", "David.Smith@cern.ch", 46),
		c("Luis Antonio Obis Aparicio", "luis.antonio.obis@gmail.com", 536),
		c("Luis Antonio Obis Aparicio", "luis.obis@cern.ch", 15),
		c("Guilherme Amadio", "amadio@cern.ch", 142),
		c("Guilherme Amadio", "guilherme@amadio.org", 2),
		c("Niels Alexander Buegel", "niels.alexander.bugel@cern.ch", 14),
		c("Niels Bugel", "niels.alexander.bugel@cern.ch", 3),
		c("Niels Bugel", "bugel.niels@gmail.com", 1),
		c("Ryan Taylor", "rptaylor@uvic.ca", 33),
		c("R. P. Taylor", "rptaylor@uvic.ca", 2),
		c("Jesse Geens", "jgeens@cern.ch", 3),
		c("Jesse Geens", "jesse.geens@gmail.com", 2),
		c("Pablo Oliver Cortés", "pablo.oliver.cortes@cern.ch", 8),
		c("Pablo Oliver Cortes", "pablo.oliver.cortes@cern.ch", 1),
		c("Unknown", "root@vmeos03.cern.ch", 25),
		c("root", "root@elvin-dev01.cern.ch", 1),
		c("Duo Fix CI/CD Pipeline", "service_account_group_629_0e695a98b99b4dfe646c0d85e8b7c991@noreply.gitlab.cern.ch", 2),
		c("Mr Jenkins", "jenkins@pb-d-128-141-194-219.cern.ch", 1),
		c("Someone Else", "other@cern.ch", 10),
	})
	want := map[string]int{
		"elvin.alin.sindrilaru@cern.ch": 8801,
		"david.smith@cern.ch":           151,
		"luis.obis@cern.ch":             551,
		"amadio@cern.ch":                144,
		"niels.alexander.bugel@cern.ch": 18,
		"rptaylor@uvic.ca":              35,
		"jgeens@cern.ch":                5,
		"pablo.oliver.cortes@cern.ch":   9,
		"other@cern.ch":                 10,
	}
	if len(got) != len(want) {
		t.Fatalf("len %d want %d: %+v", len(got), len(want), got)
	}
	seen := map[string]int{}
	for _, p := range got {
		seen[p.Email]++
		if want[p.Email] != p.Commits {
			t.Fatalf("%s commits %d want %d", p.Email, p.Commits, want[p.Email])
		}
	}
	for email := range want {
		if seen[email] != 1 {
			t.Fatalf("%s appeared %d times", email, seen[email])
		}
	}
}

func c(name, email string, n int) store.GitContributor {
	return store.GitContributor{Name: name, Email: email, Commits: n}
}
