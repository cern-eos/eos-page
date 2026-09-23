package chat

import "testing"

func TestParseAllowedDocURL(t *testing.T) {
	ok := []string{
		"https://eos-docs.web.cern.ch/diopside/architecture/index.html",
		"https://xrootd.org/docs.html",
		"https://www.xrootd.org/docs.html",
		"https://cta.web.cern.ch/cta/pages/documentation.html",
	}
	for _, raw := range ok {
		if _, err := parseAllowedDocURL(raw); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
	}
	if _, err := parseAllowedDocURL("https://example.com/docs"); err == nil {
		t.Fatal("expected reject")
	}
	if _, err := parseAllowedDocURL("file:///etc/passwd"); err == nil {
		t.Fatal("expected reject file")
	}
}

func TestSystemPromptRejectsOffTopic(t *testing.T) {
	p := systemPromptOpenAI("", "")
	for _, want := range []string{
		"always answer these",
		"What is EOS used for",
		"CERN disk storage",
		"CTA",
		"XRootD",
		"eos-docs.web.cern.ch",
		"xrootd.org",
		"cta.web.cern.ch",
		"structured Markdown",
	} {
		if !contains(p, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
