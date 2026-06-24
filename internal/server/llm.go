package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type generateRequest struct {
	ProviderID         string   `json:"providerId"`
	Model              string   `json:"model"`
	System             string   `json:"system"`
	Prompt             string   `json:"prompt"`
	MaxTokens          int      `json:"maxTokens"`
	FallbackOrder      []string `json:"fallbackOrder"`
	AllowBillingChange bool     `json:"allowBillingChange"`
}

func (s *Server) handleLLMGenerate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	order := req.FallbackOrder
	if req.ProviderID != "" {
		order = append([]string{req.ProviderID}, order...)
	}
	result := providers.GenerateWithFallback(r.Context(), registry, order, providers.GenerateRequest{ProviderID: req.ProviderID, Model: req.Model, System: req.System, Prompt: req.Prompt, MaxTokens: req.MaxTokens, AllowBillingChange: req.AllowBillingChange})
	writeJSON(w, http.StatusOK, result)
}
