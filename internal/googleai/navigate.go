package googleai

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// navigate opens url without waiting for window "load". Google Search and
// AI Mode often never fire that event, so chromedp.Navigate hangs until timeout.
func navigate(ctx context.Context, rawURL string) error {
	nav, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	err := chromedp.Navigate(rawURL).Do(nav)
	if err == nil {
		return nil
	}
	// Google Search often never fires "load"; the document may still be usable.
	if nav.Err() != nil && ctx.Err() == nil {
		return nil
	}
	_, _, errText, _, rawErr := page.Navigate(rawURL).Do(ctx)
	if rawErr != nil {
		return rawErr
	}
	if errText != "" {
		return fmt.Errorf("page load error %s", errText)
	}
	return nil
}

func chromeUA() string {
	if runtime.GOOS == "linux" {
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	}
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
}
