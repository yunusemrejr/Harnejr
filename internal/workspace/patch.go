package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PatchRequest struct {
	Root     string `json:"root"`
	Path     string `json:"path"`
	OldText  string `json:"oldText"`
	NewText  string `json:"newText"`
	MaxBytes int    `json:"maxBytes"`
}

type PatchResult struct {
	Path         string          `json:"path"`
	Changed      bool            `json:"changed"`
	Replacements int             `json:"replacements"`
	Bytes        int             `json:"bytes"`
	Snapshot     *SnapshotResult `json:"snapshot,omitempty"`
}

func ApplyTextPatch(req PatchRequest) (PatchResult, error) {
	if req.MaxBytes <= 0 {
		req.MaxBytes = 1024 * 1024
	}
	if req.OldText == "" {
		return PatchResult{}, fmt.Errorf("oldText is required")
	}
	resolved, err := ResolveInside(req.Root, req.Path)
	if err != nil {
		return PatchResult{}, err
	}
	if IsSensitivePath(resolved) {
		return PatchResult{}, fmt.Errorf("refusing to patch likely secret file: %s", req.Path)
	}
	unlock, err := acquireLock(req.Root, req.Path)
	if err != nil {
		return PatchResult{}, err
	}
	defer unlock()
	bytes, err := os.ReadFile(resolved)
	if err != nil {
		return PatchResult{}, err
	}
	if len(bytes) > req.MaxBytes {
		return PatchResult{}, fmt.Errorf("file exceeds max patch size")
	}
	content := string(bytes)
	count := strings.Count(content, req.OldText)
	if count == 0 {
		return PatchResult{}, fmt.Errorf("oldText not found")
	}
	snapshot, err := SnapshotFile(req.Root, req.Path, "workspace.patch")
	if err != nil {
		return PatchResult{}, err
	}
	updated := strings.Replace(content, req.OldText, req.NewText, 1)
	if err := os.WriteFile(resolved, []byte(updated), 0o644); err != nil {
		return PatchResult{}, err
	}
	rootResolved, err := normalizeRoot(req.Root)
	if err != nil {
		return PatchResult{}, err
	}
	return PatchResult{Path: cleanRel(rootResolved, resolved), Changed: updated != content, Replacements: 1, Bytes: len(updated), Snapshot: &snapshot}, nil
}

func acquireLock(root string, target string) (func(), error) {
	lockDir := filepath.Join(root, ".harnejr", "locks")
	if err := os.MkdirAll(lockDir, 0o750); err != nil {
		return nil, err
	}
	name := strings.NewReplacer("/", "_", "\\", "_", "..", "_").Replace(target) + ".lock"
	lockPath := filepath.Join(lockDir, name)
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("file is locked: %s", target)
	}
	_, _ = file.WriteString(time.Now().UTC().Format(time.RFC3339Nano))
	_ = file.Close()
	return func() { _ = os.Remove(lockPath) }, nil
}
