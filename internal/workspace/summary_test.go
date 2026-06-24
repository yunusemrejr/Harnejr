package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSummaryWritesFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "session-log.md"), []byte("first\nsecond\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := BuildSummary(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bytes), "second") {
		t.Fatalf("unexpected summary: %s", string(bytes))
	}
}
