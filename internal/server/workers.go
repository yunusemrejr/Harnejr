package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/agents"
	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type workerRunRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Task          string `json:"task"`
	Mode          string `json:"mode"`
	ProviderID    string `json:"providerId"`
	Model         string `json:"model"`
}

func (s *Server) handleWorkerRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req workerRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	plan := agents.BuildPlan(agents.PlanRequest{Task: req.Task, Mode: req.Mode, RequestedProvider: req.ProviderID, RequestedModel: req.Model})
	out := []providers.GenerateResult{}
	for _, item := range plan.Agents {
		prompt := req.Task + "\n" + item.Reason
		result := providers.GenerateWithFallback(r.Context(), registry, []string{item.ProviderID}, providers.GenerateRequest{Model: item.Model, Prompt: prompt, MaxTokens: 2048})
		out = append(out, result)
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan, "results": out})
}
