package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/ingest"
)

const commitRefresh = 45 * time.Second

func (s *Server) StartCommitSync() {
	go s.refreshGitContributors(context.Background())
	go func() {
		ctx := context.Background()
		if err := s.refreshCommits(ctx, true); err != nil {
			log.Printf("commits: initial master fetch: %v", err)
		}
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for range t.C {
			if err := s.refreshCommits(ctx, false); err != nil {
				log.Printf("commits: live refresh: %v", err)
			}
		}
	}()
}

func (s *Server) refreshCommits(ctx context.Context, full bool) error {
	s.commitMu.Lock()
	defer s.commitMu.Unlock()
	if !full && !s.commitLast.IsZero() && time.Since(s.commitLast) < commitRefresh {
		return nil
	}
	pages := 2
	if full || s.Store.CommitCount() < 50 {
		pages = 15
	}
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GH_TOKEN"))
	}
	commits, src, err := ingest.FetchMasterCommits(ctx, ingest.CommitSource{
		GitHubURL: s.Store.Setting("github"),
		GitLabURL: s.Store.Setting("gitlab"),
		Branch:    "master",
		Token:     token,
		GitDir:    strings.TrimSpace(os.Getenv("EOS_GIT_DIR")),
		MaxPages:  pages,
	})
	if err != nil {
		return err
	}
	if err := s.Store.UpsertCommits(commits); err != nil {
		return err
	}
	s.commitLast = time.Now()
	_ = s.Store.SetSetting("commits_fetched_at", s.commitLast.UTC().Format(time.RFC3339))
	_ = s.Store.SetSetting("commits_source", src)
	log.Printf("commits: synced %d from %s (total %d)", len(commits), src, s.Store.CommitCount())
	return nil
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	full := s.Store.CommitCount() == 0
	if err := s.refreshCommits(r.Context(), full); err != nil && s.Store.CommitCount() == 0 {
		writeError(w, http.StatusBadGateway, "could not load the live master log: "+err.Error())
		return
	}
	commits, err := s.Store.SearchCommits(q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"query":     q,
		"branch":    "master",
		"source":    s.Store.Setting("commits_source"),
		"fetchedAt": s.Store.Setting("commits_fetched_at"),
		"total":     s.Store.CommitCount(),
		"commits":   commits,
	})
}

func (s *Server) refreshGitContributors(ctx context.Context) {
	dir := ingest.FindEOSGitDir()
	if dir == "" {
		return
	}
	people, err := ingest.FetchGitShortlog(ctx, dir)
	if err != nil {
		log.Printf("wall of fame: git shortlog failed: %v", err)
		return
	}
	if err := s.Store.ReplaceGitContributors(people); err != nil {
		log.Printf("wall of fame: store failed: %v", err)
		return
	}
	log.Printf("wall of fame: loaded %d contributors from git shortlog -sne --all (%s)", len(people), dir)
}
