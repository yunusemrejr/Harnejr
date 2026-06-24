package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yunusemrejr/Harnejr/internal/mcp"
	"github.com/yunusemrejr/Harnejr/internal/providers"
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
			report.fail(name, err.Error())
		} else {
			report.pass(name, "found")
		}
	}
	providerPath := filepath.Join(configDir, "providers.default.json")
	registry, err := providers.LoadRegistry(providerPath)
	if err != nil {
		report.fail("provider-registry", err.Error())
	} else if issues := providers.ValidateRegistry(registry); len(issues) > 0 {
		report.fail("provider-registry", fmt.Sprintf("%d validation issue(s): %s", len(issues), issues[0].Message))
	} else {
		report.pass("provider-registry", fmt.Sprintf("%d providers validated", len(registry.Providers)))
	}
	if len(report.Tools) == 0 {
		report.fail("builtin-tools", "no built-in tools registered")
	} else {
		report.pass("builtin-tools", "built-in tools registered")
	}
	if len(report.MCPSystems) == 0 {
		report.fail("builtin-mcp", "no built-in MCP systems registered")
	} else {
		report.pass("builtin-mcp", "built-in MCP systems registered")
	}
	return report
}

func (r *Report) pass(id string, message string) {
	r.Checks = append(r.Checks, Check{ID: id, Status: "pass", Message: message})
}

func (r *Report) fail(id string, message string) {
	r.Checks = append(r.Checks, Check{ID: id, Status: "fail", Message: message})
	r.Status = "degraded"
}
