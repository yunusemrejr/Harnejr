package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Defaults struct {
	Providers map[string]any `json:"providers"`
	Policy    map[string]any `json:"policy"`
	Agents    map[string]any `json:"agents"`
	MCP       map[string]any `json:"mcp"`
	Skills    map[string]any `json:"skills"`
}

func LoadDefaults(configDir string) (Defaults, error) {
	load := func(name string) (map[string]any, error) {
		path := filepath.Join(configDir, name)
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var out map[string]any
		if err := json.Unmarshal(bytes, &out); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		return out, nil
	}

	providers, err := load("providers.default.json")
	if err != nil {
		return Defaults{}, err
	}
	policy, err := load("policy.default.json")
	if err != nil {
		return Defaults{}, err
	}
	agents, err := load("agents.default.json")
	if err != nil {
		return Defaults{}, err
	}
	mcp, err := load("mcp.default.json")
	if err != nil {
		return Defaults{}, err
	}
	skills, err := load("skills.default.json")
	if err != nil {
		return Defaults{}, err
	}

	return Defaults{Providers: providers, Policy: policy, Agents: agents, MCP: mcp, Skills: skills}, nil
}
