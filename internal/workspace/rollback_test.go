package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotFileBacksUpExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := SnapshotFile(root, "a.txt", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Existed || snapshot.Backup == "" {
		t.Fatalf("expected existing backup, got %#v", snapshot)
	}
	bytes, err := os.ReadFile(snapshot.Backup)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != "before" {
		t.Fatalf("unexpected backup content: %q", string(bytes))
	}
}

func TestChangeFileWithBackup(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ChangeFileWithBackup(root, "b.txt", "new", 0, "test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Backup.Backup == "" || result.Result.Path != "b.txt" {
		t.Fatalf("unexpected result: %#v", result)
	}
	bytes, err := os.ReadFile(filepath.Join(root, "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != "new" {
		t.Fatalf("unexpected changed content: %q", string(bytes))
	}
}
