package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrepareSessionWorkspaceUsesExistingGitRootAndMemory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "src", "app")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := PrepareSessionWorkspace(context.Background(), PrepareOptions{
		Root:        sub,
		SessionID:   "s-1",
		UserRequest: "build workspace memory",
		Now:         time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.GitStatus != GitExisting {
		t.Fatalf("expected existing git status, got %#v", result)
	}
	if result.GitRoot != root {
		t.Fatalf("expected git root %s, got %s", root, result.GitRoot)
	}
	if result.MemoryDir != filepath.Join(root, ".harnejr") {
		t.Fatalf("memory should live at repo root, got %s", result.MemoryDir)
	}
	logBytes, err := os.ReadFile(filepath.Join(root, ".harnejr", "session-log.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBytes), "build workspace memory") {
		t.Fatalf("session log does not include request: %s", string(logBytes))
	}
}

func TestPrepareSessionWorkspaceSkipsBroadUserFolders(t *testing.T) {
	home := t.TempDir()
	desktop := filepath.Join(home, "Desktop")
	if err := os.Mkdir(desktop, 0o755); err != nil {
		t.Fatal(err)
	}
	broad, reason := isBroadWorkspaceRootWithHome(desktop, home)
	if !broad {
		t.Fatalf("expected Desktop to be broad")
	}
	if !strings.Contains(reason, "Desktop") {
		t.Fatalf("expected reason to mention Desktop, got %q", reason)
	}
}

func TestPrepareSessionWorkspaceSkipsParentWithNestedRepos(t *testing.T) {
	root := t.TempDir()
	childGit := filepath.Join(root, "child", ".git")
	if err := os.MkdirAll(childGit, 0o700); err != nil {
		t.Fatal(err)
	}

	result, err := PrepareSessionWorkspace(context.Background(), PrepareOptions{Root: root, SessionID: "s-2"})
	if err != nil {
		t.Fatal(err)
	}
	if result.GitStatus != GitSkippedNested {
		t.Fatalf("expected nested skip, got %#v", result)
	}
	if len(result.NestedGitRoots) != 1 || result.NestedGitRoots[0] != filepath.Join(root, "child") {
		t.Fatalf("unexpected nested roots: %#v", result.NestedGitRoots)
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); !os.IsNotExist(err) {
		t.Fatalf("parent git repository should not have been created")
	}
	if _, err := os.Stat(filepath.Join(root, ".harnejr")); !os.IsNotExist(err) {
		t.Fatalf("parent memory should not be created when nested repos are present")
	}
}
