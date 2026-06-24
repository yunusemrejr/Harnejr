package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type sessionMessageRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	ProviderID    string `json:"providerId"`
	ModelID       string `json:"modelId"`
	Command       string `json:"command"`
	Prompt        string `json:"prompt"`
}

func (s *Server) handleSessionMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req sessionMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.Prompt) == "" && strings.TrimSpace(req.Command) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "prompt or command is required"})
		return
	}
	prepared, err := workspace.PrepareSessionWorkspace(r.Context(), workspace.PrepareOptions{
		Root:        req.WorkspaceRoot,
		SessionID:   req.SessionID,
		UserRequest: firstNonEmpty(req.Command, req.Prompt),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if prepared.MemoryDir == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "workspace cannot accept session messages", "workspace": prepared})
		return
	}
	entry := fmt.Sprintf("\n## %s\n\n- Session: `%s`\n- Provider: `%s`\n- Model: `%s`\n- Command: `%s`\n\n%s\n", time.Now().UTC().Format(time.RFC3339), cleanLog(req.SessionID), cleanLog(req.ProviderID), cleanLog(req.ModelID), cleanLog(req.Command), strings.TrimSpace(req.Prompt))
	if err := appendSessionMessage(filepath.Join(prepared.MemoryDir, "session-log.md"), entry); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accepted": true,
		"status":   "queued_for_future_provider_runtime",
		"message":  "Prompt was stored in workspace memory. Provider execution is not implemented in this scaffold yet.",
		"workspace": prepared,
	})
}

func appendSessionMessage(path string, text string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	return err
}

func cleanLog(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	if len(value) > 160 {
		return value[:160] + "..."
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
