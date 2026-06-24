package quality

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LoCOptions struct {
	Root     string `json:"root"`
	MaxLines int    `json:"maxLines"`
}

type LoCFile struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
}

type LoCReport struct {
	Root       string    `json:"root"`
	MaxLines   int       `json:"maxLines"`
	Scanned    int       `json:"scanned"`
	Oversized  []LoCFile `json:"oversized"`
	Completion string    `json:"completion"`
}

func ScanLoC(opts LoCOptions) (LoCReport, error) {
	root := strings.TrimSpace(opts.Root)
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return LoCReport{}, err
	}
	if opts.MaxLines <= 0 {
		opts.MaxLines = 1000
	}
	report := LoCReport{Root: abs, MaxLines: opts.MaxLines, Completion: "pass"}
	err = filepath.WalkDir(abs, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := entry.Name()
		if entry.IsDir() {
			if shouldSkipDir(name) && path != abs {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSourceFile(name) || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		lines, err := countLines(path)
		if err != nil {
			return err
		}
		report.Scanned++
		if lines > opts.MaxLines {
			rel, err := filepath.Rel(abs, path)
			if err != nil {
				rel = path
			}
			report.Oversized = append(report.Oversized, LoCFile{Path: rel, Lines: lines})
		}
		return nil
	})
	if err != nil {
		return LoCReport{}, err
	}
	sort.Slice(report.Oversized, func(i, j int) bool {
		return report.Oversized[i].Lines > report.Oversized[j].Lines
	})
	if len(report.Oversized) > 0 {
		report.Completion = "review_required"
	}
	return report, nil
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	return lines, scanner.Err()
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".harnejr", "node_modules", "vendor", "dist", "build", "coverage", ".cache", ".pnpm-store":
		return true
	default:
		return false
	}
}

func isSourceFile(name string) bool {
	switch filepath.Ext(name) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".py", ".rs", ".java", ".kt", ".c", ".h", ".cpp", ".hpp", ".cs", ".php", ".rb", ".swift", ".vue", ".svelte", ".css", ".scss", ".html":
		return true
	default:
		return false
	}
}
