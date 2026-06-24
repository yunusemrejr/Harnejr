package doctor

import (
	"os"
	"path/filepath"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/mcp"
	"github.com/yunusemrejr/Harnejr/internal/tools"
)

type Check struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Report struct {
	Name       string              `json:"name"`
	Timestamp  time.Time           `json:"timestamp"`
	Status     string              `json:"status"`
	Checks     []Check             `json:"checks"`
	Tools      []tools.Tool        `json:"tools"`
	MCPSystems []mcp.BuiltinSystem `json:"mcpSystems"`
}

func Run(configDir string) Report {
	report := Report{
		Name:       "harnejr-doctor",
		Timestamp:  time.Now().UTC(),
		Status:     "ready",
		Tools:      tools.Builtins(),
		MCPSystems: mcp.BuiltinSystems(),
	}
	for _, name := range []string{"providers.default.json", "policy.default.json", "agents.default.json", "mcp.default.json", "skills.default.json"} {
		path := filepath.Join(configDir, name)
		if _, err := os.Stat(path); err != nil {
			report.Checks = append(report.Checks, Check{ID: name, Status: "fail", Message: err.Error()})
			report.Status = "degraded"
		} else {
			report.Checks = append(report.Checks, Check{ID: name, Status: "pass", Message: "found"})
		}
	}
	if len(report.Tools) == 0 {
		report.Checks = append(report.Checks, Check{ID: "builtin-tools", Status: "fail", Message: "no built-in tools registered"})
		report.Status = "degraded"
	} else {
		report.Checks = append(report.Checks, Check{ID: "builtin-tools", Status: "pass", Message: "built-in tools registered"})
	}
	if len(report.MCPSystems) == 0 {
		report.Checks = append(report.Checks, Check{ID: "builtin-mcp", Status: "fail", Message: "no built-in MCP systems registered"})
		report.Status = "degraded"
	} else {
		report.Checks = append(report.Checks, Check{ID: "builtin-mcp", Status: "pass", Message: "built-in MCP systems registered"})
	}
	return report
}
