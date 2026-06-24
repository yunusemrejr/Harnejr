package workspace

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SummaryResult struct {
	Path        string   `json:"path"`
	Sources     []string `json:"sources"`
	GeneratedAt string   `json:"generatedAt"`
}

func BuildSummary(memoryDir string) (SummaryResult, error) {
	when := time.Now().UTC().Format(time.RFC3339)
	outPath := filepath.Join(memoryDir, "summary.md")
	files := []string{"session-log.md", "requests.md", "decisions.md", "notices.md", "goal.json", "control.json"}
	var out strings.Builder
	out.WriteString("# Harnejr compact memory\n\n")
	out.WriteString("Generated: " + when + "\n\n")
	result := SummaryResult{Path: outPath, GeneratedAt: when}
	for _, name := range files {
		lines, err := lastLines(filepath.Join(memoryDir, name), 24)
		if os.IsNotExist(err) || len(lines) == 0 {
			continue
		}
		if err != nil {
			return SummaryResult{}, err
		}
		result.Sources = append(result.Sources, name)
		out.WriteString("## " + name + "\n\n")
		for _, line := range lines {
			out.WriteString("- " + line + "\n")
		}
		out.WriteString("\n")
	}
	if len(result.Sources) == 0 {
		out.WriteString("No source notes yet.\n")
	}
	if err := os.WriteFile(outPath, []byte(out.String()), 0o640); err != nil {
		return SummaryResult{}, err
	}
	return result, nil
}

func lastLines(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if len(line) > 220 {
			line = line[:220] + "..."
		}
		lines = append(lines, line)
		if len(lines) > limit {
			lines = lines[1:]
		}
	}
	return lines, scanner.Err()
}
