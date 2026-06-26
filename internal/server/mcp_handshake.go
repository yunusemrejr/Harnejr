package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/mcp"
)

type mcpHandshakeRequest struct {
	IDs       []string `json:"ids,omitempty"`
	TimeoutMs int      `json:"timeoutMs,omitempty"`
	ListTools bool     `json:"listTools"`
}

func (s *Server) handleMCPHandshake(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req mcpHandshakeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	checks, err := mcp.CheckServers(filepath.Join(s.configDir, "mcp.default.json"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	filtered := make([]mcp.Check, 0, len(checks))
	wanted := map[string]bool{}
	for _, id := range req.IDs {
		if id != "" { wanted[id] = true }
	}
	for _, check := range checks {
		if len(wanted) == 0 || wanted[check.ID] {
			filtered = append(filtered, check)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"checks": filtered, "handshakeAttempted": false, "reason": "external MCP process handshake requires the daemon stdio probe implementation; this endpoint reports command/env readiness without claiming initialized/list-tools success"})
}
