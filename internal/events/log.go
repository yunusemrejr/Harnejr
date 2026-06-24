package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	SessionID string         `json:"sessionId,omitempty"`
	Type      string         `json:"type"`
	Workspace string         `json:"workspace,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type redactionRule struct {
	pattern     *regexp.Regexp
	replacement string
}

var redactionRules = []redactionRule{
	{regexp.MustCompile(`(?i)sk-[A-Za-z0-9_\-]{12,}`), `[REDACTED]`},
	{regexp.MustCompile(`(?i)nvapi-[A-Za-z0-9_\-]{12,}`), `[REDACTED]`},
	{regexp.MustCompile(`(?i)(Bearer\s+)[A-Za-z0-9_\-.]{12,}`), `${1}[REDACTED]`},
	{regexp.MustCompile(`(?i)((api[-_]?key|authorization|token|secret|password)["'\s:=]+)[^"'\s,}]+`), `${1}[REDACTED]`},
}

func Append(memoryDir string, event Event) error {
	if strings.TrimSpace(memoryDir) == "" {
		return nil
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Payload != nil {
		event.Payload = redactMap(event.Payload)
	}
	if err := os.MkdirAll(memoryDir, 0o750); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(memoryDir, "events.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return err
	}
	defer file.Close()
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = file.Write(append(line, '\n'))
	return err
}

func redactMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = RedactValue(v)
	}
	return out
}

func RedactValue(value any) any {
	switch typed := value.(type) {
	case string:
		return RedactString(typed)
	case map[string]any:
		return redactMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = RedactValue(item)
		}
		return out
	default:
		return value
	}
}

func RedactString(value string) string {
	redacted := value
	for _, rule := range redactionRules {
		redacted = rule.pattern.ReplaceAllString(redacted, rule.replacement)
	}
	return redacted
}
