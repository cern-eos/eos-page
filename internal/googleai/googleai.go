// Package googleai runs Google Search in a real Chrome and returns the AI Mode
// answer plus cited sources. The UI control is found by label, not a brittle
// CSS path; udm=50 is the fallback when the tab is missing.
package googleai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type AIResult struct {
	Query   string   `json:"query"`
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
	IsAI    bool     `json:"isAI,omitempty"`
}

type Options struct {
	ProfileDir string
	Headed     bool
	Timeout    time.Duration
}

func GoogleAISearch(parent context.Context, query string) (AIResult, error) {
	return Search(parent, query, Options{})
}

func Search(parent context.Context, query string, opt Options) (AIResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return AIResult{}, fmt.Errorf("empty query")
	}
	if parent != nil && parent.Err() != nil {
		return AIResult{Query: query}, parent.Err()
	}

	if chromeSkipped() {
		return AIResult{Query: query}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}

	timeout := opt.Timeout
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	if parent != nil {
		if dl, ok := parent.Deadline(); ok {
			remain := time.Until(dl) - time.Second
			if remain > 0 && remain < timeout {
				timeout = remain
			}
		}
	}

	var result AIResult
	err := defaultSession.run(timeout, func(ctx context.Context) error {
		got, err := runSearch(ctx, query)
		result = got
		return err
	})
	result.Query = query
	if err != nil {
		if isBlockedErr(err) || isChromeGone(err) {
			markChromeUnusable()
		}
		return result, err
	}
	return result, nil
}

func ProfileDir() string {
	if p := strings.TrimSpace(os.Getenv("CHAT_CHROME_PROFILE")); p != "" {
		return p
	}
	data := strings.TrimSpace(os.Getenv("DATA_DIR"))
	if data == "" {
		data = "data"
	}
	return filepath.Join(data, "chrome-profile")
}

func WriteJSON(w io.Writer, result AIResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func searchURL(query string, aiMode bool) string {
	u := "https://www.google.com/search?hl=en&pws=0&q=" + url.QueryEscape(query)
	if aiMode {
		u += "&udm=50&aep=11"
	}
	return u
}

func runSearch(ctx context.Context, query string) (AIResult, error) {
	if err := navigate(ctx, searchURL(query, false)); err != nil {
		return AIResult{}, err
	}
	dismissConsent(ctx)
	if blocked, blockErr := pageBlocked(ctx); blockErr == nil && blocked {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}

	clicked := clickAIMode(ctx)
	if !clicked {
		if err := navigate(ctx, searchURL(query, true)); err != nil {
			return AIResult{}, err
		}
		dismissConsent(ctx)
		clickAIMode(ctx)
	}

	extracted, waitErr := waitForAnswer(ctx)
	if extracted.Blocked {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}
	if waitErr == nil {
		answer := cleanAnswer(extracted.Answer)
		if extracted.IsAI && answer != "" {
			var pageHTML string
			_ = chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery).Do(ctx)
			return AIResult{
				Answer:  answer,
				Sources: mergeSources(jsSources(extracted.Sources), sourcesFromHTML(pageHTML)),
				IsAI:    true,
			}, nil
		}
	}

	serp, err := extractSERP(ctx, query)
	if err == nil && strings.TrimSpace(serp.Answer) != "" {
		return serp, nil
	}
	if waitErr != nil {
		return AIResult{}, waitErr
	}
	if err != nil {
		return AIResult{}, err
	}
	return AIResult{}, fmt.Errorf("empty search results")
}

func pageBlocked(ctx context.Context) (bool, error) {
	var got extractResult
	if err := chromedp.Evaluate(extractJS, &got).Do(ctx); err != nil {
		return false, err
	}
	return got.Blocked, nil
}

func extractSERP(ctx context.Context, query string) (AIResult, error) {
	var got extractResult
	_ = chromedp.Evaluate(extractSERPJS, &got).Do(ctx)
	if got.Blocked {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}
	if len(got.Sources) < 2 {
		if err := navigate(ctx, searchURL(query, false)); err != nil {
			return AIResult{}, err
		}
		dismissConsent(ctx)
		_ = chromedp.Sleep(400 * time.Millisecond).Do(ctx)
		_ = chromedp.Evaluate(extractSERPJS, &got).Do(ctx)
	}
	if got.Blocked {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}
	sources := jsSources(got.Sources)
	answer := cleanAnswer(got.Answer)
	if answer == "" && len(sources) > 0 {
		var b strings.Builder
		for _, s := range sources {
			b.WriteString(s.Title)
			b.WriteString(" - ")
			b.WriteString(s.URL)
			b.WriteByte('\n')
		}
		answer = strings.TrimSpace(b.String())
	}
	if answer == "" {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	return AIResult{Answer: answer, Sources: sources, IsAI: false}, nil
}

func dismissConsent(ctx context.Context) {
	for _, sel := range []string{
		`#L2AGLb`,
		`button[aria-label="Accept all"]`,
		`button[aria-label="Alle akzeptieren"]`,
		`button[aria-label="Tout accepter"]`,
	} {
		if err := chromedp.Click(sel, chromedp.ByQuery).Do(ctx); err == nil {
			_ = chromedp.Sleep(400 * time.Millisecond).Do(ctx)
			return
		}
	}
	var clicked bool
	if err := chromedp.Evaluate(dismissConsentJS, &clicked).Do(ctx); err == nil && clicked {
		_ = chromedp.Sleep(500 * time.Millisecond).Do(ctx)
	}
}

func clickAIMode(ctx context.Context) bool {
	var label string
	if err := chromedp.Evaluate(clickAIModeJS, &label).Do(ctx); err != nil || strings.TrimSpace(label) == "" {
		return false
	}
	_ = chromedp.Sleep(400 * time.Millisecond).Do(ctx)
	return true
}

func waitForAnswer(ctx context.Context) (extractResult, error) {
	var last extractResult
	stable := 0
	lastLen := -1
	deadline := time.Now().Add(8 * time.Second)
	for {
		var got extractResult
		if err := chromedp.Evaluate(extractJS, &got).Do(ctx); err == nil {
			last = got
			if got.Blocked {
				return got, nil
			}
			n := len(strings.TrimSpace(got.Answer))
			if got.IsAI && n > 80 {
				if n == lastLen {
					stable++
					if stable >= 3 {
						return got, nil
					}
				} else {
					stable = 0
					lastLen = n
				}
			}
		}
		if time.Now().After(deadline) {
			if last.Answer != "" {
				return last, nil
			}
			return last, fmt.Errorf("timed out waiting for AI Mode")
		}
		if err := chromedp.Sleep(350 * time.Millisecond).Do(ctx); err != nil {
			return last, err
		}
	}
}
