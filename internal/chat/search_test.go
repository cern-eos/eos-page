package chat

import (
	"net/url"
	"strings"
	"testing"
)

func TestSearchQuery(t *testing.T) {
	if q := searchQuery("What is QuarkDB?"); !strings.Contains(q, "EOS CERN storage") {
		t.Fatalf("expected EOS bias, got %q", q)
	}
	if q := searchQuery("EOS FUSE mount"); q != "EOS FUSE mount" {
		t.Fatalf("got %q", q)
	}
	if _, err := url.Parse("https://www.google.com/search?q=" + url.QueryEscape(searchQuery("eosxd"))); err != nil {
		t.Fatal(err)
	}
}

func TestClipSearchText(t *testing.T) {
	if clipSearchText("  hello   world  ") != "hello world" {
		t.Fatal("normalize")
	}
}
