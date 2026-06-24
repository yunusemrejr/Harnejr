package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInsideAllowsWorkspacePath(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveInside(root, "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(resolved) != "file.txt" {
		t.Fatalf("unexpected resolved path: %s", resolved)
	}
}

func TestResolveInsideRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "..", "outside.txt")
	_, err := ResolveInside(root, outside)
	if err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func TestResolveInsideRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := ResolveInside(root, "escape/file.txt")
	if err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}
