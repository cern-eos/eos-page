package store

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchTalksAndDocs(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "eos.db"), filepath.Join(dir, "media"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	talks, slides, recs, err := st.TalkCounts()
	if err != nil {
		t.Fatal(err)
	}
	if talks < 300 || slides < 200 || recs < 200 {
		t.Fatalf("index too small: talks=%d slides=%d recs=%d", talks, slides, recs)
	}

	res, err := st.Search("QuarkDB", "talks", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Talks) == 0 {
		t.Fatal("expected QuarkDB talks")
	}

	docs, err := st.Search("FUSE", "docs", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs.Docs) == 0 {
		t.Fatal("expected FUSE documentation hits")
	}

	ext, err := st.Search("NFS", "external", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(ext.Talks) == 0 {
		t.Fatal("expected external NFS presentation")
	}
	if len(ext.Docs) != 0 {
		t.Fatal("external search should not return docs")
	}

	ws, err := st.ListWorkshops(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) < 10 {
		t.Fatalf("expected 10 workshops, got %d", len(ws))
	}
	var haveFirst bool
	for _, w := range ws {
		if w.ID == "591485" && w.Edition == 1 && w.Year == 2017 {
			haveFirst = true
		}
	}
	if !haveFirst {
		t.Fatal("expected first EOS workshop 2017 (591485)")
	}

	cards, err := st.ListCards("", true)
	if err != nil {
		t.Fatal(err)
	}
	var haveCommits, haveStatus, haveOldStatus, havePres, haveWorkshopSearch bool
	for _, c := range cards {
		if c.ID == "r-commits" && c.Href == "/commits" {
			haveCommits = true
		}
		if c.ID == "s-status" && c.Kind == "service" {
			haveStatus = true
		}
		if c.ID == "r-status" {
			haveOldStatus = true
		}
		if c.ID == "r-pres" && c.Href == "/search?kind=external" && c.Title == "External Presentations" {
			havePres = true
		}
		if c.ID == "r-search" && c.Href == "/search?kind=workshop" {
			haveWorkshopSearch = true
		}
	}
	if !haveCommits || !haveStatus || haveOldStatus {
		t.Fatalf("commit/status cards: commits=%v status=%v old=%v", haveCommits, haveStatus, haveOldStatus)
	}
	if !havePres || !haveWorkshopSearch {
		t.Fatalf("resource search cards: external=%v workshop=%v", havePres, haveWorkshopSearch)
	}
	pres, err := st.ListConferenceTalks()
	if err != nil {
		t.Fatal(err)
	}
	if len(pres) < 20 {
		t.Fatalf("expected conference presentations, got %d", len(pres))
	}

	if err := st.UpsertCommits([]Commit{{
		SHA: "abc123def", Date: "2026-09-22T10:00:00Z", Year: 2026,
		Author: "Andreas Peters", Title: "Fix QuarkDB lease", Body: "Keep the master log live.",
		URL: "https://github.com/cern-eos/eos/commit/abc123def",
	}}); err != nil {
		t.Fatal(err)
	}
	hits, err := st.SearchCommits("QuarkDB", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 || hits[0].Title != "Fix QuarkDB lease" {
		t.Fatalf("commit search: %+v", hits)
	}
}

func TestAddListChats(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "eos.db"), filepath.Join(dir, "media"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	if _, err := st.AddChat("", "answer", "gpt-4o-mini", "127.0.0.1"); err != ErrInvalid {
		t.Fatalf("empty question: %v", err)
	}
	got, err := st.AddChat("What is EOS used for?", "EOS Open Storage holds LHC data.", "gpt-4o-mini", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == "" || got.Question != "What is EOS used for?" {
		t.Fatalf("saved %+v", got)
	}
	list, err := st.ListChats(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Answer != "EOS Open Storage holds LHC data." || list[0].Model != "gpt-4o-mini" {
		t.Fatalf("list %+v", list)
	}
}

func TestGitContributorsSeed(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "eos.db"), filepath.Join(dir, "media"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	people, err := st.ListGitContributors()
	if err != nil {
		t.Fatal(err)
	}
	if len(people) < 70 {
		t.Fatalf("expected seeded shortlog, got %d", len(people))
	}
	if people[0].Name != "Elvin Alin Sindrilaru" || people[0].Commits < 9000 {
		t.Fatalf("first %+v", people[0])
	}
	if people[1].Name != "Andreas Joachim Peters" {
		t.Fatalf("second %+v", people[1])
	}
	seen := map[string]int{}
	for _, p := range people {
		if strings.EqualFold(p.Name, "root") || strings.EqualFold(p.Name, "unknown") || strings.HasPrefix(strings.ToLower(p.Email), "root@") {
			t.Fatalf("automation still listed: %+v", p)
		}
		if strings.Contains(strings.ToLower(p.Email), "jenkins") || strings.Contains(strings.ToLower(p.Name), "ci/cd") {
			t.Fatalf("automation still listed: %+v", p)
		}
		key := strings.ToLower(p.Email)
		if key != "" {
			seen[key]++
			if seen[key] > 1 {
				t.Fatalf("duplicate identity %s", p.Email)
			}
		}
	}
}
