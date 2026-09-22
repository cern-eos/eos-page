package googleai

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var httpClient = &http.Client{
	Timeout: 12 * time.Second,
	Transport: &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	},
}

var (
	resultPairRE = regexp.MustCompile(`(?is)<a[^>]+href="(/url\?q=|https?://)([^"]+)"[^>]*>\s*<h3[^>]*>(.*?)</h3>`)
	urlQueryRE   = regexp.MustCompile(`(?i)/url\?q=(https?://[^&"]+)`)
	ddgResultRE  = regexp.MustCompile(`(?is)<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	ddgSnippetRE = regexp.MustCompile(`(?is)class="result__snippet"[^>]*>(.*?)</(?:a|td|div)`)
	ddgLiteRE    = regexp.MustCompile(`(?is)<a[^>]*(?:class="[^"]*result-link[^"]*"[^>]*href="([^"]+)"|href="([^"]+)"[^>]*class="[^"]*result-link[^"]*")[^>]*>(.*?)</a>`)
	ddgAnyHREFRE = regexp.MustCompile(`(?is)<a[^>]+href="(https?://[^"]+)"[^>]*>(.*?)</a>`)
	bingH2RE     = regexp.MustCompile(`(?is)<h2[^>]*>\s*<a[^>]+href="([^"]+)"[^>]*>(.*?)</a>`)
	tagRE        = regexp.MustCompile(`(?s)<[^>]+>`)
)

// HTTPSearch fetches live web results without Chrome. Google's HTML search
// now requires JavaScript (or a captcha), so this uses DuckDuckGo's HTML API.
func HTTPSearch(parent context.Context, query string) (AIResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return AIResult{}, fmt.Errorf("empty query")
	}
	ctx := parent
	if ctx == nil {
		ctx = context.Background()
	}
	for _, fn := range []func(context.Context, string) (AIResult, error){
		fetchBing,
		func(ctx context.Context, q string) (AIResult, error) { return fetchDDG(ctx, q, true) },
		func(ctx context.Context, q string) (AIResult, error) { return fetchDDG(ctx, q, false) },
	} {
		got, err := fn(ctx, query)
		if err != nil {
			continue
		}
		got.Sources = filterSources(query, got.Sources)
		if len(got.Sources) == 0 {
			continue
		}
		got.Answer = formatSources(got.Sources)
		got.Query = query
		return got, nil
	}
	return AIResult{Query: query}, fmt.Errorf("empty search results")
}

func formatSources(sources []Source) string {
	var b strings.Builder
	for i, s := range sources {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Title)
		b.WriteByte('\n')
		b.WriteString(s.URL)
	}
	return b.String()
}

func queryKeys(q string) []string {
	stop := map[string]bool{"what": true, "is": true, "the": true, "for": true, "and": true, "how": true, "does": true, "a": true, "an": true, "to": true, "of": true, "in": true, "on": true, "at": true}
	var keys []string
	for _, w := range strings.Fields(strings.ToLower(q)) {
		w = strings.Trim(w, "?,.!\"'")
		if len(w) < 3 || stop[w] {
			continue
		}
		keys = append(keys, w)
	}
	low := strings.ToLower(q)
	if strings.Contains(low, "eosxd") || strings.Contains(low, "eos-") {
		keys = append(keys, "eos", "fuse")
	}
	return keys
}

func filterSources(query string, in []Source) []Source {
	keys := queryKeys(query)
	var out []Source
	for _, s := range in {
		blob := strings.ToLower(s.Title + " " + s.URL)
		if strings.Contains(blob, "microsoft.com") && strings.Contains(blob, "bing") {
			continue
		}
		if len(keys) == 0 {
			out = append(out, s)
			continue
		}
		hit := 0
		for _, k := range keys {
			if strings.Contains(blob, k) {
				hit++
			}
		}
		if hit == 0 {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		for _, s := range in {
			blob := strings.ToLower(s.Title + " " + s.URL)
			if strings.Contains(blob, "cern") || strings.Contains(blob, "eos") || strings.Contains(blob, "xrootd") {
				out = append(out, s)
			}
		}
	}
	return out
}

func fetchDDG(ctx context.Context, query string, htmlAPI bool) (AIResult, error) {
	var req *http.Request
	var err error
	if htmlAPI {
		form := url.Values{"q": {query}, "kl": {"us-en"}}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, "https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
		if err != nil {
			return AIResult{}, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", "https://html.duckduckgo.com/html/")
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://lite.duckduckgo.com/lite/?q="+url.QueryEscape(query), nil)
		if err != nil {
			return AIResult{}, err
		}
	}
	req.Header.Set("User-Agent", chromeUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	res, err := httpClient.Do(req)
	if err != nil {
		return AIResult{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return AIResult{}, err
	}
	if res.StatusCode != http.StatusOK {
		return AIResult{}, fmt.Errorf("web search http %s", res.Status)
	}
	raw := string(body)
	if htmlAPI {
		return parseDDGHTML(raw)
	}
	return parseDDGLite(raw)
}

func fetchBing(ctx context.Context, query string) (AIResult, error) {
	u := "https://www.bing.com/search?setlang=en-US&cc=us&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return AIResult{}, err
	}
	req.Header.Set("User-Agent", chromeUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	res, err := httpClient.Do(req)
	if err != nil {
		return AIResult{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return AIResult{}, err
	}
	if res.StatusCode != http.StatusOK {
		return AIResult{}, fmt.Errorf("web search http %s", res.Status)
	}
	return parseBingHTML(string(body))
}

func decodeBingURL(href string) string {
	href = html.UnescapeString(href)
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	raw := u.Query().Get("u")
	raw = strings.TrimPrefix(raw, "a1")
	if raw != "" {
		if dec, err := base64.RawURLEncoding.DecodeString(raw); err == nil {
			if out := sanitizeURL(string(dec)); out != "" {
				return out
			}
		}
		if pad := raw + strings.Repeat("=", (4-len(raw)%4)%4); pad != raw {
			if dec, err := base64.URLEncoding.DecodeString(pad); err == nil {
				if out := sanitizeURL(string(dec)); out != "" {
					return out
				}
			}
		}
	}
	if strings.Contains(strings.ToLower(u.Hostname()), "bing.") {
		return ""
	}
	return sanitizeURL(href)
}

func parseBingHTML(raw string) (AIResult, error) {
	seen := map[string]bool{}
	var sources []Source
	for _, m := range bingH2RE.FindAllStringSubmatch(raw, 16) {
		href := decodeBingURL(m[1])
		if href == "" || seen[href] || skipURL(href) {
			continue
		}
		seen[href] = true
		title := strings.Join(strings.Fields(stripTags(m[2])), " ")
		if title == "" {
			title = hostTitle(href)
		}
		sources = append(sources, Source{Title: title, URL: href})
		if len(sources) >= 8 {
			break
		}
	}
	if len(sources) == 0 {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	var b strings.Builder
	for i, s := range sources {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Title)
		b.WriteByte('\n')
		b.WriteString(s.URL)
	}
	return AIResult{Answer: b.String(), Sources: sources}, nil
}

func parseDDGLite(raw string) (AIResult, error) {
	seen := map[string]bool{}
	var sources []Source
	add := func(title, href string) {
		href = sanitizeURL(html.UnescapeString(href))
		if href == "" || seen[href] || skipURL(href) || strings.Contains(href, "duckduckgo.com") {
			return
		}
		seen[href] = true
		title = strings.Join(strings.Fields(stripTags(title)), " ")
		if title == "" {
			title = hostTitle(href)
		}
		sources = append(sources, Source{Title: title, URL: href})
	}
	for _, m := range ddgLiteRE.FindAllStringSubmatch(raw, 16) {
		href := m[1]
		if href == "" {
			href = m[2]
		}
		add(m[3], href)
		if len(sources) >= 8 {
			break
		}
	}
	if len(sources) == 0 && strings.Contains(raw, "result-link") {
		for _, m := range ddgAnyHREFRE.FindAllStringSubmatch(raw, 24) {
			add(m[2], m[1])
			if len(sources) >= 8 {
				break
			}
		}
	}
	if len(sources) == 0 {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	var b strings.Builder
	for i, s := range sources {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Title)
		b.WriteByte('\n')
		b.WriteString(s.URL)
	}
	return AIResult{Answer: b.String(), Sources: sources}, nil
}

func parseDDGHTML(raw string) (AIResult, error) {
	if strings.Contains(raw, `href="https://duckduckgo.com/"`) && !strings.Contains(raw, "result__a") {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	seen := map[string]bool{}
	var sources []Source
	snips := ddgSnippetRE.FindAllStringSubmatch(raw, 16)
	for i, m := range ddgResultRE.FindAllStringSubmatch(raw, 16) {
		href := sanitizeURL(html.UnescapeString(m[1]))
		if href == "" || seen[href] || skipURL(href) {
			continue
		}
		seen[href] = true
		title := strings.Join(strings.Fields(stripTags(m[2])), " ")
		if title == "" {
			title = hostTitle(href)
		}
		if i < len(snips) {
			if snippet := strings.Join(strings.Fields(stripTags(snips[i][1])), " "); snippet != "" {
				title = title + " - " + clipRunes(snippet, 180)
			}
		}
		sources = append(sources, Source{Title: title, URL: href})
		if len(sources) >= 8 {
			break
		}
	}
	if len(sources) == 0 {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	var b strings.Builder
	for i, s := range sources {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Title)
		b.WriteString("\n")
		b.WriteString(s.URL)
	}
	return AIResult{Answer: b.String(), Sources: sources, IsAI: false}, nil
}

func parseGoogleHTML(raw string) (AIResult, error) {
	low := strings.ToLower(raw)
	if strings.Contains(low, "unusual traffic") || strings.Contains(low, "are you a robot") || strings.Contains(low, "detected unusual") {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}
	if strings.Contains(low, "before you continue to google") && utf8.RuneCountInString(raw) < 8000 {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}

	seen := map[string]bool{}
	var sources []Source
	add := func(title, href string) {
		href = sanitizeURL(html.UnescapeString(href))
		if href == "" || seen[href] || skipURL(href) {
			return
		}
		seen[href] = true
		title = strings.Join(strings.Fields(stripTags(title)), " ")
		if title == "" {
			title = hostTitle(href)
		}
		sources = append(sources, Source{Title: title, URL: href})
	}

	for _, m := range resultPairRE.FindAllStringSubmatch(raw, 20) {
		href := html.UnescapeString(m[2])
		if m[1] == "/url?q=" || strings.HasPrefix(m[1], "/url") {
			if q := urlQueryRE.FindStringSubmatch("/url?q=" + href); len(q) > 1 {
				href = q[1]
			} else if decoded, err := url.QueryUnescape(href); err == nil {
				href = decoded
			}
		} else {
			href = m[1] + href
		}
		add(m[3], href)
		if len(sources) >= 8 {
			break
		}
	}

	if len(sources) < 2 {
		for _, m := range urlQueryRE.FindAllStringSubmatch(raw, 24) {
			add("", m[1])
			if len(sources) >= 8 {
				break
			}
		}
	}

	if len(sources) == 0 {
		return AIResult{}, fmt.Errorf("empty search results")
	}

	var b strings.Builder
	for i, s := range sources {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Title)
		b.WriteString(" - ")
		b.WriteString(s.URL)
	}
	return AIResult{Answer: b.String(), Sources: sources, IsAI: false}, nil
}

func stripTags(s string) string {
	return html.UnescapeString(tagRE.ReplaceAllString(s, " "))
}

func clipRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}
