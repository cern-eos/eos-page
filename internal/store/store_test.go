package store

import (
	"path/filepath"
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

	ws, err := st.ListWorkshops(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) < 9 {
		t.Fatalf("expected 9 workshops, got %d", len(ws))
	}

	cards, err := st.ListCards("", true)
	if err != nil {
		t.Fatal(err)
	}
	var haveCommits, haveStatus, haveOldStatus bool
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
	}
	if !haveCommits || !haveStatus || haveOldStatus {
		t.Fatalf("commit/status cards: commits=%v status=%v old=%v", haveCommits, haveStatus, haveOldStatus)
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
