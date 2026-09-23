package chat

import "testing"

func TestHTMLToMarkdown(t *testing.T) {
	md := htmlToMarkdown(`<html><body><div class="body">
		<h1>QuarkDB</h1>
		<p>High-available <strong>RAFT</strong> store.</p>
		<ul><li>Node A</li><li>Node B</li></ul>
		<pre><code>eos ns</code></pre>
		<p>See <a href="https://eos-docs.web.cern.ch/diopside/">the docs</a>.</p>
	</div></body></html>`)
	for _, want := range []string{
		"# QuarkDB",
		"**RAFT**",
		"- Node A",
		"```",
		"eos ns",
		"[the docs](https://eos-docs.web.cern.ch/diopside/)",
	} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestFormatDocument(t *testing.T) {
	got := formatDocument("MGM", "https://eos-docs.web.cern.ch/x/", "# Hello")
	if !contains(got, "### Document: MGM") || !contains(got, "Source: https://eos-docs.web.cern.ch/x/") {
		t.Fatalf("got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && stringIndex(s, sub) >= 0))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
