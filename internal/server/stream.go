package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type streamRequest struct {
	ProviderID          string `json:"providerId"`
	Model               string `json:"model"`
	System              string `json:"system"`
	Prompt              string `json:"prompt"`
	MaxTokens           int    `json:"maxTokens"`
	CacheMode           string `json:"cacheMode,omitempty"`
	CacheStablePrefix   string `json:"cacheStablePrefix,omitempty"`
	CacheDynamicContext string `json:"cacheDynamicContext,omitempty"`
}

func (s *Server) handleLLMStream(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req streamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if req.ProviderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "providerId is required for streaming to avoid accidental quota usage"})
		return
	}
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
	provider, ok := providers.FindProvider(registry, req.ProviderID)
	if !ok { writeJSON(w, http.StatusBadRequest, map[string]any{"error": "provider not found"}); return }
	w.Header().Set("content-type", "text/event-stream")
	w.Header().Set("cache-control", "no-cache")
	flusher, _ := w.(http.Flusher)
	emit := func(event providers.StreamEvent) error {
		payload, err := json.Marshal(event)
		if err != nil { return err }
		_, err = fmt.Fprintf(w, "data: %s\n\n", payload)
		if flusher != nil { flusher.Flush() }
		return err
	}
	providers.Stream(r.Context(), provider, providers.GenerateRequest{Model: req.Model, System: req.System, Prompt: req.Prompt, MaxTokens: req.MaxTokens, CacheMode: req.CacheMode, CacheStablePrefix: req.CacheStablePrefix, CacheDynamicContext: req.CacheDynamicContext}, emit)
}
