package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type resetProviderHealthRequest struct {
	ProviderID string `json:"providerId"`
}

func (s *Server) providerHealthPath() string {
	root := strings.TrimSpace(s.configDir)
	if root == "" {
		if dir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(dir) != "" {
			root = filepath.Join(dir, "harnejr")
		} else {
			root = ".harnejr-config"
		}
	}
	return filepath.Join(root, "state", "provider-health.json")
}

func (s *Server) handleProviderHealth(w http.ResponseWriter, r *http.Request) {
	ledger, err := providers.LoadHealthLedger(s.providerHealthPath())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": s.providerHealthPath(), "ledger": ledger})
}

func (s *Server) handleProviderHealthReset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req resetProviderHealthRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	ledger, err := providers.ResetHealth(s.providerHealthPath(), req.ProviderID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": s.providerHealthPath(), "ledger": ledger})
}
