package ingest

import "testing"

func TestSplitCommitMessage(t *testing.T) {
	title, body := SplitCommitMessage("Fix FUSE recycle\n\nKeep the g visible after spray.")
	if title != "Fix FUSE recycle" {
		t.Fatalf("title: %q", title)
	}
	if body != "Keep the g visible after spray." {
		t.Fatalf("body: %q", body)
	}
	title, body = SplitCommitMessage("  only heading  ")
	if title != "only heading" || body != "" {
		t.Fatalf("heading only: %q %q", title, body)
	}
}
