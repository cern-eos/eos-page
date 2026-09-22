package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/apeters/eospage/internal/ingest"
	"github.com/apeters/eospage/internal/store"
)

func (s *Server) handleControllerLogin(w http.ResponseWriter, r *http.Request) {
	if s.ControllerSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "CONTROLLER_SECRET not configured")
		return
	}
	var body struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Secret != s.ControllerSecret {
		writeError(w, http.StatusUnauthorized, "invalid secret")
		return
	}
	sid, err := s.Store.CreateControllerSession()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.SetCookie(w, sessionCookie(controllerCookie, sid, 60*60*12, r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(controllerCookie); err == nil {
		s.Store.RevokeControllerSession(c.Value)
	}
	http.SetCookie(w, sessionCookie(controllerCookie, "", -1, r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerState(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	settings, err := s.Store.Settings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	pages, _ := s.Store.ListPages()
	cards, _ := s.Store.ListCards("", false)
	news, _ := s.Store.ListNews(false)
	people, _ := s.Store.ListPeople(false)
	workshops, _ := s.Store.ListWorkshops(false)
	subs, _ := s.Store.ListSubscribers()
	inbox, _ := s.Store.ListInbox()
	writeJSON(w, http.StatusOK, map[string]any{
		"settings":    settings,
		"pages":       pages,
		"cards":       cards,
		"news":        news,
		"people":      people,
		"workshops":   workshops,
		"subscribers": subs,
		"inbox":       inbox,
		"index":       s.Store.IndexStats(),
	})
}

func (s *Server) handleControllerSettings(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	for k, v := range body {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if err := s.Store.SetSetting(k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerPage(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var p store.Page
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p.ID = r.PathValue("id")
	if err := s.Store.UpsertPage(p); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleControllerUpsertCard(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var c store.Card
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if id := r.PathValue("id"); id != "" {
		c.ID = id
	}
	if err := s.Store.UpsertCard(c); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleControllerDeleteCard(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	if err := s.Store.DeleteCard(r.PathValue("id")); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerUpsertNews(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var n store.News
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if id := r.PathValue("id"); id != "" {
		n.ID = id
	}
	if err := s.Store.UpsertNews(n); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *Server) handleControllerDeleteNews(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	if err := s.Store.DeleteNews(r.PathValue("id")); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerUpsertPerson(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var p store.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if id := r.PathValue("id"); id != "" {
		p.ID = id
	}
	if err := s.Store.UpsertPerson(p); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleControllerDeletePerson(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	if err := s.Store.DeletePerson(r.PathValue("id")); err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleControllerSubscriber(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	out, err := s.Store.SetSubscriberStatus(r.PathValue("id"), body.Status)
	if err != nil {
		writeError(w, storeStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleRefreshIndico(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	ids := ingest.ParseEventIDs(s.Store.Setting("indico_event_ids"))
	workshops, talks, err := ingest.FetchWorkshops(ids)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	for _, ws := range workshops {
		if err := s.Store.UpsertWorkshop(ws); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := s.Store.ReplaceTalks(talks); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "workshops": len(workshops), "talks": len(talks), "index": s.Store.IndexStats()})
}

func (s *Server) handleRefreshDocs(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	docs, err := ingest.FetchDocs()
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := s.Store.ReplaceDocs(docs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "docs": len(docs), "index": s.Store.IndexStats()})
}

func (s *Server) handleReloadSeed(w http.ResponseWriter, r *http.Request) {
	if !s.requireController(w, r) {
		return
	}
	if err := s.Store.LoadSeedIndex(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "index": s.Store.IndexStats()})
}
