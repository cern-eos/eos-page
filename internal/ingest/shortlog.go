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
	if asst, err := FetchGitAssisted(ctx, dir); err == nil {
		people = ApplyAssistedCounts(people, asst)
	}
	return people, nil
}

func FetchGitAssisted(ctx context.Context, dir string) ([]store.GitContributor, error) {
	if !isGitDir(dir) {
		return nil, fmt.Errorf("not a git repository: %s", dir)
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "log", "--all",
		"--regexp-ignore-case",
		"--grep=^Co-authored-by:",
		"--grep=^Assisted-by:",
		"--format=%an <%ae>")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log assisted: %w", err)
	}
	return ParseGitAssistedLog(string(out)), nil
}

func ParseGitAssistedLog(raw string) []store.GitContributor {
	type bucket struct {
		name, email string
		n           int
	}
	merged := map[string]*bucket{}
	order := []string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, email := splitNameEmail(line)
		if name == "" {
			continue
		}
		key := strings.ToLower(email)
		if key == "" {
			key = "name:" + strings.ToLower(name)
		}
		if b, ok := merged[key]; ok {
			b.n++
			continue
		}
		merged[key] = &bucket{name: name, email: email, n: 1}
		order = append(order, key)
	}
	out := make([]store.GitContributor, 0, len(order))
	for _, key := range order {
		b := merged[key]
		out = append(out, store.GitContributor{Name: b.name, Email: b.email, Assisted: b.n})
	}
	return out
}

func ApplyAssistedCounts(people []store.GitContributor, assisted []store.GitContributor) []store.GitContributor {
	counts := map[string]int{}
	for _, a := range assisted {
		if isAutomationIdentity(a.Name, a.Email) {
			continue
		}
		name, email := canonicalIdentity(a.Name, a.Email)
		if name == "" {
			continue
		}
		key := strings.ToLower(email)
		if key == "" {
			key = "name:" + strings.ToLower(name)
		}
		counts[key] += a.Assisted
	}
	for i, p := range people {
		key := strings.ToLower(p.Email)
		if key == "" {
			key = "name:" + strings.ToLower(p.Name)
		}
		people[i].Assisted = counts[key]
	}
	return people
}

func ParseGitShortlog(raw string) []store.GitContributor {
	var out []store.GitContributor
	for _, m := range shortlogLine.FindAllStringSubmatch(raw, -1) {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		name, email := splitNameEmail(m[2])
		if name == "" {
			continue
		}
		out = append(out, store.GitContributor{
			Commits: n,
			Name:    name,
			Email:   email,
		})
	}
	return NormalizeGitContributors(out)
}

func splitNameEmail(rest string) (string, string) {
	rest = strings.TrimSpace(rest)
	name, email := rest, ""
	if em := shortlogEmail.FindStringSubmatch(rest); len(em) == 2 {
		email = strings.TrimSpace(em[1])
		name = strings.TrimSpace(shortlogEmail.ReplaceAllString(rest, ""))
	}
	if name == "" {
		name = email
	}
	return name, email
}
