package skills

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Hash   string `json:"hash,omitempty"`
	Source string `json:"source"`
}

type Report struct {
	Roots   []string `json:"roots"`
	Entries []Entry  `json:"entries"`
}

func DefaultRoots(workspace string) []string {
	roots := []string{"~/skills", "~/.agents", "~/.codex/skills"}
	if strings.TrimSpace(workspace) != "" {
		roots = append(roots, filepath.Join(workspace, ".harnejr", "skills"))
	}
	return roots
}

func Scan(roots []string) (Report, error) {
	var report Report
	for _, root := range roots {
		expanded := expandHome(root)
		report.Roots = append(report.Roots, expanded)
		if _, err := os.Stat(expanded); err != nil {
			continue
		}
		entries, err := scanRoot(expanded)
		if err != nil {
			return Report{}, err
		}
		report.Entries = append(report.Entries, entries...)
	}
	sort.Slice(report.Entries, func(i, j int) bool { return report.Entries[i].Path < report.Entries[j].Path })
	return report, nil
}

func scanRoot(root string) ([]Entry, error) {
	var entries []Entry
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			if shouldSkip(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		name := entry.Name()
		if name != "SKILL.md" && filepath.Ext(name) != ".md" {
			return nil
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(bytes)
		kind := "agent"
		if name == "SKILL.md" || strings.Contains(strings.ToLower(path), "skill") {
			kind = "skill"
		}
		entries = append(entries, Entry{
			Name:   strings.TrimSuffix(name, filepath.Ext(name)),
			Path:   path,
			Kind:   kind,
			Hash:   fmt.Sprintf("%x", sum[:8]),
			Source: root,
		})
		return nil
	})
	return entries, err
}

func shouldSkip(name string) bool {
	switch name {
	case ".git", "node_modules", "dist", "build", ".cache":
		return true
	default:
		return false
	}
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
