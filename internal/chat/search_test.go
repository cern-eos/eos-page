package chat

import (
	"net/url"
	"strings"
	"testing"
)

func TestSearchQuery(t *testing.T) {
	tag := "CERN's EOS Storage System"
	if q := searchQuery("What is QuarkDB?"); !strings.Contains(q, tag) || !strings.HasPrefix(q, "What is QuarkDB?") {
		t.Fatalf("expected CERN EOS tag, got %q", q)
	}
	if q := searchQuery("EOS FUSE mount"); !strings.Contains(q, tag) {
		t.Fatalf("expected tag on EOS questions, got %q", q)
	}
	already := "How does CTA work? This question is related to CERN's EOS Storage System."
	if q := searchQuery(already); q != already {
		t.Fatalf("should not double-tag, got %q", q)
	}
	if _, err := url.Parse("https://www.google.com/search?q=" + url.QueryEscape(searchQuery("eosxd"))); err != nil {
		t.Fatal(err)
	}
}

func TestWebQuery(t *testing.T) {
	if q := webQuery("What is QuarkDB?"); !strings.Contains(strings.ToLower(q), "quarkdb") || strings.Contains(q, "This question") {
		t.Fatalf("got %q", q)
	}
	if q := webQuery("How do I mount eosxd?"); q != "eosxd FUSE mount CERN EOS" {
		t.Fatalf("expected eosxd search, got %q", q)
	}
	if q := webQuery("how do I mount fuse?"); !strings.Contains(q, "CERN EOS") {
		t.Fatalf("expected EOS hint, got %q", q)
	}
}

func TestClipSearchText(t *testing.T) {
	if clipSearchText("  hello   world  ") != "hello world" {
		t.Fatal("normalize")
	}
}
