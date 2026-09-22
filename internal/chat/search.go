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

	"github.com/apeters/eospage/internal/googleai"
	"github.com/chromedp/chromedp"
)

const maxSearchText = 8000

var searchMu sync.Mutex

func liveSearch(parent context.Context, question string) (ai bool, text string, sources []Source, err error) {
	q := searchQuery(question)
	result, err := googleai.GoogleAISearch(parent, q)
	if err == nil && strings.TrimSpace(result.Answer) != "" {
		return true, result.Answer, toChatSources(result.Sources), nil
	}
	text, sources, fallbackErr := GoogleSearch(parent, q)
	if fallbackErr != nil {
		if err != nil {
			return false, "", nil, err
		}
		return false, "", nil, fallbackErr
	}
	return false, text, sources, nil
}

func toChatSources(in []googleai.Source) []Source {
	out := make([]Source, 0, len(in))
	for _, s := range in {
		out = append(out, Source{Title: s.Title, URL: s.URL})
	}
	return out
}

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

const eosQuestionTag = "This question is related to CERN's EOS Storage System."

func searchQuery(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return q
	}
	low := strings.ToLower(q)
	if strings.Contains(low, "cern's eos storage system") || strings.Contains(low, "cerns eos storage system") {
		return q
	}
	return q + " — " + eosQuestionTag
}
