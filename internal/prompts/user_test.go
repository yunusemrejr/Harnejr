package prompts

import (
	"strings"
	"testing"
)

func TestSaveAndLoadUserPrompt(t *testing.T) {
	dir := t.TempDir()
	saved, err := SaveUserPrompt(dir, "Prefer small files.\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(saved.Content, "Prefer small files.") {
		t.Fatalf("saved prompt missing content: %#v", saved)
	}
	loaded, err := LoadUserPrompt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Content != saved.Content {
		t.Fatalf("loaded prompt mismatch")
	}
}

func TestComposeSystemPromptAppendsUserPrompt(t *testing.T) {
	composed := ComposeSystemPrompt("core harness prompt", UserPrompt{Content: "personal rule"})
	if !strings.Contains(composed, "core harness prompt") || !strings.Contains(composed, "personal rule") {
		t.Fatalf("composed prompt missing parts: %s", composed)
	}
	if !strings.Contains(composed, "USER_LEVEL_SYSTEM_PROMPT") {
		t.Fatalf("composed prompt missing boundary marker: %s", composed)
	}
}
