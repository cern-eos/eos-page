package api

import (
	"net/http"
	"strings"

	"github.com/apeters/eospage/internal/docsview"
)

func (s *Server) handleDocsView(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" {
		writeError(w, http.StatusBadRequest, "missing url")
		return
	}
	page, err := docsview.Fetch(r.Context(), raw)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, page)
}
