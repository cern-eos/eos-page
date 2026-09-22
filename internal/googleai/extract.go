package googleai

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxAnswer = 8000

type extractResult struct {
	Answer  string              `json:"answer"`
	Sources []map[string]string `json:"sources"`
	IsAI    bool                `json:"isAI"`
	Ready   bool                `json:"ready"`
	Blocked bool                `json:"blocked"`
}

const clickAIModeJS = `(() => {
  const match = (s) => {
    s = (s || '').replace(/\s+/g, ' ').trim();
    if (!s || s.length > 48) return false;
    return /^(ai[\s-]?mode|ki[\s-]?modus|modus[\s-]?ki|mode[\s-]?ia|modo[\s-]?ia|modo de ia|modalit[aà][\s-]?ai)$/i.test(s)
      || /ai mode|ki-modus|mode ia|modo ia/i.test(s);
  };
  const nodes = document.querySelectorAll('a, button, [role="tab"], [role="link"], [role="button"], span, div');
  for (const n of nodes) {
    const label = n.getAttribute('aria-label') || '';
    const text = n.childElementCount > 6 ? '' : (n.innerText || '');
    if (!match(label) && !match(text)) continue;
    const clickable = n.closest('a, button, [role="tab"], [role="link"], [role="button"]') || n;
    clickable.click();
    return (label || text).replace(/\s+/g, ' ').trim();
  }
  return '';
})()`

const dismissConsentJS = `(() => {
  const re = /^(accept all|alle akzeptieren|tout accepter|aceptar todo|accetta tutto|i agree|agree)$/i;
  for (const b of document.querySelectorAll('button, [role="button"]')) {
    if (re.test((b.innerText || '').replace(/\s+/g, ' ').trim())) {
      b.click();
      return true;
    }
  }
  return false;
})()`

const extractJS = `(() => {
  const body = (document.body && document.body.innerText) || '';
  const blocked = /unusual traffic|are you a robot|detected unusual|enable javascript|sorry\/index|why did this happen/i.test(body + ' ' + location.href)
    || (/before you continue to google/i.test(body) && body.length < 1200);
  const urlAI = /[?&]udm=50(?:&|$)/.test(location.search);
  const snAI = !!(window.google && window.google.sn === 'aim');
  const pills = document.querySelectorAll('button[data-amic="true"], button[data-icl-uuid], [data-icl-uuid]').length;
  const follow = /ask a follow-up|ask anything|follow-up question/i.test(body);
  const isAI = urlAI || snAI || pills > 0 || follow;

  const junkHost = (u) => {
    try {
      const h = new URL(u, location.href).hostname;
      return /(^|\.)google\./.test(h) || /(^|\.)gstatic\.com$/.test(h) || h === 'g.co';
    } catch { return true; }
  };
  const sources = [];
  const seen = new Set();
  for (const a of document.querySelectorAll('a[href]')) {
    const href = a.href;
    if (!href || junkHost(href) || seen.has(href)) continue;
    const title = ((a.getAttribute('aria-label') || a.innerText || '')).replace(/\s+/g, ' ').trim();
    seen.add(href);
    sources.push({title: title.slice(0, 160), url: href});
  }

  const root = document.querySelector('[data-container-id], main, #rcnt, #center_col') || document.body;
  const answer = ((root && root.innerText) || body).replace(/\s+/g, ' ').trim();
  return {
    answer,
    sources: sources.slice(0, 16),
    isAI,
    ready: isAI && answer.length > 80 && !blocked,
    blocked: !!blocked
  };
})()`

const extractSERPJS = `(() => {
  const body = (document.body && document.body.innerText) || '';
  const blocked = /unusual traffic|are you a robot|detected unusual|enable javascript|sorry\/index|why did this happen/i.test(body + ' ' + location.href)
    || (/before you continue to google/i.test(body) && body.length < 1200);
  const junkHost = (u) => {
    try {
      const h = new URL(u, location.href).hostname;
      return /(^|\.)google\./.test(h) || /(^|\.)gstatic\.com$/.test(h) || h === 'g.co';
    } catch { return true; }
  };
  const items = [];
  const seen = new Set();
  const add = (title, href, snippet) => {
    try { href = new URL(href, location.href).href; } catch { return; }
    if (href.includes('/url?')) {
      try { href = new URL(href).searchParams.get('q') || href; } catch {}
    }
    if (!href || seen.has(href) || junkHost(href)) return;
    seen.add(href);
    items.push({
      title: (title || '').replace(/\s+/g, ' ').trim().slice(0, 160),
      url: href,
      snippet: (snippet || '').replace(/\s+/g, ' ').trim().slice(0, 240)
    });
  };
  for (const h of document.querySelectorAll('#search h3, #rso h3, #center_col h3, div[data-sokoban-container] h3')) {
    const a = h.closest('a');
    if (!a) continue;
    const block = (h.closest('div[data-sokoban-container], .g, [data-hveid]') || h.parentElement || {}).innerText || '';
    const extra = block.split('\n').filter((l) => l && l !== (h.innerText || '').trim()).slice(0, 2).join(' ');
    add(h.innerText, a.href || '', extra);
    if (items.length >= 8) break;
  }
  const lines = items.map((i) => i.title + (i.snippet ? ' - ' + i.snippet : ''));
  return {
    answer: lines.join('\n'),
    sources: items,
    isAI: false,
    ready: items.length > 0 && !blocked,
    blocked: !!blocked
  };
})()`

var (
	sv6CommentRE = regexp.MustCompile(`<!--Sv6Kpe(\[[\s\S]*?)-->`)
	httpURLRE    = regexp.MustCompile(`https?://[^\s"'\\<>]+`)
)

func sourcesFromHTML(html string) []Source {
	seen := map[string]bool{}
	var out []Source
	for _, m := range sv6CommentRE.FindAllStringSubmatch(html, -1) {
		for _, raw := range httpURLRE.FindAllString(m[1], -1) {
			u := sanitizeURL(raw)
			if u == "" || seen[u] || skipURL(u) {
				continue
			}
			seen[u] = true
			out = append(out, Source{Title: hostTitle(u), URL: u})
			if len(out) >= 12 {
				return out
			}
		}
	}
	return out
}

func jsSources(raw []map[string]string) []Source {
	seen := map[string]bool{}
	var out []Source
	for _, row := range raw {
		u := sanitizeURL(strings.TrimSpace(row["url"]))
		if u == "" || seen[u] || skipURL(u) {
			continue
		}
		seen[u] = true
		title := strings.TrimSpace(row["title"])
		if title == "" {
			title = hostTitle(u)
		}
		out = append(out, Source{Title: title, URL: u})
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func mergeSources(primary, extra []Source) []Source {
	seen := map[string]bool{}
	var out []Source
	add := func(s Source) {
		u := sanitizeURL(s.URL)
		if u == "" || seen[u] || skipURL(u) {
			return
		}
		seen[u] = true
		if strings.TrimSpace(s.Title) == "" {
			s.Title = hostTitle(u)
		}
		s.URL = u
		out = append(out, s)
	}
	for _, s := range primary {
		add(s)
	}
	for _, s := range extra {
		add(s)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func sanitizeURL(raw string) string {
	raw = strings.TrimRight(raw, "].),;\"'")
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	u.Fragment = ""
	return u.String()
}

func skipURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return true
	}
	h := strings.ToLower(u.Hostname())
	return strings.Contains(h, "google.") ||
		strings.HasSuffix(h, "gstatic.com") ||
		h == "g.co" ||
		strings.Contains(h, "googleusercontent.com") ||
		strings.Contains(h, "bing.com") ||
		strings.Contains(h, "duckduckgo.com")
}

func hostTitle(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	h := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if h == "" {
		return raw
	}
	return h
}

func cleanAnswer(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	for _, cut := range []string{
		"People also ask",
		"Related searches",
		"Weitere Fragen",
		"Ähnliche Suchanfragen",
	} {
		if i := strings.Index(s, cut); i > 160 {
			s = strings.TrimSpace(s[:i])
		}
	}
	if utf8.RuneCountInString(s) > maxAnswer {
		s = string([]rune(s)[:maxAnswer]) + "…"
	}
	return s
}
