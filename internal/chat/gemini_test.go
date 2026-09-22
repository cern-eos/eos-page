package chat

import (
	"context"
	"strings"
	"testing"
)

func TestClipQuestion(t *testing.T) {
	if _, err := ClipQuestion("  "); err == nil {
		t.Fatal("expected empty error")
	}
	q, err := ClipQuestion("  What is QuarkDB?  ")
	if err != nil || q != "What is QuarkDB?" {
		t.Fatalf("got %q %v", q, err)
	}
}

func TestClipHistory(t *testing.T) {
	in := []Message{
		{Role: "system", Text: "nope"},
		{Role: "user", Text: "one"},
		{Role: "assistant", Text: "two"},
		{Role: "", Text: "skip"},
	}
	out := ClipHistory(in)
	if len(out) != 2 || out[0].Role != "user" || out[1].Role != "assistant" {
		t.Fatalf("got %#v", out)
	}
}

func TestNewWithoutKey(t *testing.T) {
	c, err := New(context.Background(), "")
	if err != nil || c == nil {
		t.Fatalf("expected search-only client, got %v %v", c, err)
	}
	if c.api != nil {
		t.Fatal("expected no Gemini client")
	}
}

func TestFallbackReply(t *testing.T) {
	r := fallbackReply("QuarkDB", "- Doc: QuarkDB — persistency\n", "QuarkDB is the EOS metadata store.", []Source{{Title: "QuarkDB", URL: "https://example.test"}}, nil)
	if !strings.Contains(r.Text, "QuarkDB") || len(r.Sources) != 1 {
		t.Fatalf("got %#v", r)
	}
}
