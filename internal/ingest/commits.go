package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/store"
)

const defaultCommitBranch = "master"

type CommitSource struct {
	GitHubURL string
	GitLabURL string
	Branch    string
	Token     string
	GitDir    string
	MaxPages  int
}

func FetchMasterCommits(ctx context.Context, src CommitSource) ([]store.Commit, string, error) {
	if src.Branch == "" {
		src.Branch = defaultCommitBranch
	}
	if src.MaxPages <= 0 {
		src.MaxPages = 15
	}
	if src.GitDir != "" {
		commits, err := fetchGitLog(ctx, src)
		if err == nil && len(commits) > 0 {
			return commits, "git " + src.Branch, nil
		}
	}
	commits, err := fetchGitHubCommits(ctx, src)
	if err == nil && len(commits) > 0 {
		return commits, "github " + src.Branch, nil
	}
	ghErr := err
	commits, err = fetchGitLabCommits(ctx, src)
	if err == nil && len(commits) > 0 {
		return commits, "gitlab " + src.Branch, nil
	}
	if ghErr != nil {
		return nil, "", fmt.Errorf("github: %v; gitlab: %w", ghErr, err)
	}
	if err != nil {
		return nil, "", err
	}
	return nil, "", fmt.Errorf("no commits on %s", src.Branch)
}

func SplitCommitMessage(msg string) (title, body string) {
	msg = strings.ReplaceAll(strings.ReplaceAll(msg, " — ", " - "), "—", "-")
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.TrimSpace(msg)
	title, rest, ok := strings.Cut(msg, "\n")
	title = strings.TrimSpace(title)
	if ok {
		body = strings.TrimSpace(rest)
	}
	if title == "" {
		title = "(no subject)"
	}
	if len(body) > 8000 {
		body = body[:8000]
	}
	return
}

func fetchGitHubCommits(ctx context.Context, src CommitSource) ([]store.Commit, error) {
	owner, repo := parseGitHubRepo(src.GitHubURL)
	client := &http.Client{Timeout: 45 * time.Second}
	var out []store.Commit
	for page := 1; page <= src.MaxPages; page++ {
		u := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=100&page=%d", owner, repo, src.Branch, page)
		body, status, err := getBytes(ctx, client, u, src.Token, "application/vnd.github+json")
		if err != nil {
			return out, err
		}
		if status == http.StatusNotFound || status == http.StatusConflict || status == 422 {
			return out, fmt.Errorf("github %s branch %s: %d", owner+"/"+repo, src.Branch, status)
		}
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			if len(out) > 0 {
				return out, nil
			}
			return nil, fmt.Errorf("github rate limited (%d)", status)
		}
		if status != http.StatusOK {
			return out, fmt.Errorf("github commits: %d", status)
		}
		var raw []struct {
			SHA     string `json:"sha"`
			HTMLURL string `json:"html_url"`
			Commit  struct {
				Message string `json:"message"`
				Author  struct {
					Name string `json:"name"`
					Date string `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return out, err
		}
		if len(raw) == 0 {
			break
		}
		for _, row := range raw {
			title, msg := SplitCommitMessage(row.Commit.Message)
			out = append(out, store.Commit{
				SHA:    row.SHA,
				Short:  shortSHA(row.SHA),
				Date:   row.Commit.Author.Date,
				Year:   commitYear(row.Commit.Author.Date),
				Author: strings.TrimSpace(row.Commit.Author.Name),
				Title:  title,
				Body:   msg,
				URL:    row.HTMLURL,
			})
		}
		if len(raw) < 100 {
			break
		}
	}
	return out, nil
}

func fetchGitLabCommits(ctx context.Context, src CommitSource) ([]store.Commit, error) {
	host, project := parseGitLabProject(src.GitLabURL)
	client := &http.Client{Timeout: 45 * time.Second}
	var out []store.Commit
	for page := 1; page <= src.MaxPages; page++ {
		u := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits?ref_name=%s&per_page=100&page=%d",
			host, project, src.Branch, page)
		body, status, err := getBytes(ctx, client, u, "", "application/json")
		if err != nil {
			return out, err
		}
		if status != http.StatusOK {
			return out, fmt.Errorf("gitlab commits: %d", status)
		}
		var raw []struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Message      string `json:"message"`
			AuthorName   string `json:"author_name"`
			AuthoredDate string `json:"authored_date"`
			WebURL       string `json:"web_url"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return out, err
		}
		if len(raw) == 0 {
			break
		}
		for _, row := range raw {
			title, msg := SplitCommitMessage(row.Message)
			if title == "(no subject)" && row.Title != "" {
				title = strings.TrimSpace(row.Title)
			}
			out = append(out, store.Commit{
				SHA:    row.ID,
				Short:  shortSHA(row.ID),
				Date:   row.AuthoredDate,
				Year:   commitYear(row.AuthoredDate),
				Author: strings.TrimSpace(row.AuthorName),
				Title:  title,
				Body:   msg,
				URL:    row.WebURL,
			})
		}
		if len(raw) < 100 {
			break
		}
	}
	return out, nil
}

func fetchGitLog(ctx context.Context, src CommitSource) ([]store.Commit, error) {
	dir := src.GitDir
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, fmt.Errorf("git dir: %w", err)
	}
	ref := src.Branch
	_ = exec.CommandContext(ctx, "git", "-C", dir, "fetch", "--quiet", "origin", src.Branch).Run()
	if exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--verify", "origin/"+src.Branch).Run() == nil {
		ref = "origin/" + src.Branch
	}
	limit := src.MaxPages * 100
	if limit <= 0 {
		limit = 1500
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "log", ref, fmt.Sprintf("-n%d", limit),
		"--format=%H%x1f%aI%x1f%an%x1f%B%x1e")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	owner, repo := parseGitHubRepo(src.GitHubURL)
	var commits []store.Commit
	for _, rec := range strings.Split(string(out), "\x1e") {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		parts := strings.SplitN(rec, "\x1f", 4)
		if len(parts) < 4 {
			continue
		}
		title, body := SplitCommitMessage(parts[3])
		sha := strings.TrimSpace(parts[0])
		commits = append(commits, store.Commit{
			SHA:    sha,
			Short:  shortSHA(sha),
			Date:   strings.TrimSpace(parts[1]),
			Year:   commitYear(parts[1]),
			Author: strings.TrimSpace(parts[2]),
			Title:  title,
			Body:   body,
			URL:    fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, sha),
		})
	}
	return commits, nil
}

func getBytes(ctx context.Context, client *http.Client, rawURL, token, accept string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Accept", accept)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 12<<20))
	if err != nil {
		return nil, res.StatusCode, err
	}
	return body, res.StatusCode, nil
}

func parseGitHubRepo(raw string) (owner, repo string) {
	owner, repo = "cern-eos", "eos"
	raw = strings.TrimSuffix(strings.TrimSpace(raw), ".git")
	raw = strings.TrimSuffix(raw, "/")
	const needle = "github.com/"
	i := strings.Index(raw, needle)
	if i < 0 {
		return
	}
	parts := strings.Split(raw[i+len(needle):], "/")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1]
	}
	return
}

func parseGitLabProject(raw string) (host, project string) {
	host, project = "https://gitlab.cern.ch", "dss%2Feos"
	raw = strings.TrimSuffix(strings.TrimSpace(raw), ".git")
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return
	}
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	slash := strings.Index(raw, "/")
	if slash < 0 {
		return
	}
	host = "https://" + raw[:slash]
	path := strings.Trim(raw[slash+1:], "/")
	if path != "" {
		project = strings.ReplaceAll(path, "/", "%2F")
	}
	return
}

func commitYear(iso string) int {
	iso = strings.TrimSpace(iso)
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t.Year()
		}
	}
	if len(iso) >= 4 {
		var y int
		if _, err := fmt.Sscanf(iso[:4], "%d", &y); err == nil {
			return y
		}
	}
	return 0
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
