package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type prepareWorkspaceRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	UserRequest   string `json:"userRequest"`
}

func (s *Server) handlePrepareWorkspace(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req prepareWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	root := req.WorkspaceRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		root = cwd
	}
	result, err := workspace.PrepareSessionWorkspace(r.Context(), workspace.PrepareOptions{
		Root:        root,
		SessionID:   req.SessionID,
		UserRequest: req.UserRequest,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		s.logger.Warn("workspace preparation failed", "error", err, "root", root)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
