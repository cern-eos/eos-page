package ingest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/apeters/eospage/internal/store"
)

var shortlogLine = regexp.MustCompile(`(?m)^\s*(\d+)\t(.+)$`)
var shortlogEmail = regexp.MustCompile(`\s*<([^>]+)>\s*$`)

func FindEOSGitDir() string {
	if d := strings.TrimSpace(os.Getenv("EOS_GIT_DIR")); d != "" {
		if isGitDir(d) {
			return d
		}
	}
	for _, cand := range []string{"../eos", "eos", filepath.Join("..", "EOS")} {
		if isGitDir(cand) {
			return cand
		}
	}
	return ""
}

func isGitDir(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil && st != nil {
		return true
	}
	if st, err := os.Stat(filepath.Join(dir, "HEAD")); err == nil && !st.IsDir() {
		return true
	}
	return false
}

func FetchGitShortlog(ctx context.Context, dir string) ([]store.GitContributor, error) {
	if !isGitDir(dir) {
		return nil, fmt.Errorf("not a git repository: %s", dir)
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "shortlog", "-sne", "--all")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git shortlog: %w", err)
	}
	people := ParseGitShortlog(string(out))
	if len(people) == 0 {
		return nil, fmt.Errorf("empty git shortlog")
	}
	return people, nil
}

func ParseGitShortlog(raw string) []store.GitContributor {
	var out []store.GitContributor
	for _, m := range shortlogLine.FindAllStringSubmatch(raw, -1) {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		rest := strings.TrimSpace(m[2])
		name, email := rest, ""
		if em := shortlogEmail.FindStringSubmatch(rest); len(em) == 2 {
			email = strings.TrimSpace(em[1])
			name = strings.TrimSpace(shortlogEmail.ReplaceAllString(rest, ""))
		}
		if name == "" {
			name = email
		}
		if name == "" || isRootContributor(name, email) {
			continue
		}
		out = append(out, store.GitContributor{
			Sort:    len(out) + 1,
			Commits: n,
			Name:    name,
			Email:   email,
		})
	}
	return out
}

func isRootContributor(name, email string) bool {
	if strings.EqualFold(strings.TrimSpace(name), "root") {
		return true
	}
	local, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	return local == "root"
}
