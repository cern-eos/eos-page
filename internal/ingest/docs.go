package ingest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/store"
)

var DefaultDocSources = []docSource{
	{"intro", "Introduction", "https://eos-docs.web.cern.ch/diopside/_sources/introduction/index.rst.txt", "https://eos-docs.web.cern.ch/diopside/introduction/index.html"},
	{"architecture", "Design & Architecture", "https://eos-docs.web.cern.ch/diopside/_sources/architecture/index.rst.txt", "https://eos-docs.web.cern.ch/diopside/architecture/index.html"},
	{"releases", "Releases", "https://eos-docs.web.cern.ch/diopside/_sources/releases/index.rst.txt", "https://eos-docs.web.cern.ch/diopside/releases/index.html"},
	{"diopside-release", "Diopside Release Notes", "https://eos-docs.web.cern.ch/diopside/_sources/releases/diopside-release.rst.txt", "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html"},
	{"getting-started", "Getting Started", "https://eos-docs.web.cern.ch/diopside/_sources/manual/getting-started.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/getting-started.html"},
	{"hardware", "Hardware Requirements", "https://eos-docs.web.cern.ch/diopside/_sources/manual/hardware-installation.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/hardware-installation.html"},
	{"configuration", "Configuration", "https://eos-docs.web.cern.ch/diopside/_sources/manual/configuration.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/configuration.html"},
	{"interfaces", "Interfaces", "https://eos-docs.web.cern.ch/diopside/_sources/manual/interfaces.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/interfaces.html"},
	{"protocols", "Protocols & APIs", "https://eos-docs.web.cern.ch/diopside/_sources/manual/protocols.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/protocols.html"},
	{"using", "Using EOS", "https://eos-docs.web.cern.ch/diopside/_sources/manual/using.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/using.html"},
	{"microservices", "MGM Microservices", "https://eos-docs.web.cern.ch/diopside/_sources/manual/microservices.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/microservices.html"},
	{"develop", "Developing EOS", "https://eos-docs.web.cern.ch/diopside/_sources/manual/develop.rst.txt", "https://eos-docs.web.cern.ch/diopside/manual/develop.html"},
	{"faq", "FAQ", "https://eos-docs.web.cern.ch/diopside/_sources/faq/exotic.rst.txt", "https://eos-docs.web.cern.ch/diopside/faq/exotic.html"},
	{"blog", "Blog / Features", "https://eos-docs.web.cern.ch/diopside/_sources/blog/features.rst.txt", "https://eos-docs.web.cern.ch/diopside/blog/features.html"},
}

type docSource struct {
	ID, Title, RST, HTML string
}

func FetchDocs() ([]store.Doc, error) {
	client := &http.Client{Timeout: 45 * time.Second}
	var out []store.Doc
	seen := map[string]int{}
	for _, src := range DefaultDocSources {
		body, err := getText(client, src.RST)
		if err != nil {
			return nil, fmt.Errorf("docs %s: %w", src.ID, err)
		}
		for _, d := range chunkRST(body, src.ID, src.Title, src.HTML) {
			if n := seen[d.ID]; n > 0 {
				d.ID = fmt.Sprintf("%s-%d", d.ID, n)
			}
			seen[d.ID]++
			d.Sort = len(out)
			out = append(out, d)
		}
	}
	return out, nil
}

func getText(client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUA)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d for %s", res.StatusCode, rawURL)
	}
	return string(b), nil
}

var (
	rstRole   = regexp.MustCompile(`:(?:doc|ref|code|math):` + "`" + `([^` + "`" + `]+)` + "`")
	rstLink   = regexp.MustCompile("`([^`<]+)\\s*<[^>]+>`_")
	rstCode   = regexp.MustCompile("``([^`]+)``")
	rstBold   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	rstEm     = regexp.MustCompile(`\*([^*]+)\*`)
	rstID     = regexp.MustCompile(`[^a-z0-9]+`)
	headingOK = regexp.MustCompile(`^[A-Z0-9].{2,90}$`)
)

func chunkRST(text, pageID, pageTitle, htmlURL string) []store.Doc {
	plain := rstPlain(text)
	lines := strings.Split(plain, "\n")
	var out []store.Doc
	cur := pageTitle
	var buf []string
	flush := func() {
		body := strings.TrimSpace(strings.Join(buf, "\n"))
		body = regexp.MustCompile(`\n{3,}`).ReplaceAllString(body, "\n\n")
		if len(body) < 40 {
			return
		}
		id := strings.Trim(rstID.ReplaceAllString(strings.ToLower(pageID+"-"+cur), "-"), "-")
		if len(id) > 80 {
			id = id[:80]
		}
		sum := strings.Join(strings.Fields(body), " ")
		if len(sum) > 280 {
			sum = sum[:280]
		}
		if len(body) > 8000 {
			body = body[:8000]
		}
		out = append(out, store.Doc{
			ID: id, PageID: pageID, Title: clip(cur, 160), Section: pageTitle,
			URL: htmlURL, Summary: sum, Body: body,
		})
	}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line != "" && !strings.HasPrefix(line, " ") && len(line) >= 3 && len(line) <= 90 &&
			headingOK.MatchString(line) && i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" && len(buf) > 4 {
			flush()
			cur = strings.TrimSpace(line)
			buf = nil
			continue
		}
		buf = append(buf, line)
	}
	flush()
	if len(out) == 0 {
		body := strings.TrimSpace(plain)
		out = append(out, store.Doc{
			ID: pageID, PageID: pageID, Title: pageTitle, Section: pageTitle,
			URL: htmlURL, Summary: clip(strings.Join(strings.Fields(body), " "), 280), Body: clip(body, 8000),
		})
	}
	return out
}

func rstPlain(text string) string {
	var out []string
	skip := false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, ".. ") && !strings.HasPrefix(line, ".. _") &&
			!strings.Contains(line, "code-block") && !strings.Contains(line, "note::") && !strings.Contains(line, "warning::") {
			if strings.HasPrefix(line, ".. image::") || strings.HasPrefix(line, ".. toctree::") ||
				strings.HasPrefix(line, ".. index::") || strings.HasPrefix(line, ".. highlight") {
				skip = true
				continue
			}
		}
		if skip {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, ":") || strings.TrimSpace(line) == "" {
				if strings.TrimSpace(line) == "" {
					skip = false
				}
				continue
			}
			skip = false
		}
		line = rstLink.ReplaceAllString(line, "$1")
		line = rstCode.ReplaceAllString(line, "$1")
		line = rstBold.ReplaceAllString(line, "$1")
		line = rstEm.ReplaceAllString(line, "$1")
		line = rstRole.ReplaceAllString(line, "$1")
		trim := strings.TrimSpace(line)
		if len(trim) >= 3 && strings.Trim(trim, "=-~^#*+") == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func DocHTMLURL(base, rstURL string) string {
	u, err := url.Parse(rstURL)
	if err != nil {
		return base
	}
	u.Path = strings.Replace(u.Path, "/_sources/", "/", 1)
	u.Path = strings.TrimSuffix(u.Path, ".rst.txt") + ".html"
	return u.String()
}
