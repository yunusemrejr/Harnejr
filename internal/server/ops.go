package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/events"
	"github.com/yunusemrejr/Harnejr/internal/mcp"
	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type patchRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Path          string `json:"path"`
	OldText       string `json:"oldText"`
	NewText       string `json:"newText"`
	MaxBytes      int    `json:"maxBytes"`
}

func (s *Server) handleFilePatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req patchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "patch "+req.Path)
	if !ok { return }
	result, err := workspace.ApplyTextPatch(workspace.PatchRequest{Root: prepared.WorkspaceRoot, Path: req.Path, OldText: req.OldText, NewText: req.NewText, MaxBytes: req.MaxBytes})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	_ = events.Append(prepared.MemoryDir, events.Event{SessionID: req.SessionID, Type: "workspace.patch", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"path": result.Path, "changed": result.Changed}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "result": result})
}

func (s *Server) handleMCPCheck(w http.ResponseWriter, r *http.Request) {
	checks, err := mcp.CheckServers(filepath.Join(s.configDir, "mcp.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"checks": checks})
}
