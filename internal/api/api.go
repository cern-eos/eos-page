package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/chat"
	"github.com/apeters/eospage/internal/store"
)

const controllerCookie = "eos_controller"

type Options struct {
	ControllerSecret string
	PublicBaseURL    string
	Chat             *chat.Client
}

type Server struct {
	Store            *store.Store
	ControllerSecret string
	PublicBaseURL    string
	Chat             *chat.Client
	chatLimit        *chatLimiter
	PublicFS         fs.FS
	ControllerFS     fs.FS
}

func New(st *store.Store, opt Options, publicFS, controllerFS fs.FS) *Server {
	return &Server{
		Store:            st,
		ControllerSecret: opt.ControllerSecret,
		PublicBaseURL:    strings.TrimRight(opt.PublicBaseURL, "/"),
		Chat:             opt.Chat,
		chatLimit:        newChatLimiter(),
		PublicFS:         publicFS,
		ControllerFS:     controllerFS,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/catalog", s.handleCatalog)
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("POST /api/newsletter", s.handleNewsletter)
	mux.HandleFunc("POST /api/inbox", s.handleInbox)
	mux.HandleFunc("GET /api/chat", s.handleChatStatus)
	mux.HandleFunc("POST /api/chat", s.handleChat)
	mux.HandleFunc("GET /api/docs/view", s.handleDocsView)

	mux.HandleFunc("POST /api/controller/login", s.handleControllerLogin)
	mux.HandleFunc("POST /api/controller/logout", s.handleControllerLogout)
	mux.HandleFunc("GET /api/controller/state", s.handleControllerState)
	mux.HandleFunc("PATCH /api/controller/settings", s.handleControllerSettings)
	mux.HandleFunc("PUT /api/controller/pages/{id}", s.handleControllerPage)
	mux.HandleFunc("POST /api/controller/cards", s.handleControllerUpsertCard)
	mux.HandleFunc("PATCH /api/controller/cards/{id}", s.handleControllerUpsertCard)
	mux.HandleFunc("DELETE /api/controller/cards/{id}", s.handleControllerDeleteCard)
	mux.HandleFunc("POST /api/controller/news", s.handleControllerUpsertNews)
	mux.HandleFunc("PATCH /api/controller/news/{id}", s.handleControllerUpsertNews)
	mux.HandleFunc("DELETE /api/controller/news/{id}", s.handleControllerDeleteNews)
	mux.HandleFunc("POST /api/controller/people", s.handleControllerUpsertPerson)
	mux.HandleFunc("PATCH /api/controller/people/{id}", s.handleControllerUpsertPerson)
	mux.HandleFunc("DELETE /api/controller/people/{id}", s.handleControllerDeletePerson)
	mux.HandleFunc("POST /api/controller/refresh/indico", s.handleRefreshIndico)
	mux.HandleFunc("POST /api/controller/refresh/docs", s.handleRefreshDocs)
	mux.HandleFunc("POST /api/controller/refresh/seed", s.handleReloadSeed)
	mux.HandleFunc("PATCH /api/controller/subscribers/{id}", s.handleControllerSubscriber)

	mux.HandleFunc("GET /media/{name}", s.handleMedia)
	mux.HandleFunc("GET /controller", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/controller/", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /controller/", s.serveController)
	mux.HandleFunc("GET /static/", s.servePublic)
	mux.HandleFunc("/", s.servePublic)

	return mux
}

func (s *Server) servePublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/")
	if strings.HasPrefix(p, "static/") {
		p = strings.TrimPrefix(p, "static/")
	}
	if p == "" || !strings.Contains(p, ".") {
		p = "index.html"
	}
	s.serveFS(w, r, s.PublicFS, p)
}

func (s *Server) serveController(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/controller/")
	if strings.HasPrefix(p, "static/") {
		p = strings.TrimPrefix(p, "static/")
	}
	if p == "" || !strings.Contains(p, ".") {
		p = "index.html"
	}
	s.serveFS(w, r, s.ControllerFS, p)
}

func (s *Server) serveFS(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	name = path.Clean("/" + name)[1:]
	if name == "" || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	if fsys == s.PublicFS {
		disk := filepath.Join("web", "public", filepath.FromSlash(name))
		if st, err := os.Stat(disk); err == nil && !st.IsDir() {
			http.ServeFile(w, r, disk)
			return
		}
	}
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func (s *Server) handleMedia(w http.ResponseWriter, r *http.Request) {
	name := path.Base(r.PathValue("name"))
	if name == "." || name == "/" {
		http.NotFound(w, r)
		return
	}
	disk := filepath.Join(s.Store.MediaDir(), name)
	if st, err := os.Stat(disk); err == nil && !st.IsDir() {
		http.ServeFile(w, r, disk)
		return
	}
	s.serveFS(w, r, s.PublicFS, path.Join("media", name))
}

func sessionCookie(name, value string, maxAge int, r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name: name, Value: value, Path: "/", HttpOnly: true, Secure: r.TLS != nil,
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	}
}

func (s *Server) requireController(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie(controllerCookie)
	if err != nil || !s.Store.ValidControllerSession(c.Value) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func storeStatus(err error) int {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, store.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, store.ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
