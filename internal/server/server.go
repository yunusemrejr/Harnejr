package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/config"
	"github.com/yunusemrejr/Harnejr/internal/policy"
)

type Options struct {
	Listen    string
	ConfigDir string
	WebDir    string
	Logger    *slog.Logger
}

type Server struct {
	httpServer *http.Server
	configDir  string
	webDir     string
	logger     *slog.Logger
}

func New(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	s := &Server{configDir: opts.ConfigDir, webDir: opts.WebDir, logger: logger}

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/config/defaults", s.handleConfigDefaults)
	mux.HandleFunc("GET /api/doctor", s.handleDoctor)
	mux.HandleFunc("GET /api/tools", s.handleTools)
	mux.HandleFunc("GET /api/mcp/systems", s.handleMCPSystems)
	mux.HandleFunc("GET /api/mcp/check", s.handleMCPCheck)
	mux.HandleFunc("GET /api/prompts/user", s.handleGetUserPrompt)
	mux.HandleFunc("GET /api/prompts/composed", s.handleGetComposedPrompt)
	mux.HandleFunc("GET /api/providers/probe", s.handleProviderProbe)
	mux.HandleFunc("GET /api/providers/registry", s.handleProviderRegistry)
	mux.HandleFunc("PUT /api/providers/registry", s.handleProviderRegistry)
	mux.HandleFunc("PUT /api/providers/secret", s.handleProviderSecret)
	mux.HandleFunc("POST /api/providers/probe", s.handleProviderProbe)
	mux.HandleFunc("POST /api/providers/route", s.handleProviderRoute)
	mux.HandleFunc("PUT /api/prompts/user", s.handleSaveUserPrompt)
	mux.HandleFunc("POST /api/session/message", s.handleSessionMessage)
	mux.HandleFunc("POST /api/session/export", s.handleSessionExport)
	mux.HandleFunc("POST /api/memory/summary", s.handleMemorySummary)
	mux.HandleFunc("POST /api/goals/start", s.handleGoalStart)
	mux.HandleFunc("POST /api/goals/status", s.handleGoalStatus)
	mux.HandleFunc("POST /api/goals/checkpoint", s.handleGoalCheckpoint)
	mux.HandleFunc("POST /api/llm/generate", s.handleLLMGenerate)
	mux.HandleFunc("POST /api/llm/stream", s.handleLLMStream)
	mux.HandleFunc("POST /api/workers/run", s.handleWorkerRun)
	mux.HandleFunc("POST /api/review/run", s.handleReviewRun)
	mux.HandleFunc("POST /api/policy/classify-shell", s.handleClassifyShell)
	mux.HandleFunc("POST /api/shell/run", s.handleShellRun)
	mux.HandleFunc("POST /api/workspaces/prepare", s.handlePrepareWorkspace)
	mux.HandleFunc("POST /api/workspace/files/list", s.handleFileList)
	mux.HandleFunc("POST /api/workspace/files/read", s.handleFileRead)
	mux.HandleFunc("POST /api/workspace/files/write", s.handleFileWrite)
	mux.HandleFunc("POST /api/workspace/files/patch", s.handleFilePatch)
	mux.HandleFunc("POST /api/control/apply", s.handleApplyControl)
	mux.HandleFunc("POST /api/quality/loc", s.handleLoCScan)
	mux.HandleFunc("POST /api/healing/plan", s.handleHealingPlan)
	mux.HandleFunc("POST /api/agents/plan", s.handleAgentPlan)
	mux.HandleFunc("POST /api/completion/check", s.handleCompletionCheck)
	mux.HandleFunc("POST /api/skills/discover", s.handleSkillsDiscover)

	s.httpServer = &http.Server{Addr: opts.Listen, Handler: withSecurityHeaders(mux), ReadHeaderTimeout: 5 * time.Second}
	return s
}

func (s *Server) ListenAndServe() error { return s.httpServer.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if s.serveBuiltWeb(w, r) {
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8" /><meta name="viewport" content="width=device-width, initial-scale=1" /><title>Harnejr</title><style>body{margin:0;background:#242424;color:#f2ead7;font-family:Arial,Helvetica,sans-serif}main{max-width:920px;margin:0 auto;padding:48px 24px}section{border:1px solid #6f6a5e;background:#303030;border-radius:18px;padding:24px}h1{font-size:44px;margin:0 0 12px;font-weight:700}p{color:#d8d0bd;line-height:1.55}code{background:#1d1d1d;color:#f2ead7;padding:2px 6px;border-radius:6px}</style></head><body><main><section><h1>Harnejr daemon</h1><p>The daemon is running, but built web assets were not found. Run <code>pnpm install</code>, <code>pnpm build</code>, then start the daemon with <code>--web-dir apps/web/dist</code>, or reinstall with <code>bash install.sh</code>.</p></section></main></body></html>`))
}

func (s *Server) serveBuiltWeb(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(s.webDir) == "" {
		return false
	}
	root, err := filepath.Abs(s.webDir)
	if err != nil {
		return false
	}
	index := filepath.Join(root, "index.html")
	if _, err := os.Stat(index); err != nil {
		return false
	}
	rel := strings.TrimPrefix(r.URL.Path, "/")
	if rel == "" || strings.HasSuffix(r.URL.Path, "/") {
		http.ServeFile(w, r, index)
		return true
	}
	candidate := filepath.Join(root, filepath.Clean(rel))
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	if absCandidate != root && !strings.HasPrefix(absCandidate, root+string(os.PathSeparator)) {
		http.Error(w, "invalid asset path", http.StatusBadRequest)
		return true
	}
	if info, err := os.Stat(absCandidate); err == nil && !info.IsDir() {
		http.ServeFile(w, r, absCandidate)
		return true
	}
	http.ServeFile(w, r, index)
	return true
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "name": "harnejrd", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) handleConfigDefaults(w http.ResponseWriter, r *http.Request) {
	defaults, err := config.LoadDefaults(s.configDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, defaults)
}

type classifyShellRequest struct {
	Command string `json:"command"`
}

func (s *Server) handleClassifyShell(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req classifyShellRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	writeJSON(w, http.StatusOK, policy.ClassifyShell(req.Command))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("referrer-policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
