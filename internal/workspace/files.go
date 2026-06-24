package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Path  string `json:"path"`
	Kind  string `json:"kind"`
	Bytes int64  `json:"bytes,omitempty"`
}

type ListResult struct {
	Root      string      `json:"root"`
	Path      string      `json:"path"`
	Entries   []FileEntry `json:"entries"`
	Truncated bool        `json:"truncated"`
}

type ReadResult struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
}

type WriteResult struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

func List(root string, target string, limit int) (ListResult, error) {
	if limit <= 0 {
		limit = 200
	}
	resolved, err := ResolveInside(root, defaultTarget(target))
	if err != nil {
		return ListResult{}, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return ListResult{}, err
	}
	if !info.IsDir() {
		return ListResult{}, fmt.Errorf("not a directory: %s", target)
	}
	rootResolved, err := normalizeRoot(root)
	if err != nil {
		return ListResult{}, err
	}
	var result ListResult
	result.Root = rootResolved
	result.Path = cleanRel(rootResolved, resolved)
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return ListResult{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if len(result.Entries) >= limit {
			result.Truncated = true
			break
		}
		if entry.Name() == ".git" {
			continue
		}
		path := filepath.Join(resolved, entry.Name())
		entryInfo, err := entry.Info()
		if err != nil {
			return ListResult{}, err
		}
		kind := "file"
		if entry.IsDir() {
			kind = "dir"
		} else if entryInfo.Mode()&fs.ModeSymlink != 0 {
			kind = "symlink"
		}
		result.Entries = append(result.Entries, FileEntry{Path: cleanRel(rootResolved, path), Kind: kind, Bytes: entryInfo.Size()})
	}
	return result, nil
}

func Read(root string, target string, maxBytes int) (ReadResult, error) {
	if maxBytes <= 0 {
		maxBytes = 256 * 1024
	}
	resolved, err := ResolveInside(root, target)
	if err != nil {
		return ReadResult{}, err
	}
	if IsSensitivePath(resolved) {
		return ReadResult{}, fmt.Errorf("refusing to read likely secret file: %s", target)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return ReadResult{}, err
	}
	if info.IsDir() {
		return ReadResult{}, fmt.Errorf("cannot read directory: %s", target)
	}
	file, err := os.Open(resolved)
	if err != nil {
		return ReadResult{}, err
	}
	defer file.Close()
	buffer := make([]byte, maxBytes+1)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return ReadResult{}, err
	}
	truncated := n > maxBytes
	if truncated {
		n = maxBytes
	}
	rootResolved, err := normalizeRoot(root)
	if err != nil {
		return ReadResult{}, err
	}
	return ReadResult{Path: cleanRel(rootResolved, resolved), Content: string(buffer[:n]), Bytes: n, Truncated: truncated}, nil
}

func Write(root string, target string, content string, maxBytes int) (WriteResult, error) {
	if maxBytes <= 0 {
		maxBytes = 1024 * 1024
	}
	if len(content) > maxBytes {
		return WriteResult{}, fmt.Errorf("content exceeds max write size")
	}
	resolved, err := ResolveInside(root, target)
	if err != nil {
		return WriteResult{}, err
	}
	if IsSensitivePath(resolved) {
		return WriteResult{}, fmt.Errorf("refusing to write likely secret file: %s", target)
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return WriteResult{}, err
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return WriteResult{}, err
	}
	rootResolved, err := normalizeRoot(root)
	if err != nil {
		return WriteResult{}, err
	}
	return WriteResult{Path: cleanRel(rootResolved, resolved), Bytes: len(content)}, nil
}

func IsSensitivePath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	base := strings.ToLower(filepath.Base(path))
	if strings.Contains(lower, "/.ssh/") || strings.Contains(lower, "/.aws/") || strings.Contains(lower, "/.gnupg/") {
		return true
	}
	if strings.HasPrefix(base, ".env") || strings.Contains(base, "secret") || strings.Contains(base, "credential") || strings.Contains(base, "token") {
		return true
	}
	return base == "id_rsa" || base == "id_ed25519" || base == "known_hosts"
}

func defaultTarget(target string) string {
	if strings.TrimSpace(target) == "" {
		return "."
	}
	return target
}

func cleanRel(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}
