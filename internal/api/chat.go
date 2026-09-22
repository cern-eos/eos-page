package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/apeters/eospage/internal/chat"
)

type chatLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newChatLimiter() *chatLimiter {
	return &chatLimiter{hits: map[string][]time.Time{}}
}

func (l *chatLimiter) allow(ip string) bool {
	now := time.Now()
	cut := now.Add(-10 * time.Minute)
	l.mu.Lock()
	defer l.mu.Unlock()
	keep := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cut) {
			keep = append(keep, t)
		}
	}
	if len(keep) >= 20 {
		l.hits[ip] = keep
		return false
	}
	l.hits[ip] = append(keep, now)
	return true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) handleChatStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": s.Chat != nil})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if s.Chat == nil {
		writeError(w, http.StatusServiceUnavailable, "Ask EOS is not configured on this server.")
		return
	}
	if !s.chatLimit.allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "Too many questions - try again in a few minutes.")
		return
	}
	var body struct {
		Message string         `json:"message"`
		History []chat.Message `json:"history"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 24<<10)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 75*time.Second)
	defer cancel()
	reply, err := s.Chat.Ask(ctx, body.Message, s.chatSiteContext(body.Message), body.History)
	if err != nil {
		msg := err.Error()
		status := http.StatusBadGateway
		if strings.Contains(msg, "empty question") || strings.Contains(msg, "too long") {
			status = http.StatusBadRequest
		}
		writeError(w, status, publicChatError(msg))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "text": reply.Text, "sources": reply.Sources})
}

func (s *Server) chatSiteContext(q string) string {
	if s.Store == nil {
		return ""
	}
	res, err := s.Store.Search(q, "all", 0, 4)
	if err != nil {
		return ""
	}
	var b strings.Builder
	for i, d := range res.Docs {
		if i >= 3 {
			break
		}
		b.WriteString("- Doc: ")
		b.WriteString(clipRunes(d.Title, 80))
		if d.URL != "" {
			b.WriteString(" - ")
			b.WriteString(d.URL)
		}
		if d.Summary != "" {
			b.WriteString(" - ")
			b.WriteString(clipRunes(d.Summary, 180))
		}
		b.WriteByte('\n')
	}
	for i, t := range res.Talks {
		if i >= 3 {
			break
		}
		b.WriteString("- Talk: ")
		b.WriteString(clipRunes(t.Title, 90))
		if t.WorkshopTitle != "" {
			b.WriteString(" (")
			b.WriteString(clipRunes(t.WorkshopTitle, 40))
			b.WriteByte(')')
		}
		if t.URL != "" {
			b.WriteString(" - ")
			b.WriteString(t.URL)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func clipRunes(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

func publicChatError(msg string) string {
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "empty question"):
		return "Please type a question."
	case strings.Contains(low, "too long"):
		return "That question is too long."
	case strings.Contains(low, "not configured"):
		return "Ask EOS is starting up - try again in a moment."
	default:
		return "Ask EOS could not answer just now. Try again, or write to eos-support@cern.ch."
	}
}
