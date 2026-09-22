package googleai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLockPID(t *testing.T) {
	if got := lockPID("andreass-mbp-4.home-55828"); got != 55828 {
		t.Fatalf("got %d", got)
	}
	if lockPID("nope") != 0 || lockPID("") != 0 {
		t.Fatal("expected 0")
	}
}

func TestClearStaleProfileLocks(t *testing.T) {
	dir := t.TempDir()
	lock := filepath.Join(dir, "SingletonLock")
	if err := os.Symlink("dead-host-99999999", lock); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "DevToolsActivePort"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	clearStaleProfileLocks(dir)
	if _, err := os.Lstat(lock); !os.IsNotExist(err) {
		t.Fatalf("stale lock still there: %v", err)
	}
}
