package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyTextPatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ApplyTextPatch(PatchRequest{Root: root, Path: "a.txt", OldText: "world", NewText: "harnejr"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.Replacements != 1 {
		t.Fatalf("unexpected patch result: %#v", result)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != "hello harnejr" {
		t.Fatalf("unexpected file content: %q", string(bytes))
	}
}

func TestApplyTextPatchRejectsSensitivePath(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("A=B"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyTextPatch(PatchRequest{Root: root, Path: ".env", OldText: "A", NewText: "C"}); err == nil {
		t.Fatal("expected sensitive patch rejection")
	}
}
