package googleai

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
)

type session struct {
	mu          sync.Mutex
	allocCancel context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

var defaultSession session

func (s *session) run(timeout time.Duration, fn func(context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensure(); err != nil {
		return err
	}
	tab, cancel := chromedp.NewContext(s.ctx)
	defer cancel()
	if timeout <= 0 {
		timeout = 22 * time.Second
	}
	tab, cancel = context.WithTimeout(tab, timeout)
	defer cancel()
	if err := chromedp.Run(tab); err != nil {
		if s.ctx.Err() != nil || isChromeGone(err) {
			s.reset()
		}
		return err
	}
	err := fn(tab)
	if err != nil && (s.ctx.Err() != nil || isChromeGone(err)) {
		s.reset()
	}
	return err
}

func (s *session) ensure() error {
	if s.ctx != nil && s.ctx.Err() == nil {
		if s.healthy() {
			return nil
		}
		s.reset()
	}
	s.reset()
	profile := ProfileDir()
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return err
	}
	clearStaleProfileLocks(profile)

	headed := os.Getenv("CHAT_SEARCH_HEADED") == "1"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(profile),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.UserAgent(chromeUA()),
	)
	if headed {
		opts = append(opts, chromedp.Flag("headless", false))
	} else {
		opts = append(opts, chromedp.Flag("headless", "new"))
	}
	if bin := strings.TrimSpace(os.Getenv("CHAT_CHROME_BIN")); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	start, startCancel := context.WithTimeout(ctx, 12*time.Second)
	defer startCancel()
	if err := chromedp.Run(start); err != nil {
		cancel()
		allocCancel()
		return err
	}
	s.allocCancel = allocCancel
	s.ctx = ctx
	s.cancel = cancel
	return nil
}

func (s *session) healthy() bool {
	if s.ctx == nil || s.ctx.Err() != nil {
		return false
	}
	probe, cancel := context.WithTimeout(s.ctx, 2*time.Second)
	defer cancel()
	var n int
	if err := chromedp.Evaluate("1+1", &n).Do(probe); err != nil {
		return false
	}
	return n == 2
}

func (s *session) reset() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.allocCancel != nil {
		s.allocCancel()
	}
	s.ctx = nil
	s.cancel = nil
	s.allocCancel = nil
}

func clearStaleProfileLocks(dir string) {
	lock := filepath.Join(dir, "SingletonLock")
	pid := 0
	if target, err := os.Readlink(lock); err == nil {
		pid = lockPID(target)
	} else if data, err := os.ReadFile(lock); err == nil {
		pid = lockPID(strings.TrimSpace(string(data)))
	} else {
		return
	}
	if pid > 0 && pidAlive(pid) {
		return
	}
	_ = os.Remove(lock)
	_ = os.Remove(filepath.Join(dir, "SingletonCookie"))
	_ = os.Remove(filepath.Join(dir, "SingletonSocket"))
	_ = os.Remove(filepath.Join(dir, "DevToolsActivePort"))
}

func lockPID(target string) int {
	i := strings.LastIndex(target, "-")
	if i < 0 || i+1 >= len(target) {
		return 0
	}
	pid, err := strconv.Atoi(target[i+1:])
	if err != nil {
		return 0
	}
	return pid
}

func pidAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}

var (
	skipChromeMu    sync.Mutex
	skipChromeUntil time.Time
)

func chromeSkipped() bool {
	skipChromeMu.Lock()
	defer skipChromeMu.Unlock()
	return time.Now().Before(skipChromeUntil)
}

func markChromeUnusable() {
	skipChromeMu.Lock()
	skipChromeUntil = time.Now().Add(10 * time.Minute)
	skipChromeMu.Unlock()
}

func isBlockedErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "captcha") || strings.Contains(s, "blocked the browser") || strings.Contains(s, "unusual traffic")
}

func isChromeGone(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "websocket") ||
		strings.Contains(s, "connection") ||
		strings.Contains(s, "context canceled") ||
		strings.Contains(s, "context deadline") ||
		(strings.Contains(s, "chrome") && strings.Contains(s, "killed")) ||
		strings.Contains(s, "ctx.done") ||
		strings.Contains(s, "use of closed")
}
