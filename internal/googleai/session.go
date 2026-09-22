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
		timeout = 12 * time.Second
	}
	tab, cancel = context.WithTimeout(tab, timeout)
	defer cancel()
	err := chromedp.Run(tab, chromedp.ActionFunc(fn))
	if err != nil && s.ctx.Err() != nil {
		s.reset()
	}
	return err
}

func (s *session) ensure() error {
	if s.ctx != nil && s.ctx.Err() == nil {
		return nil
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
		chromedp.Flag("headless", !headed),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	start, startCancel := context.WithTimeout(ctx, 8*time.Second)
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
	target, err := os.Readlink(lock)
	if err != nil {
		return
	}
	pid := lockPID(target)
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
