package chat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/docsview"
	"github.com/apeters/eospage/internal/googleai"
)

const (
	docUserAgent = "eos-page-ask/1.0 (+https://eos.web.cern.ch)"
	maxFetchBody = 2 << 20
)

var allowedDocHosts = map[string]bool{
	"eos-docs.web.cern.ch": true,
	"xrootd.org":           true,
	"www.xrootd.org":       true,
	"cta.web.cern.ch":      true,
}

var fetchClient = &http.Client{
	Timeout: 18 * time.Second,
	Transport: &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	},
}

func parseAllowedDocURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("invalid documentation url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid documentation url")
	}
	host := strings.ToLower(u.Hostname())
	if !allowedDocHosts[host] {
		return nil, fmt.Errorf("only eos-docs.web.cern.ch, xrootd.org, and cta.web.cern.ch can be fetched")
	}
	u.Fragment = ""
	return u, nil
}

func fetchDoc(ctx context.Context, raw string) (string, error) {
	u, err := parseAllowedDocURL(raw)
	if err != nil {
		return "", err
	}
	if strings.EqualFold(u.Hostname(), docsview.AllowedHost) {
		page, err := docsview.Fetch(ctx, u.String())
		if err != nil {
			return "", err
		}
		return formatDocument(page.Title, page.URL, htmlToMarkdown(page.HTML)), nil
	}
	title, body, err := fetchHTML(ctx, u.String())
	if err != nil {
		return "", err
	}
	return formatDocument(title, u.String(), htmlToMarkdown(body)), nil
}

func fetchHTML(ctx context.Context, raw string) (title, html string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", docUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res, err := fetchClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, maxFetchBody+1))
	if err != nil {
		return "", "", err
	}
	if len(body) > maxFetchBody {
		return "", "", fmt.Errorf("page is too large")
	}
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("docs HTTP %d", res.StatusCode)
	}
	if final := res.Request.URL; final != nil {
		if _, err := parseAllowedDocURL(final.String()); err != nil {
			return "", "", err
		}
	}
	rawHTML := string(body)
	title = extractHTMLTitle(rawHTML)
	return title, rawHTML, nil
}

func extractHTMLTitle(raw string) string {
	low := strings.ToLower(raw)
	start := strings.Index(low, "<title>")
	end := strings.Index(low, "</title>")
	if start < 0 || end < start {
		return ""
	}
	return strings.Join(strings.Fields(raw[start+7:end]), " ")
}

func searchAllowedSites(ctx context.Context, query, site string) (string, []Source, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", nil, fmt.Errorf("empty query")
	}
	switch strings.ToLower(strings.TrimSpace(site)) {
	case "", "all":
		q = q + " (site:eos-docs.web.cern.ch OR site:xrootd.org OR site:cta.web.cern.ch)"
	case "eos", "eos-docs", "docs":
		q = q + " site:eos-docs.web.cern.ch"
	case "xrootd":
		q = q + " site:xrootd.org"
	case "cta":
		q = q + " site:cta.web.cern.ch"
	default:
		return "", nil, fmt.Errorf("site must be eos-docs, xrootd, cta, or all")
	}
	got, err := googleai.HTTPSearch(ctx, q)
	if err != nil {
		return "", nil, err
	}
	var sources []Source
	var b strings.Builder
	for _, s := range got.Sources {
		if _, err := parseAllowedDocURL(s.URL); err != nil {
			continue
		}
		sources = append(sources, Source{Title: s.Title, URL: s.URL})
		b.WriteString("- ")
		if s.Title != "" {
			b.WriteString(s.Title)
			b.WriteString(" - ")
		}
		b.WriteString(s.URL)
		b.WriteByte('\n')
		if len(sources) >= 8 {
			break
		}
	}
	if len(sources) == 0 {
		return "", nil, fmt.Errorf("no results on the allowed documentation sites")
	}
	return strings.TrimSpace(b.String()), sources, nil
}
