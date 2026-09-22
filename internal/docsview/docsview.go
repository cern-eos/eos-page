package docsview

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	AllowedHost = "eos-docs.web.cern.ch"
	userAgent   = "eos-page/1.0 (+https://eos.web.cern.ch)"
	maxBody     = 2 << 20
	cacheTTL    = 10 * time.Minute
	cacheMax    = 64
)

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Page struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	HTML  string `json:"html"`
	Prev  *Link  `json:"prev,omitempty"`
	Next  *Link  `json:"next,omitempty"`
}

var client = &http.Client{
	Timeout: 18 * time.Second,
	Transport: &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	},
}

type cacheEntry struct {
	page Page
	at   time.Time
}

var (
	cacheMu sync.Mutex
	cache   = map[string]cacheEntry{}
)

func Allowed(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("invalid docs url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid docs url")
	}
	if strings.ToLower(u.Hostname()) != AllowedHost {
		return nil, fmt.Errorf("only %s can be viewed here", AllowedHost)
	}
	u.Fragment = ""
	return u, nil
}

func Fetch(ctx context.Context, raw string) (Page, error) {
	u, err := Allowed(raw)
	if err != nil {
		return Page{}, err
	}
	key := u.String()
	cacheMu.Lock()
	if e, ok := cache[key]; ok && time.Since(e.at) < cacheTTL {
		cacheMu.Unlock()
		return e.page, nil
	}
	cacheMu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, key, nil)
	if err != nil {
		return Page{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res, err := client.Do(req)
	if err != nil {
		return Page{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBody+1))
	if err != nil {
		return Page{}, err
	}
	if len(body) > maxBody {
		return Page{}, fmt.Errorf("docs page is too large")
	}
	if res.StatusCode != http.StatusOK {
		return Page{}, fmt.Errorf("docs HTTP %d", res.StatusCode)
	}
	final := res.Request.URL
	if _, err := Allowed(final.String()); err != nil {
		return Page{}, err
	}
	page, err := Parse(body, final)
	if err != nil {
		return Page{}, err
	}
	cacheMu.Lock()
	if len(cache) >= cacheMax {
		cache = map[string]cacheEntry{}
	}
	cache[key] = cacheEntry{page: page, at: time.Now()}
	cacheMu.Unlock()
	return page, nil
}

func Parse(raw []byte, base *url.URL) (Page, error) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return Page{}, err
	}
	title := textContent(find(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "title"
	}))
	title = strings.TrimSpace(strings.Split(title, "—")[0])
	title = strings.TrimSpace(strings.Split(title, "&#8212;")[0])

	main := find(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && (hasClass(n, "body") || attr(n, "role") == "main")
	})
	if main == nil {
		return Page{}, fmt.Errorf("could not find the article on that docs page")
	}
	walk(main, base)
	if h1 := find(main, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1"
	}); h1 != nil {
		if t := strings.TrimSpace(textContent(h1)); t != "" {
			title = t
		}
	}
	plainDash := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, " — ", " - "), "—", "-")
	}
	title = plainDash(title)
	var buf bytes.Buffer
	for c := main.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			return Page{}, err
		}
	}
	page := Page{Title: title, URL: base.String(), HTML: plainDash(buf.String())}
	related := find(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && hasClass(n, "related")
	})
	if related != nil {
		page.Prev, page.Next = relatedLinks(related, base)
	}
	return page, nil
}

func relatedLinks(root *html.Node, base *url.URL) (prev, next *Link) {
	var walkLinks func(*html.Node)
	walkLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := attr(n, "href")
			label := strings.ToLower(strings.TrimSpace(textContent(n)))
			key := strings.ToLower(attr(n, "accesskey"))
			abs := resolve(base, href)
			if abs != "" && AllowedHostOK(abs) {
				title := strings.TrimSpace(attr(n, "title"))
				if title == "" {
					title = strings.TrimSpace(textContent(n))
				}
				link := &Link{Title: title, URL: abs}
				if key == "p" || label == "previous" || strings.HasPrefix(label, "previous") {
					prev = link
				}
				if key == "n" || label == "next" || strings.HasPrefix(label, "next") {
					next = link
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkLinks(c)
		}
	}
	walkLinks(root)
	return prev, next
}

func AllowedHostOK(raw string) bool {
	_, err := Allowed(raw)
	return err == nil
}

var dropTag = map[string]bool{
	"script": true, "style": true, "iframe": true, "object": true, "embed": true,
	"form": true, "input": true, "button": true, "textarea": true, "select": true,
	"noscript": true, "link": true, "meta": true, "svg": true,
}

func walk(n *html.Node, base *url.URL) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		if c.Type == html.ElementNode && dropTag[c.Data] {
			n.RemoveChild(c)
			c = next
			continue
		}
		if c.Type == html.ElementNode {
			rewrite(c, base)
			if hasClass(c, "headerlink") {
				n.RemoveChild(c)
				c = next
				continue
			}
		}
		walk(c, base)
		c = next
	}
}

func rewrite(n *html.Node, base *url.URL) {
	kept := n.Attr[:0]
	for _, a := range n.Attr {
		key := strings.ToLower(a.Key)
		if strings.HasPrefix(key, "on") || key == "srcdoc" || key == "style" {
			continue
		}
		if key == "href" || key == "src" {
			abs := resolve(base, a.Val)
			if abs == "" {
				continue
			}
			a.Val = abs
			if key == "href" && !AllowedHostOK(abs) && !strings.HasPrefix(abs, "mailto:") {
				kept = append(kept, html.Attribute{Key: "target", Val: "_blank"})
				kept = append(kept, html.Attribute{Key: "rel", Val: "noreferrer"})
			}
		}
		if key == "srcset" {
			continue
		}
		kept = append(kept, a)
	}
	n.Attr = kept
}

func resolve(base *url.URL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "javascript:") || strings.HasPrefix(raw, "data:") {
		return ""
	}
	if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "mailto:") {
		return raw
	}
	u, err := base.Parse(raw)
	if err != nil {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "mailto" {
		return ""
	}
	return u.String()
}

func find(n *html.Node, ok func(*html.Node) bool) *html.Node {
	if n == nil {
		return nil
	}
	if ok(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := find(c, ok); found != nil {
			return found
		}
	}
	return nil
}

func hasClass(n *html.Node, name string) bool {
	for _, part := range strings.Fields(attr(n, "class")) {
		if part == name {
			return true
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textContent(c))
	}
	return b.String()
}
