package googleai

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestGoogleAISearchEmpty(t *testing.T) {
	if _, err := GoogleAISearch(context.Background(), "  "); err == nil {
		t.Fatal("expected empty query error")
	}
}

func TestSearchURL(t *testing.T) {
	u := searchURL("What happened in AI today?", true)
	if !strings.Contains(u, "udm=50") || !strings.Contains(u, "q=What") {
		t.Fatalf("got %s", u)
	}
}

func TestSourcesFromHTML(t *testing.T) {
	html := `<!--Sv6Kpe[["uuid-1",["CERN EOS","docs"],["https://eos-docs.web.cern.ch/page","https://www.google.com/url"]]]-->`
	got := sourcesFromHTML(html)
	if len(got) != 1 || got[0].URL != "https://eos-docs.web.cern.ch/page" {
		t.Fatalf("got %#v", got)
	}
}

func TestSkipAndMerge(t *testing.T) {
	if !skipURL("https://www.google.com/search") {
		t.Fatal("google host")
	}
	merged := mergeSources(
		[]Source{{Title: "EOS", URL: "https://eos-docs.web.cern.ch/"}},
		[]Source{{Title: "", URL: "https://eos-docs.web.cern.ch/"}, {Title: "X", URL: "https://xrootd.org/"}},
	)
	if len(merged) != 2 || merged[0].Title != "EOS" || merged[1].URL != "https://xrootd.org/" {
		t.Fatalf("got %#v", merged)
	}
}

func TestCleanAnswer(t *testing.T) {
	prefix := strings.Repeat("EOS stores LHC data. ", 12)
	in := prefix + "People also ask What is CTA?"
	if got := cleanAnswer(in); got != strings.TrimSpace(prefix) {
		t.Fatalf("got %q", got)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, AIResult{
		Query:  "What happened in AI today?",
		Answer: "A summary.",
		Sources: []Source{{Title: "News", URL: "https://example.test/"}},
	}); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `"query"`) || !strings.Contains(s, `"answer"`) || !strings.Contains(s, "example.test") {
		t.Fatalf("got %s", s)
	}
}

func TestProfileDirDefault(t *testing.T) {
	t.Setenv("CHAT_CHROME_PROFILE", "")
	t.Setenv("DATA_DIR", "data")
	if got := ProfileDir(); got != "data/chrome-profile" && got != "data\\chrome-profile" {
		t.Fatalf("got %q", got)
	}
	t.Setenv("CHAT_CHROME_PROFILE", "/tmp/eos-chrome")
	if ProfileDir() != "/tmp/eos-chrome" {
		t.Fatal(ProfileDir())
	}
}
