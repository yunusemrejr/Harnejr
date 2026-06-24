package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveInside(root string, target string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	if strings.TrimSpace(target) == "" {
		return "", fmt.Errorf("target path is required")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	absTarget := target
	if !filepath.IsAbs(absTarget) {
		absTarget = filepath.Join(absRoot, target)
	}
	absTarget, err = filepath.Abs(absTarget)
	if err != nil {
		return "", fmt.Errorf("resolve target: %w", err)
	}

	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", fmt.Errorf("real root: %w", err)
	}

	realTarget := absTarget
	if existing, err := existingPathForSymlinkCheck(absTarget); err == nil {
		realTarget, err = filepath.EvalSymlinks(existing)
		if err != nil {
			return "", fmt.Errorf("real target: %w", err)
		}
		if existing != absTarget {
			rel, err := filepath.Rel(existing, absTarget)
			if err != nil {
				return "", fmt.Errorf("target rel: %w", err)
			}
			realTarget = filepath.Join(realTarget, rel)
		}
	}

	if realTarget == realRoot || strings.HasPrefix(realTarget, realRoot+string(os.PathSeparator)) {
		return realTarget, nil
	}
	return "", fmt.Errorf("path escapes workspace: %s", target)
}

func existingPathForSymlinkCheck(path string) (string, error) {
	current := path
	for {
		if _, err := os.Lstat(current); err == nil {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", fmt.Errorf("no existing path component")
		}
		current = next
	}
}
