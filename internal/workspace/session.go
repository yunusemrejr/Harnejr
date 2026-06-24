package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	GitExisting        = "existing"
	GitInitialized     = "initialized"
	GitSkippedBroad    = "skipped_broad_path"
	GitSkippedNested   = "skipped_nested_repos"
	GitInitUnavailable = "git_unavailable"
)

type PrepareOptions struct {
	Root        string
	SessionID   string
	UserRequest string
	Now         time.Time
}

type PrepareResult struct {
	WorkspaceRoot  string   `json:"workspaceRoot"`
	GitRoot        string   `json:"gitRoot,omitempty"`
	GitStatus      string   `json:"gitStatus"`
	GitMessage     string   `json:"gitMessage"`
	NestedGitRoots []string `json:"nestedGitRoots,omitempty"`
	MemoryDir      string   `json:"memoryDir,omitempty"`
	MemoryFiles    []string `json:"memoryFiles,omitempty"`
	MemoryCreated  bool     `json:"memoryCreated"`
}

func PrepareSessionWorkspace(ctx context.Context, opts PrepareOptions) (PrepareResult, error) {
	root, err := normalizeRoot(opts.Root)
	if err != nil {
		return PrepareResult{}, err
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}

	result := PrepareResult{WorkspaceRoot: root}
	if gitRoot, ok, err := FindGitRoot(root); err != nil {
		return result, err
	} else if ok {
		result.GitRoot = gitRoot
		result.GitStatus = GitExisting
		result.GitMessage = "using existing local git repository"
		return prepareMemory(result, opts)
	}

	if broad, reason := IsBroadWorkspaceRoot(root); broad {
		result.GitStatus = GitSkippedBroad
		result.GitMessage = reason
		return result, nil
	}

	nested, err := FindNestedGitRepos(root)
	if err != nil {
		return result, err
	}
	if len(nested) > 0 {
		result.GitStatus = GitSkippedNested
		result.GitMessage = "workspace contains child git repositories; choose a specific child project instead"
		result.NestedGitRoots = nested
		return result, nil
	}

	if err := InitGitRepository(ctx, root); err != nil {
		if isMissingGitExecutable(err) {
			result.GitStatus = GitInitUnavailable
			result.GitMessage = "git executable was not found; workspace memory was still created"
			return prepareMemory(result, opts)
		}
		return result, err
	}

	result.GitRoot = root
	result.GitStatus = GitInitialized
	result.GitMessage = "initialized a local git repository for this workspace"
	return prepareMemory(result, opts)
}

func FindGitRoot(start string) (string, bool, error) {
	current, err := normalizeRoot(start)
	if err != nil {
		return "", false, err
	}
	for {
		if hasGitMarker(current) {
			return current, true, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", false, nil
		}
		current = next
	}
}

func FindNestedGitRepos(root string) ([]string, error) {
	root, err := normalizeRoot(root)
	if err != nil {
		return nil, err
	}
	var nested []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := entry.Name()
		if path == root {
			return nil
		}
		if name == ".git" {
			nested = append(nested, filepath.Dir(path))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() && shouldSkipNestedRepoScanDir(name) {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(nested)
	return nested, nil
}

func InitGitRepository(ctx context.Context, root string) error {
	root, err := normalizeRoot(root)
	if err != nil {
		return err
	}
	args := []string{"-C", root, "init"}
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("local git initialization failed in %s: %w: %s", root, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func IsBroadWorkspaceRoot(root string) (bool, string) {
	home, _ := os.UserHomeDir()
	return isBroadWorkspaceRootWithHome(root, home)
}

func isBroadWorkspaceRootWithHome(root string, home string) (bool, string) {
	clean := filepath.Clean(root)
	if clean == string(os.PathSeparator) {
		return true, "refusing broad filesystem root"
	}
	protectedAbsolute := map[string]string{
		"/home": "refusing broad /home directory",
		"/Users": "refusing broad /Users directory",
		"/etc":  "refusing system configuration root",
		"/usr":  "refusing system package root",
		"/var":  "refusing system state root",
		"/tmp":  "refusing whole temporary directory",
	}
	if reason, ok := protectedAbsolute[clean]; ok {
		return true, reason
	}
	if home == "" {
		return false, ""
	}
	home = filepath.Clean(home)
	if samePath(clean, home) {
		return true, "refusing user home directory"
	}
	for _, name := range []string{"Desktop", "Documents", "Downloads", "Pictures", "Music", "Videos", "Public", "Templates"} {
		candidate := filepath.Join(home, name)
		if samePath(clean, candidate) {
			return true, fmt.Sprintf("refusing broad user folder %s", name)
		}
	}
	return false, ""
}

func prepareMemory(result PrepareResult, opts PrepareOptions) (PrepareResult, error) {
	memoryRoot := result.GitRoot
	if memoryRoot == "" {
		memoryRoot = result.WorkspaceRoot
	}
	memory, err := EnsureMemory(memoryRoot, MemoryEntry{
		Timestamp:   opts.Now,
		SessionID:   opts.SessionID,
		Workspace:   result.WorkspaceRoot,
		GitRoot:     result.GitRoot,
		GitStatus:   result.GitStatus,
		GitMessage:  result.GitMessage,
		UserRequest: opts.UserRequest,
	})
	if err != nil {
		return result, err
	}
	result.MemoryDir = memory.Dir
	result.MemoryFiles = memory.Files
	result.MemoryCreated = memory.Created
	return result, nil
}

func normalizeRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace root is not a directory: %s", abs)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return filepath.Clean(real), nil
}

func hasGitMarker(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return true
	}
	return false
}

func shouldSkipNestedRepoScanDir(name string) bool {
	switch name {
	case ".harnejr", "node_modules", "vendor", ".cache", ".pnpm-store", "dist", "build":
		return true
	default:
		return false
	}
}

func isMissingGitExecutable(err error) bool {
	var execErr *exec.Error
	return errorsAs(err, &execErr) && execErr.Err == exec.ErrNotFound
}

func samePath(a string, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
