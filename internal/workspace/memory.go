package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultMemoryFiles = []string{
	"README.md",
	"session-log.md",
	"requests.md",
	"decisions.md",
	"errors.md",
	"notices.md",
}

type MemoryEntry struct {
	Timestamp   time.Time
	SessionID   string
	Workspace   string
	GitRoot     string
	GitStatus   string
	GitMessage  string
	UserRequest string
}

type MemoryResult struct {
	Dir     string   `json:"dir"`
	Files   []string `json:"files"`
	Created bool     `json:"created"`
}

func EnsureMemory(root string, entry MemoryEntry) (MemoryResult, error) {
	root, err := normalizeRoot(root)
	if err != nil {
		return MemoryResult{}, err
	}
	dir := filepath.Join(root, ".harnejr")
	created := false
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		created = true
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return MemoryResult{}, err
	}

	for _, file := range defaultMemoryFiles {
		path := filepath.Join(dir, file)
		if err := ensureMemoryFile(path, memorySeedContent(file, root)); err != nil {
			return MemoryResult{}, err
		}
	}
	if err := appendMemoryEntry(dir, entry); err != nil {
		return MemoryResult{}, err
	}

	files := make([]string, 0, len(defaultMemoryFiles))
	for _, file := range defaultMemoryFiles {
		files = append(files, filepath.Join(dir, file))
	}
	return MemoryResult{Dir: dir, Files: files, Created: created}, nil
}

func ensureMemoryFile(path string, seed string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(seed), 0o640)
}

func memorySeedContent(file string, root string) string {
	switch file {
	case "README.md":
		return "# Harnejr workspace memory\n\nThis folder is maintained by Harnejr for this workspace. It stores compact, local notes that help future sessions understand what happened here without relying on chat history.\n\nDo not place secrets in this folder. Provider keys and credentials belong in local environment variables or future secret stores.\n"
	case "session-log.md":
		return "# Session log\n\nChronological Harnejr session entries for this workspace.\n"
	case "requests.md":
		return "# User requests\n\nUser requests and task intents observed while Harnejr worked in this workspace.\n"
	case "decisions.md":
		return "# Decisions\n\nImportant implementation, routing, safety, and architecture decisions made while working here.\n"
	case "errors.md":
		return "# Errors\n\nErrors, failed commands, failed provider calls, blocked actions, and recovery notes.\n"
	case "notices.md":
		return "# Notices\n\nThings noticed, risks, TODOs, incomplete proof, and future-session reminders.\n"
	default:
		return fmt.Sprintf("# %s\n\nHarnejr memory file for %s.\n", file, root)
	}
}

func appendMemoryEntry(dir string, entry MemoryEntry) error {
	when := entry.Timestamp.UTC().Format(time.RFC3339)
	if when == "0001-01-01T00:00:00Z" {
		when = time.Now().UTC().Format(time.RFC3339)
	}
	session := cleanMemoryLine(entry.SessionID)
	if session == "" {
		session = "unspecified"
	}
	workspace := cleanMemoryLine(entry.Workspace)
	gitRoot := cleanMemoryLine(entry.GitRoot)
	if gitRoot == "" {
		gitRoot = "none"
	}
	request := cleanMemoryLine(entry.UserRequest)
	if request == "" {
		request = "not provided"
	}
	block := fmt.Sprintf("\n## %s\n\n- Session: `%s`\n- Workspace: `%s`\n- Git root: `%s`\n- Git status: `%s`\n- Git note: %s\n- User request: %s\n", when, session, workspace, gitRoot, cleanMemoryLine(entry.GitStatus), cleanMemoryLine(entry.GitMessage), request)
	if err := appendText(filepath.Join(dir, "session-log.md"), block); err != nil {
		return err
	}
	if entry.UserRequest != "" {
		requestBlock := fmt.Sprintf("\n## %s\n\n- Session: `%s`\n- Request: %s\n", when, session, request)
		if err := appendText(filepath.Join(dir, "requests.md"), requestBlock); err != nil {
			return err
		}
	}
	return nil
}

func appendText(path string, text string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	return err
}

func cleanMemoryLine(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	for strings.Contains(value, "  ") {
		value = strings.ReplaceAll(value, "  ", " ")
	}
	if len(value) > 600 {
		return value[:600] + "..."
	}
	return value
}
