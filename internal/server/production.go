package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/agents"
	"github.com/yunusemrejr/Harnejr/internal/events"
	"github.com/yunusemrejr/Harnejr/internal/judge"
	"github.com/yunusemrejr/Harnejr/internal/providers"
	"github.com/yunusemrejr/Harnejr/internal/shell"
	"github.com/yunusemrejr/Harnejr/internal/skills"
	"github.com/yunusemrejr/Harnejr/internal/workspace"
)

type shellRunRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Command       string `json:"command"`
	TimeoutMs     int    `json:"timeoutMs"`
	MaxOutput     int    `json:"maxOutput"`
}

func (s *Server) handleShellRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req shellRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, req.Command)
	if !ok {
		return
	}
	if prepared.MemoryDir == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "workspace did not receive Harnejr memory", "workspace": prepared})
		return
	}
	result, err := shell.Run(r.Context(), shell.RunRequest{Command: req.Command, Workspace: prepared.WorkspaceRoot, Timeout: time.Duration(req.TimeoutMs) * time.Millisecond, MaxOutput: req.MaxOutput})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	_ = events.Append(prepared.MemoryDir, events.Event{SessionID: req.SessionID, Type: "shell.run", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"command": req.Command, "decision": result.Decision.Action, "ran": result.Ran, "exitCode": result.ExitCode, "stdout": result.Stdout, "stderr": result.Stderr}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "result": result})
}

type fileRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
	Path          string `json:"path"`
	Content       string `json:"content"`
	Limit         int    `json:"limit"`
	MaxBytes      int    `json:"maxBytes"`
}

func (s *Server) handleFileList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req fileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	root := req.WorkspaceRoot
	if root == "" {
		cwd, _ := os.Getwd()
		root = cwd
	}
	result, err := workspace.List(root, req.Path, req.Limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleFileRead(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req fileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	root := req.WorkspaceRoot
	if root == "" {
		cwd, _ := os.Getwd()
		root = cwd
	}
	result, err := workspace.Read(root, req.Path, req.MaxBytes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleFileWrite(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req fileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "write "+req.Path)
	if !ok {
		return
	}
	result, err := workspace.Write(prepared.WorkspaceRoot, req.Path, req.Content, req.MaxBytes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "workspace": prepared})
		return
	}
	_ = events.Append(prepared.MemoryDir, events.Event{SessionID: req.SessionID, Type: "workspace.write", Workspace: prepared.WorkspaceRoot, Payload: map[string]any{"path": result.Path, "bytes": result.Bytes}})
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "result": result})
}

type agentPlanRequest struct {
	Task              string `json:"task"`
	Mode              string `json:"mode"`
	RequestedProvider string `json:"requestedProvider"`
	RequestedModel    string `json:"requestedModel"`
}

func (s *Server) handleAgentPlan(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req agentPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	writeJSON(w, http.StatusOK, agents.BuildPlan(agents.PlanRequest(req)))
}

func (s *Server) handleCompletionCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req judge.CompletionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	writeJSON(w, http.StatusOK, judge.Evaluate(req))
}

type providerProbeRequest struct {
	Live       bool   `json:"live"`
	ProviderID string `json:"providerId"`
}

func (s *Server) handleProviderProbe(w http.ResponseWriter, r *http.Request) {
	registry, err := providers.LoadRegistry(filepath.Join(s.configDir, "providers.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	var req providerProbeRequest
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
			return
		}
	}
	if req.Live {
		if req.ProviderID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "live provider probe requires providerId to avoid accidental quota usage"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"providers": providers.ProbeLive(r.Context(), registry, req.ProviderID)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": providers.ProbeStatic(registry)})
}

type skillsRequest struct {
	WorkspaceRoot string   `json:"workspaceRoot"`
	Roots         []string `json:"roots"`
}

func (s *Server) handleSkillsDiscover(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req skillsRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	roots := req.Roots
	if len(roots) == 0 {
		roots = skills.DefaultRoots(req.WorkspaceRoot)
	}
	report, err := skills.Scan(roots)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

type exportRequest struct {
	WorkspaceRoot string `json:"workspaceRoot"`
	SessionID     string `json:"sessionId"`
}

func (s *Server) handleSessionExport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prepared, ok := s.prepareForRequest(w, r, req.WorkspaceRoot, req.SessionID, "export session")
	if !ok {
		return
	}
	path := filepath.Join(prepared.MemoryDir, "events.jsonl")
	bytes, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace": prepared, "path": path, "jsonl": string(bytes)})
}

func (s *Server) prepareForRequest(w http.ResponseWriter, r *http.Request, root string, sessionID string, userRequest string) (workspace.PrepareResult, bool) {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return workspace.PrepareResult{}, false
		}
		root = cwd
	}
	prepared, err := workspace.PrepareSessionWorkspace(r.Context(), workspace.PrepareOptions{Root: root, SessionID: sessionID, UserRequest: userRequest, Now: time.Now().UTC()})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return workspace.PrepareResult{}, false
	}
	return prepared, true
}
