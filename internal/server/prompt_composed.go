package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/yunusemrejr/Harnejr/internal/prompts"
)

func (s *Server) handleGetComposedPrompt(w http.ResponseWriter, r *http.Request) {
	corePath := filepath.Join(s.configDir, "harness.prompt.md")
	coreBytes, err := os.ReadFile(corePath)
	if os.IsNotExist(err) {
		coreBytes = []byte("Harnejr core rules remain active.")
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	user, err := prompts.LoadUserPrompt(s.configDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"content": prompts.ComposeSystemPrompt(string(coreBytes), user), "corePath": corePath, "userPath": user.Path})
}
