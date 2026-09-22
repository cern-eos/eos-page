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

	timeout := opt.Timeout
	if timeout <= 0 {
		timeout = 12 * time.Second
	}

	var result AIResult
	err := defaultSession.run(timeout, func(ctx context.Context) error {
		got, err := runAIMode(ctx, query)
		result = got
		return err
	})
	result.Query = query
	if err != nil {
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
	u := "https://www.google.com/search?hl=en&q=" + url.QueryEscape(query)
	if aiMode {
		u += "&udm=50&aep=11"
	}
	return u
}

func runAIMode(ctx context.Context, query string) (AIResult, error) {
	if err := chromedp.Navigate(searchURL(query, false)).Do(ctx); err != nil {
		return AIResult{}, err
	}
	_ = chromedp.Sleep(400 * time.Millisecond).Do(ctx)
	dismissConsent(ctx)

	clicked := clickAIMode(ctx)
	if !clicked {
		if err := chromedp.Navigate(searchURL(query, true)).Do(ctx); err != nil {
			return AIResult{}, err
		}
		_ = chromedp.Sleep(400 * time.Millisecond).Do(ctx)
		dismissConsent(ctx)
		clickAIMode(ctx)
	}

	extracted, err := waitForAnswer(ctx)
	if err != nil {
		return AIResult{}, err
	}
	if extracted.Blocked {
		return AIResult{}, fmt.Errorf("google blocked the browser (consent or captcha)")
	}

	var pageHTML string
	_ = chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery).Do(ctx)

	sources := mergeSources(jsSources(extracted.Sources), sourcesFromHTML(pageHTML))
	answer := cleanAnswer(extracted.Answer)
	if answer == "" {
		return AIResult{}, fmt.Errorf("empty AI Mode answer")
	}
	return AIResult{Answer: answer, Sources: sources}, nil
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
		_ = chromedp.Sleep(900 * time.Millisecond).Do(ctx)
	}
}

func clickAIMode(ctx context.Context) bool {
	var label string
	if err := chromedp.Evaluate(clickAIModeJS, &label).Do(ctx); err != nil || strings.TrimSpace(label) == "" {
		return false
	}
	_ = chromedp.Sleep(500 * time.Millisecond).Do(ctx)
	return true
}

func waitForAnswer(ctx context.Context) (extractResult, error) {
	var last extractResult
	stable := 0
	lastLen := -1
	deadline := time.Now().Add(6 * time.Second)
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
		if err := chromedp.Sleep(450 * time.Millisecond).Do(ctx); err != nil {
			return last, err
		}
	}
}
