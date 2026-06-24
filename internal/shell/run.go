package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/policy"
)

type RunRequest struct {
	Command   string        `json:"command"`
	Workspace string        `json:"workspace"`
	Timeout   time.Duration `json:"-"`
	MaxOutput int           `json:"maxOutput"`
}

type RunResult struct {
	Decision   policy.Decision `json:"decision"`
	Ran        bool            `json:"ran"`
	ExitCode   int             `json:"exitCode"`
	Stdout     string          `json:"stdout"`
	Stderr     string          `json:"stderr"`
	TimedOut   bool            `json:"timedOut"`
	Truncated  bool            `json:"truncated"`
	DurationMs int64           `json:"durationMs"`
}

func Run(ctx context.Context, req RunRequest) (RunResult, error) {
	decision := policy.ClassifyShell(req.Command)
	result := RunResult{Decision: decision, ExitCode: -1}
	if decision.Action != policy.ActionAllow {
		return result, nil
	}
	if req.Workspace == "" {
		return result, fmt.Errorf("workspace is required")
	}
	if req.Timeout <= 0 {
		req.Timeout = 60 * time.Second
	}
	if req.MaxOutput <= 0 {
		req.MaxOutput = 64 * 1024
	}
	start := time.Now()
	runCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, "bash", "-lc", req.Command)
	cmd.Dir = req.Workspace
	var stdout cappedBuffer
	var stderr cappedBuffer
	stdout.limit = req.MaxOutput
	stderr.limit = req.MaxOutput
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result.Ran = true
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.Truncated = stdout.truncated || stderr.truncated
	result.DurationMs = time.Since(start).Milliseconds()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		return result, nil
	}
	if err == nil {
		result.ExitCode = 0
		return result, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}
	return result, err
}

type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - b.Buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		b.truncated = true
		_, _ = b.Buffer.Write(p[:remaining])
		return len(p), nil
	}
	return b.Buffer.Write(p)
}
