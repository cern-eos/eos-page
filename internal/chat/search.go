package chat

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/chromedp/chromedp"
)

const maxSearchText = 8000

var searchMu sync.Mutex

func GoogleSearch(parent context.Context, query string) (string, []Source, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil, fmt.Errorf("empty query")
	}

	searchMu.Lock()
	defer searchMu.Unlock()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", os.Getenv("CHAT_SEARCH_HEADED") != "1"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(parent, opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	searchURL := "https://www.google.com/search?hl=en&num=8&q=" + url.QueryEscape(query)
	var body string
	var raw []map[string]string
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(dismissGoogleConsent),
		chromedp.Text("body", &body, chromedp.ByQuery),
		chromedp.Evaluate(`([...document.querySelectorAll('#search a h3')].slice(0,8).map((h) => {
			const a = h.closest('a');
			return {title: (h.innerText || '').trim(), url: a ? a.href : ''};
		}))`, &raw),
	)
	if err != nil {
		return "", nil, err
	}
	body = clipSearchText(body)
	if body == "" {
		return "", nil, fmt.Errorf("empty search page")
	}
	return body, searchSources(raw), nil
}

func dismissGoogleConsent(ctx context.Context) error {
	for _, sel := range []string{`#L2AGLb`, `button[aria-label="Accept all"]`, `button[aria-label="Alle akzeptieren"]`} {
		if err := chromedp.Click(sel, chromedp.ByQuery).Do(ctx); err == nil {
			_ = chromedp.Sleep(1200 * time.Millisecond).Do(ctx)
			return nil
		}
	}
	return nil
}

func clipSearchText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= maxSearchText {
		return s
	}
	return string([]rune(s)[:maxSearchText]) + "…"
}

func searchSources(raw []map[string]string) []Source {
	seen := map[string]bool{}
	var out []Source
	for _, row := range raw {
		u := strings.TrimSpace(row["url"])
		title := strings.TrimSpace(row["title"])
		if u == "" || seen[u] || strings.Contains(u, "google.com") {
			continue
		}
		seen[u] = true
		if title == "" {
			title = u
		}
		out = append(out, Source{Title: title, URL: u})
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func searchQuery(question string) string {
	q := strings.TrimSpace(question)
	low := strings.ToLower(q)
	if !strings.Contains(low, "eos") && !strings.Contains(low, "cern") {
		q += " EOS CERN storage"
	}
	return q
}
