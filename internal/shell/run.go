package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
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
	Sandbox    string          `json:"sandbox"`
}

func Run(ctx context.Context, req RunRequest) (RunResult, error) {
	decision := policy.ClassifyShell(req.Command)
	result := RunResult{Decision: decision, ExitCode: -1, Sandbox: "not-run"}
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
	cmd, mode := commandFor(runCtx, req)
	result.Sandbox = mode
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

func commandFor(ctx context.Context, req RunRequest) (*exec.Cmd, string) {
	if bubblewrapUsable() {
		args := []string{"--die-with-parent", "--unshare-all", "--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp", "--bind", req.Workspace, req.Workspace, "--chdir", req.Workspace}
		for _, path := range []string{"/usr", "/bin", "/lib", "/lib64", "/etc"} {
			if _, err := os.Stat(path); err == nil {
				args = append(args, "--ro-bind", path, path)
			}
		}
		args = append(args, "bash", "-lc", req.Command)
		return exec.CommandContext(ctx, "bwrap", args...), "bubblewrap"
	}
	cmd := exec.CommandContext(ctx, "bash", "-lc", req.Command)
	cmd.Dir = req.Workspace
	return cmd, "unsandboxed-no-usable-bwrap"
}

var bwrapOnce sync.Once
var bwrapOK bool

func bubblewrapUsable() bool {
	bwrapOnce.Do(func() {
		if _, err := exec.LookPath("bwrap"); err != nil {
			bwrapOK = false
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "bwrap", "--ro-bind", "/usr", "/usr", "true")
		bwrapOK = cmd.Run() == nil
	})
	return bwrapOK
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
