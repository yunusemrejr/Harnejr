package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/yunusemrejr/Harnejr/internal/agents"
	"github.com/yunusemrejr/Harnejr/internal/events"
	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type workerRunRequest struct {
	WorkspaceRoot      string `json:"workspaceRoot"`
	SessionID          string `json:"sessionId"`
	Task               string `json:"task"`
	Mode               string `json:"mode"`
	ProviderID         string `json:"providerId"`
	Model              string `json:"model"`
	MaxConcurrency     int    `json:"maxConcurrency"`
	AllowBillingChange bool   `json:"allowBillingChange"`
}

func (s *Server) handleWorkerRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req workerRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, req.Task)
	if !ok {
		return
	}
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	plan := agents.BuildPlan(agents.PlanRequest{Task: req.Task, Mode: req.Mode, RequestedProvider: req.ProviderID, RequestedModel: req.Model})
	concurrency := req.MaxConcurrency
	if concurrency <= 0 || concurrency > 3 {
		concurrency = 2
	}
	out := make([]providers.GenerateResult, len(plan.Agents))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for index, item := range plan.Agents {
		wg.Add(1)
		go func(i int, providerID string, model string, reason string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			prompt := req.Task + "\n" + reason
			out[i] = providers.GenerateWithFallback(r.Context(), registry, []string{providerID}, providers.GenerateRequest{Model: model, Prompt: prompt, MaxTokens: 2048, AllowBillingChange: req.AllowBillingChange})
		}(index, item.ProviderID, item.Model, item.Reason)
	}
	wg.Wait()
	_ = events.Append(prepared.MemoryDir, events.Event{SessionID: req.SessionID, Type: "workers.run", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"task": req.Task, "agentCount": len(plan.Agents), "resultCount": len(out)}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "plan": plan, "results": out})
}
