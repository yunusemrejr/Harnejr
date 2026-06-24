package mcp

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Version int         `json:"version"`
	Servers []ServerDef `json:"servers"`
}

type ServerDef struct {
	ID      string            `json:"id"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type Check struct {
	ID            string   `json:"id"`
	Configured    bool     `json:"configured"`
	CommandExists bool     `json:"commandExists"`
	EnvPresent    bool     `json:"envPresent"`
	Issues        []string `json:"issues,omitempty"`
}

func CheckServers(configPath string) ([]Check, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil { return nil, err }
	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil { return nil, err }
	out := make([]Check, 0, len(cfg.Servers))
	for _, server := range cfg.Servers {
		check := Check{ID: server.ID, Configured: true, EnvPresent: true}
		if server.Command == "" {
			check.Issues = append(check.Issues, "command is empty")
		} else if commandExists(server.Command) {
			check.CommandExists = true
		} else {
			check.Issues = append(check.Issues, "command not found")
		}
		for name := range server.Env {
			if os.Getenv(name) == "" {
				check.EnvPresent = false
				check.Issues = append(check.Issues, "missing environment variable: "+name)
			}
		}
		out = append(out, check)
	}
	return out, nil
}

func commandExists(command string) bool {
	if filepath.IsAbs(command) {
		info, err := os.Stat(command)
		return err == nil && !info.IsDir()
	}
	_, err := exec.LookPath(command)
	return err == nil
}
