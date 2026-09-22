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
}
