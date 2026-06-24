package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/server"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8765", "HTTP listen address")
	configDir := flag.String("config-dir", "configs", "directory containing Harnejr config JSON files")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	srv := server.New(server.Options{
		Listen:    *listen,
		ConfigDir: *configDir,
		Logger:    logger,
	})

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting harnejr daemon", "listen", *listen, "configDir", *configDir)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown requested", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "harnejrd failed: %v\n", err)
			os.Exit(1)
		}
	}
}
