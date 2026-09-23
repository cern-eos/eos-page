package chat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitAndSearchOpsDocs(t *testing.T) {
	raw := `
================================================================================
FILE: repairs/mgm.md
================================================================================

# MGM outage

Check xrdlog.mgm and the number of xrootd processes.

================================================================================
FILE: tape/cta.md
================================================================================

# CTA tape

CTA uses EOS as the disk cache in front of tape.
`
	secs := splitOpsDocs(raw)
	if len(secs) != 2 || secs[0].Path != "repairs/mgm.md" {
		t.Fatalf("split: %#v", secs)
	}
	idx := &opsIndex{sections: secs}
	hits := idx.Search("xrdlog mgm process", 2)
	if len(hits) == 0 || hits[0].Path != "repairs/mgm.md" {
		t.Fatalf("expected mgm section first, got %#v", hits)
	}
	cta := idx.FormatHits("CTA tape cache", 1)
	if !contains(cta, "EOS operators doc: tape/cta.md") || !contains(cta, "disk cache") {
		t.Fatalf("got %q", cta)
	}
}

func TestLoadOpsDocsMissing(t *testing.T) {
	idx := loadOpsDocs(filepath.Join(t.TempDir(), "missing.md"))
	if idx == nil || len(idx.sections) != 0 {
		t.Fatalf("expected empty index, got %#v", idx)
	}
}

func TestLoadOpsDocsFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "all_docs.md")
	if err := os.WriteFile(p, []byte("================================================================================\nFILE: a.md\n================================================================================\n\nQuarkDB RAFT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := loadOpsDocs(p)
	if len(idx.Search("quarkdb", 1)) != 1 {
		t.Fatal("expected one hit")
	}
}
