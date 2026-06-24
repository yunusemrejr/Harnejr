package quality

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLoCFlagsOversizedSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.go")
	content := ""
	for i := 0; i < 5; i++ {
		content += "package main\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := ScanLoC(LoCOptions{Root: root, MaxLines: 3})
	if err != nil {
		t.Fatal(err)
	}
	if report.Completion != "review_required" {
		t.Fatalf("expected review_required, got %s", report.Completion)
	}
	if len(report.Oversized) != 1 || report.Oversized[0].Path != "large.go" {
		t.Fatalf("unexpected oversized files: %#v", report.Oversized)
	}
}

func TestScanLoCSkipsGeneratedFolders(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "node_modules", "ignored.js"), []byte("a\nb\nc\nd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := ScanLoC(LoCOptions{Root: root, MaxLines: 1})
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanned != 0 {
		t.Fatalf("expected generated folder to be skipped, scanned %d", report.Scanned)
	}
}
