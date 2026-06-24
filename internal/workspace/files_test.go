package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteInsideWorkspace(t *testing.T) {
	root := t.TempDir()
	written, err := Write(root, "notes/output.md", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	if written.Path != "notes/output.md" {
		t.Fatalf("unexpected write path: %#v", written)
	}
	read, err := Read(root, "notes/output.md", 0)
	if err != nil {
		t.Fatal(err)
	}
	if read.Content != "hello" {
		t.Fatalf("unexpected content: %q", read.Content)
	}
}

func TestWriteRejectsSensitivePath(t *testing.T) {
	root := t.TempDir()
	if _, err := Write(root, ".env", "SECRET=1", 0); err == nil {
		t.Fatal("expected sensitive path rejection")
	}
}

func TestReadRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Read(root, "escape/secret.txt", 0); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}
