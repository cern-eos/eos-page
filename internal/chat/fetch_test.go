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

func TestSystemPromptCERNDiskStorage(t *testing.T) {
	p := systemPromptOpenAI("", "")
	for _, want := range []string{
		"CERN disk storage",
		"EOS Open Storage",
		"Never write CERN Open Storage, Essential Open Storage",
		"what EOS is and what it is used for",
		"do not refuse ordinary EOS questions",
		"eos-docs.web.cern.ch",
		"xrootd.org",
		"cta.web.cern.ch",
		"structured Markdown",
		"prefer the published figures",
		"web_search",
	} {
		if !contains(p, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if contains(p, "Out of scope") || contains(p, "one-sentence refusal") {
		t.Fatal("prompt still uses hard refuse language")
	}
}

func TestFrameUserQuestion(t *testing.T) {
	got := frameUserQuestion("What is EOS used for ?")
	if !contains(got, "CERN EOS disk storage") || !contains(got, "What is EOS used for ?") {
		t.Fatalf("got %q", got)
	}
}
