package server

import (
	"encoding/json"
	"net/http"
	"time"

	hsession "github.com/yunusemrejr/Harnejr/internal/session"
	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type controlRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Mode          string `json:"mode"`
	Topic         string `json:"topic"`
	Goal          string `json:"goal"`
	LoopTask      string `json:"loopTask"`
	LoopLimit     int    `json:"loopLimit"`
	Yolo          *bool  `json:"yolo"`
	ClearGoal     bool   `json:"clearGoal"`
	ClearTopic    bool   `json:"clearTopic"`
}

func (s *Server) handleApplyControl(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req controlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, err := workspace.PrepareSessionWorkspace(r.Context(), workspace.PrepareOptions{
		Root:      req.WorkspaceRoot,
		SessionID: req.SessionID,
		Now:       time.Now().UTC(),
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if prepared.MemoryDir == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "workspace did not receive Harnejr memory; control state cannot be stored for this location", "workspace": prepared})
		return
	}
	state, err := hsession.ApplyControl(prepared.MemoryDir, hsession.ControlPatch{
		SessionID: req.SessionID,
		Mode:      hsession.Mode(req.Mode),
		Topic:     req.Topic,
		Goal:      req.Goal,
		LoopTask:  req.LoopTask,
		LoopLimit: req.LoopLimit,
		Yolo:      req.Yolo,
		ClearGoal: req.ClearGoal,
		ClearTopic: req.ClearTopic,
	}, time.Now().UTC())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "control": state})
}
