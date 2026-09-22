package googleai

import (
	"strings"
	"testing"
)

func TestParseGoogleHTMLResults(t *testing.T) {
	html := `
	<html><body>
	<a href="/url?q=https://eos-docs.web.cern.ch/diopside/&amp;sa=U"><h3>EOS Open Storage</h3></a>
	<a href="/url?q=https://www.google.com/search&amp;sa=U"><h3>Ignore</h3></a>
	<a href="https://xrootd.slac.stanford.edu/"><h3>XRootD</h3></a>
	</body></html>`
	got, err := parseGoogleHTML(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 2 {
		t.Fatalf("sources=%#v", got.Sources)
	}
	if got.Sources[0].URL != "https://eos-docs.web.cern.ch/diopside/" || got.Sources[0].Title != "EOS Open Storage" {
		t.Fatalf("first=%#v", got.Sources[0])
	}
	if got.Sources[1].URL != "https://xrootd.slac.stanford.edu/" {
		t.Fatalf("second=%#v", got.Sources[1])
	}
	if got.IsAI || got.Answer == "" {
		t.Fatalf("answer=%q ai=%v", got.Answer, got.IsAI)
	}
}

func TestParseGoogleHTMLBlocked(t *testing.T) {
	if _, err := parseGoogleHTML("<html>unusual traffic from your computer</html>"); err == nil {
		t.Fatal("expected block")
	}
	if _, err := parseGoogleHTML("<html><h3>no links</h3></html>"); err == nil {
		t.Fatal("expected empty")
	}
}

func TestParseDDGHTML(t *testing.T) {
	html := `
	<a rel="nofollow" class="result__a" href="https://eos-docs.web.cern.ch/diopside/">EOS docs</a>
	<a class="result__snippet">Open storage at CERN</a>
	<a rel="nofollow" class="result__a" href="https://github.com/cern-eos/eos">cern-eos/eos</a>
	<a class="result__snippet">The EOS storage software</a>`
	got, err := parseDDGHTML(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 2 || got.Sources[0].URL != "https://eos-docs.web.cern.ch/diopside/" {
		t.Fatalf("got %#v", got.Sources)
	}
	if !strings.Contains(got.Sources[0].Title, "EOS docs") {
		t.Fatalf("title %q", got.Sources[0].Title)
	}
}

func TestParseDDGLite(t *testing.T) {
	html := `<a href="https://quarkdb.web.cern.ch/" class="result-link">QuarkDB</a>`
	got, err := parseDDGLite(html)
	if err != nil || len(got.Sources) != 1 || got.Sources[0].URL != "https://quarkdb.web.cern.ch/" {
		t.Fatalf("got %#v err=%v", got, err)
	}
}

func TestDecodeBingURL(t *testing.T) {
	href := `https://www.bing.com/ck/a?!&&p=x&u=a1aHR0cHM6Ly9lb3MtZG9jcy53ZWIuY2Vybi5jaC9kaW9wc2lkZS8&ntb=1`
	if got := decodeBingURL(href); got != "https://eos-docs.web.cern.ch/diopside/" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBingHTML(t *testing.T) {
	html := `<h2><a href="https://www.bing.com/ck/a?u=a1aHR0cHM6Ly9lb3MtZG9jcy53ZWIuY2Vybi5jaC9kaW9wc2lkZS8">EOS docs</a></h2>`
	got, err := parseBingHTML(html)
	if err != nil || len(got.Sources) != 1 || got.Sources[0].URL != "https://eos-docs.web.cern.ch/diopside/" {
		t.Fatalf("got %#v err=%v", got, err)
	}
}

func TestFilterSources(t *testing.T) {
	in := []Source{
		{Title: "EOS docs", URL: "https://eos-docs.web.cern.ch/"},
		{Title: "Phone LCD", URL: "https://gsmserver.com/realme-c55"},
		{Title: "Bing", URL: "https://www.microsoft.com/en-us/bing/"},
	}
	got := filterSources("What is EOS storage at CERN?", in)
	if len(got) != 1 || got[0].URL != "https://eos-docs.web.cern.ch/" {
		t.Fatalf("got %#v", got)
	}
}

func TestHTTPSearchEmpty(t *testing.T) {
	if _, err := HTTPSearch(nil, "  "); err == nil {
		t.Fatal("expected empty query")
	}
}
