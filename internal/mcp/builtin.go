package mcp

import "github.com/yunusemrejr/Harnejr/internal/tools"

type BuiltinSystem struct {
	ID          string       `json:"id"`
	DisplayName string       `json:"displayName"`
	Kind        string       `json:"kind"`
	Status      string       `json:"status"`
	Description string       `json:"description"`
	Tools       []tools.Tool `json:"tools"`
}

func BuiltinSystems() []BuiltinSystem {
	bySystem := map[string][]tools.Tool{}
	for _, tool := range tools.Builtins() {
		bySystem[tool.MCPSystem] = append(bySystem[tool.MCPSystem], tool)
	}
	return []BuiltinSystem{
		{ID: "harnejr-doctor", DisplayName: "Harnejr Doctor", Kind: "builtin", Status: "ready", Description: "Read-only readiness inspection for the local harness.", Tools: bySystem["harnejr-doctor"]},
		{ID: "loc-controller", DisplayName: "LoC Controller", Kind: "builtin", Status: "ready", Description: "Source line-count scanning for oversized files.", Tools: bySystem["loc-controller"]},
		{ID: "goal-topic-controller", DisplayName: "Goal and Topic Controller", Kind: "builtin", Status: "ready", Description: "Scoped goal, topic, mode, and loop state controls for sessions.", Tools: bySystem["goal-topic-controller"]},
		{ID: "autonomous-healer", DisplayName: "Autonomous Healer", Kind: "builtin", Status: "ready", Description: "Repair planning from doctor, policy, and quality findings.", Tools: bySystem["autonomous-healer"]},
		{ID: "workspace-memory", DisplayName: "Workspace Memory", Kind: "builtin", Status: "ready", Description: "Project-local memory and workspace preparation primitives.", Tools: bySystem["workspace-memory"]},
		{ID: "context-efficiency", DisplayName: "Context Efficiency", Kind: "builtin", Status: "ready", Description: "Compact workspace state packaging for efficient agent continuation.", Tools: bySystem["context-efficiency"]},
	}
}
