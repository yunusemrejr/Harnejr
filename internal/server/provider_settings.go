package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yunusemrejr/Harnejr/internal/providers"
)

type providerRegistryResponse struct {
	Registry providers.Registry `json:"registry"`
	Secrets  map[string]bool    `json:"secrets"`
}

func (s *Server) handleProviderRegistry(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.configDir, "providers.default.json")
	switch r.Method {
	case http.MethodGet:
		registry, err := providers.LoadRegistry(path)
		if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
		writeJSON(w, http.StatusOK, providerRegistryResponse{Registry: registry, Secrets: secretMap(registry)})
	case http.MethodPut:
		defer r.Body.Close()
		var registry providers.Registry
		if err := json.NewDecoder(r.Body).Decode(&registry); err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"}); return }
		if issues := providers.ValidateRegistry(registry); len(issues) > 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error": issues[0].Message, "issues": issues}); return }
		if err := providers.SaveRegistry(path, registry); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
		writeJSON(w, http.StatusOK, providerRegistryResponse{Registry: registry, Secrets: secretMap(registry)})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type providerSecretRequest struct {
	ProviderID string `json:"providerId"`
	APIKey     string `json:"apiKey"`
}

func (s *Server) handleProviderSecret(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req providerSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"}); return }
	req.ProviderID = strings.TrimSpace(req.ProviderID)
	if req.ProviderID == "" || strings.TrimSpace(req.APIKey) == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error": "providerId and apiKey are required"}); return }
	registryPath := filepath.Join(s.configDir, "providers.default.json")
	registry, err := providers.LoadRegistry(registryPath)
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
	index := -1
	for i, provider := range registry.Providers { if provider.ID == req.ProviderID { index = i; break } }
	if index < 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error": "provider not found"}); return }
	secretDir := filepath.Join(s.configDir, "secrets")
	if err := os.MkdirAll(secretDir, 0o700); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
	secretPath := filepath.Join(secretDir, req.ProviderID+".key")
	if err := os.WriteFile(secretPath, []byte(strings.TrimSpace(req.APIKey)+"\n"), 0o600); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
	registry.Providers[index].APIKeySecretRef = secretPath
	if err := providers.SaveRegistry(registryPath, registry); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()}); return }
	writeJSON(w, http.StatusOK, map[string]any{"providerId": req.ProviderID, "secretSaved": true, "path": secretPath})
}

func secretMap(registry providers.Registry) map[string]bool {
	out := map[string]bool{}
	for _, provider := range registry.Providers {
		ok := false
		if provider.APIKeySecretRef != "" {
			if bytes, err := os.ReadFile(provider.APIKeySecretRef); err == nil && strings.TrimSpace(string(bytes)) != "" { ok = true }
		}
		out[provider.ID] = ok
	}
	return out
}
