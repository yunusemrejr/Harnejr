package shell

import (
	"context"
	"testing"
)

func TestRunDeniesDangerousCommand(t *testing.T) {
	result, err := Run(context.Background(), RunRequest{Command: "sudo rm -rf /", Workspace: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Ran {
		t.Fatal("dangerous command should not run")
	}
	if result.Decision.Action != "deny" {
		t.Fatalf("expected deny, got %#v", result.Decision)
	}
}

func TestRunAllowsReadOnlyCommand(t *testing.T) {
	result, err := Run(context.Background(), RunRequest{Command: "pwd", Workspace: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Ran || result.ExitCode != 0 {
		t.Fatalf("expected command to run successfully, got %#v", result)
	}
}
