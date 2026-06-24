package workspace

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const maxRollbackBytes int64 = 16 * 1024 * 1024

type SnapshotResult struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Backup   string `json:"backup,omitempty"`
	Manifest string `json:"manifest"`
	Existed  bool   `json:"existed"`
	Bytes    int64  `json:"bytes"`
	Reason   string `json:"reason"`
}

func SnapshotFile(root string, target string, reason string) (SnapshotResult, error) {
	rootResolved, err := normalizeRoot(root)
	if err != nil {
		return SnapshotResult{}, err
	}
	resolved, err := ResolveInside(rootResolved, target)
	if err != nil {
		return SnapshotResult{}, err
	}
	rel := cleanRel(rootResolved, resolved)
	id := time.Now().UTC().Format("20060102T150405.000000000Z")
	snapshotRoot := filepath.Join(rootResolved, ".harnejr", "rollback", id)
	manifestPath := filepath.Join(snapshotRoot, "manifest.json")
	result := SnapshotResult{ID: id, Path: rel, Manifest: manifestPath, Reason: reason}
	if err := os.MkdirAll(snapshotRoot, 0o750); err != nil {
		return SnapshotResult{}, err
	}
	info, err := os.Stat(resolved)
	if err == nil {
		if info.IsDir() {
			return SnapshotResult{}, fmt.Errorf("cannot snapshot directory: %s", rel)
		}
		if info.Size() > maxRollbackBytes {
			return SnapshotResult{}, fmt.Errorf("refusing rollback snapshot for %s: file is larger than %d bytes", rel, maxRollbackBytes)
		}
		result.Existed = true
		result.Bytes = info.Size()
		backupPath := filepath.Join(snapshotRoot, rel+".before")
		if err := os.MkdirAll(filepath.Dir(backupPath), 0o750); err != nil {
			return SnapshotResult{}, err
		}
		if err := copyFile(resolved, backupPath); err != nil {
			return SnapshotResult{}, err
		}
		result.Backup = backupPath
	} else if !os.IsNotExist(err) {
		return SnapshotResult{}, err
	}
	if err := writeSnapshotManifest(result); err != nil {
		return SnapshotResult{}, err
	}
	return result, nil
}

func writeSnapshotManifest(result SnapshotResult) error {
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	return os.WriteFile(result.Manifest, bytes, 0o640)
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
