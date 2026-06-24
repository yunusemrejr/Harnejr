package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/config"
	"github.com/yunusemrejr/Harnejr/internal/policy"
)

type Options struct {
	Listen    string
	ConfigDir string
	Logger    *slog.Logger
}

type Server struct {
	httpServer *http.Server
	configDir  string
	logger     *slog.Logger
}

func New(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	s := &Server{
		configDir: opts.ConfigDir,
		logger:    logger,
	}

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/config/defaults", s.handleConfigDefaults)
	mux.HandleFunc("POST /api/policy/classify-shell", s.handleClassifyShell)
	mux.HandleFunc("POST /api/workspaces/prepare", s.handlePrepareWorkspace)

	s.httpServer = &http.Server{
		Addr:              opts.Listen,
		Handler:           withSecurityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Harnejr</title>
  <style>
    body { margin: 0; background: #242424; color: #f2ead7; font-family: Arial, Helvetica, sans-serif; }
    main { max-width: 920px; margin: 0 auto; padding: 48px 24px; }
    section { border: 1px solid #6f6a5e; background: #303030; border-radius: 18px; padding: 24px; }
    h1 { font-size: 44px; margin: 0 0 12px; font-weight: 700; }
    p { color: #d8d0bd; line-height: 1.55; }
    code { background: #1d1d1d; color: #f2ead7; padding: 2px 6px; border-radius: 6px; }
  </style>
</head>
<body>
  <main>
    <section>
      <h1>Harnejr</h1>
      <p>The local daemon is running. The full web GUI scaffold lives under <code>apps/web</code>.</p>
      <p>API: <code>/api/health</code>, <code>/api/config/defaults</code>, <code>/api/policy/classify-shell</code>, <code>/api/workspaces/prepare</code>.</p>
    </section>
  </main>
</body>
</html>`))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"name":      "harnejrd",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
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
