package docsview

import (
	"net/url"
	"strings"
	"testing"
)

const sample = `<!DOCTYPE html><html><head><title>4.3. Getting Started — EOS DIOPSIDE documentation</title></head>
<body>
<div class="related" role="navigation">
  <a href="hardware-installation.html" title="4.1. Hardware Requirements" accesskey="P">previous</a>
  <a href="configuration.html" title="4.4. Configuration" accesskey="N">next</a>
</div>
<div class="body" role="main">
  <div class="section" id="getting-started">
    <h1>Getting Started<a class="headerlink" href="#getting-started">¶</a></h1>
    <p>Install with <a href="../architecture/index.html">architecture</a>. Fixes EOS-5043 and EOS-12.</p>
    <p>Already linked <a href="https://its.cern.ch/jira/browse/EOS-5043">EOS-5043</a>.</p>
    <script>alert(1)</script>
    <p><a href="https://gitlab.cern.ch/eos">GitLab</a></p>
    <img src="../_images/foo.png" onerror="alert(1)"/>
  </div>
</div>
</body></html>`

func TestParseExtractsAndSanitizes(t *testing.T) {
	base, _ := url.Parse("https://eos-docs.web.cern.ch/diopside/manual/getting-started.html")
	page, err := Parse([]byte(sample), base)
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "Getting Started" {
		t.Fatalf("title %q", page.Title)
	}
	if strings.Contains(page.HTML, "<script") || strings.Contains(page.HTML, "onerror") || strings.Contains(page.HTML, "headerlink") {
		t.Fatalf("unsafe html: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, "https://eos-docs.web.cern.ch/diopside/architecture/index.html") {
		t.Fatalf("relative docs link not rewritten: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, `src="https://eos-docs.web.cern.ch/diopside/_images/foo.png"`) {
		t.Fatalf("image not rewritten: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, `target="_blank"`) || !strings.Contains(page.HTML, "https://gitlab.cern.ch/eos") {
		t.Fatalf("external link: %s", page.HTML)
	}
	if page.Prev == nil || !strings.Contains(page.Prev.URL, "hardware-installation.html") {
		t.Fatalf("prev %+v", page.Prev)
	}
	if page.Next == nil || !strings.Contains(page.Next.URL, "configuration.html") {
		t.Fatalf("next %+v", page.Next)
	}
	if !strings.Contains(page.HTML, `href="https://its.cern.ch/jira/browse/EOS-12"`) {
		t.Fatalf("jira tickets not linked: %s", page.HTML)
	}
	if strings.Contains(page.HTML, `<a href="https://its.cern.ch/jira/browse/EOS-5043"><a`) {
		t.Fatalf("already-linked ticket was wrapped again: %s", page.HTML)
	}
}

func TestAllowed(t *testing.T) {
	if _, err := Allowed("https://eos-docs.web.cern.ch/diopside/"); err != nil {
		t.Fatal(err)
	}
	if _, err := Allowed("https://example.com/x"); err == nil {
		t.Fatal("expected reject")
	}
	if _, err := Allowed("javascript:alert(1)"); err == nil {
		t.Fatal("expected reject")
	}
}
