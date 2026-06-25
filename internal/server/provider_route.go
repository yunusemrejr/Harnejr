package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/providers"
)

func (s *Server) handleProviderRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req providers.RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, providers.Route(registry, req))
}
