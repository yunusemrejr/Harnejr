package server

import (
	"encoding/json"
	"net/http"

	"github.com/yunusemrejr/Harnejr/internal/doctor"
	"github.com/yunusemrejr/Harnejr/internal/healing"
	"github.com/yunusemrejr/Harnejr/internal/mcp"
	"github.com/yunusemrejr/Harnejr/internal/prompts"
	"github.com/yunusemrejr/Harnejr/internal/quality"
	"github.com/yunusemrejr/Harnejr/internal/tools"
)

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, doctor.Run(s.configDir))
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"tools": tools.Builtins()})
}

func (s *Server) handleMCPSystems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"systems": mcp.BuiltinSystems()})
}

type locScanRequest struct {
	Root     string `json:"root"`
	MaxLines int    `json:"maxLines"`
}

func (s *Server) handleLoCScan(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req locScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	report, err := quality.ScanLoC(quality.LoCOptions{Root: req.Root, MaxLines: req.MaxLines})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

type healingPlanRequest struct {
	Root     string `json:"root"`
	MaxLines int    `json:"maxLines"`
}

func (s *Server) handleHealingPlan(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req healingPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	loc, err := quality.ScanLoC(quality.LoCOptions{Root: req.Root, MaxLines: req.MaxLines})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, healing.BuildPlan(doctor.Run(s.configDir), loc))
}

func (s *Server) handleGetUserPrompt(w http.ResponseWriter, r *http.Request) {
	prompt, err := prompts.LoadUserPrompt(s.configDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, prompt)
}

type userPromptRequest struct {
	Content string `json:"content"`
}

func (s *Server) handleSaveUserPrompt(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req userPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	prompt, err := prompts.SaveUserPrompt(s.configDir, req.Content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, prompt)
}
