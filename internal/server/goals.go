package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/events"
	"github.com/yunusemrejr/Harnejr/internal/goals"
	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type goalStartRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Goal          string `json:"goal"`
}

type goalMarkRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	CheckpointID  string `json:"checkpointId"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type memorySummaryRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
}

func (s *Server) handleGoalStart(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req goalStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, req.Goal)
	if !ok {
		return
	}
	state, err := goals.Start(prepared.MemoryDir, req.SessionID, req.Goal)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	_ = events.Append(prepared.MemoryDir, events.Event{Timestamp: time.Now().UTC(), SessionID: req.SessionID, Type: "goal.start", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"goal": req.Goal, "checkpoints": len(state.Checkpoints)}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "goal": state})
}

func (s *Server) handleGoalStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req memorySummaryRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "goal status")
	if !ok {
		return
	}
	state, err := goals.Load(prepared.MemoryDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "goal": state})
}

func (s *Server) handleGoalCheckpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req goalMarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "goal checkpoint")
	if !ok {
		return
	}
	status := goals.CheckpointStatus(req.Status)
	if status == "" {
		status = goals.CheckpointDone
	}
	state, err := goals.Mark(prepared.MemoryDir, req.CheckpointID, status, req.Notes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	_ = events.Append(prepared.MemoryDir, events.Event{Timestamp: time.Now().UTC(), SessionID: req.SessionID, Type: "goal.checkpoint", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"checkpoint": req.CheckpointID, "status": status, "notes": req.Notes}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "goal": state})
}

func (s *Server) handleMemorySummary(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req memorySummaryRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "memory summary")
	if !ok {
		return
	}
	result, err := workspace.BuildSummary(prepared.MemoryDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "summary": result})
}
