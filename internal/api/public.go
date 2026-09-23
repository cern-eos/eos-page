package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	settings, err := s.Store.Settings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	pages, err := s.Store.ListPages()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cards, err := s.Store.ListCards("", true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	news, err := s.Store.ListNews(true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	people, err := s.Store.ListPeople(true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	workshops, err := s.Store.ListWorkshops(true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	docs, err := s.Store.FeaturedDocs(10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	years, _ := s.Store.WorkshopYears()
	contributors, _ := s.Store.ListGitContributors()
	writeJSON(w, http.StatusOK, map[string]any{
		"settings":     settings,
		"pages":        pages,
		"cards":        cards,
		"news":         news,
		"people":       people,
		"workshops":    workshops,
		"docs":         docs,
		"years":        years,
		"contributors": contributors,
		"index":        s.Store.IndexStats(),
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	res, err := s.Store.Search(q, kind, year, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleNewsletter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sub, err := s.Store.Subscribe(body.Email)
	if err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": sub.Status})
}

func (s *Server) handleInbox(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	msg, err := s.Store.AddInbox("contact", body.Name, body.Email, body.Message)
	if err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": msg.ID})
}
