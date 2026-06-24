package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/judge"
	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type reviewRunRequest struct {
	Input      judge.CompletionInput `json:"input"`
	ProviderID string                `json:"providerId"`
	Model      string                `json:"model"`
}

func (s *Server) handleReviewRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req reviewRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	decision := judge.Evaluate(req.Input)
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	providerID := req.ProviderID
	if providerID == "" { providerID = "nvidia-build-nim" }
	prompt := "Goal: " + req.Input.Goal + "\nAssess whether the evidence supports completion."
	result := providers.GenerateWithFallback(r.Context(), registry, []string{providerID, "stepfun-step-plan", "minimax-token-plan"}, providers.GenerateRequest{Model: req.Model, Prompt: prompt, MaxTokens: 2048})
	writeJSON(w, http.StatusOK, map[string]any{"decision": decision, "review": result})
}
