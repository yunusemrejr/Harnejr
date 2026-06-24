package prompts

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UserPrompt struct {
	Content   string    `json:"content"`
	Path      string    `json:"path"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func LoadUserPrompt(configDir string) (UserPrompt, error) {
	path := userPromptPath(configDir)
	bytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return UserPrompt{Path: path}, nil
	}
	if err != nil {
		return UserPrompt{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return UserPrompt{}, err
	}
	return UserPrompt{Content: string(bytes), Path: path, UpdatedAt: info.ModTime().UTC()}, nil
}

func SaveUserPrompt(configDir string, content string) (UserPrompt, error) {
	path := userPromptPath(configDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return UserPrompt{}, err
	}
	content = strings.TrimRight(strings.ReplaceAll(content, "\r\n", "\n"), " \t\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		return UserPrompt{}, err
	}
	return LoadUserPrompt(configDir)
}

func ComposeSystemPrompt(harnessPrompt string, user UserPrompt) string {
	base := strings.TrimSpace(harnessPrompt)
	addition := strings.TrimSpace(user.Content)
	if addition == "" {
		return base
	}
	return base + "\n\n[USER_LEVEL_SYSTEM_PROMPT]\n" + addition + "\n[/USER_LEVEL_SYSTEM_PROMPT]"
}

func userPromptPath(configDir string) string {
	return filepath.Join(configDir, "user.system.md")
}
