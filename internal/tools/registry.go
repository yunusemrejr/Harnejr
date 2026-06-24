package tools

type Permission string

type Tool struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Permission       Permission `json:"permission"`
	MCPSystem        string     `json:"mcpSystem"`
	OutputLimitBytes int        `json:"outputLimitBytes"`
	Available        bool       `json:"available"`
}

const (
	PermissionReadOnly      Permission = "read_only"
	PermissionWorkspaceSafe Permission = "workspace_safe"
	PermissionControl       Permission = "control"
)

func Builtins() []Tool {
	return []Tool{
		{
			ID:               "harnejr.doctor",
			Name:             "Doctor",
			Description:      "Inspect daemon, configuration, workspace, policy, and readiness state.",
			Permission:       PermissionReadOnly,
			MCPSystem:        "harnejr-doctor",
			OutputLimitBytes: 65536,
			Available:        true,
		},
		{
			ID:               "harnejr.loc.scan",
			Name:             "LoC Controller",
			Description:      "Scan source files and flag files that exceed configured line-count thresholds.",
			Permission:       PermissionReadOnly,
			MCPSystem:        "loc-controller",
			OutputLimitBytes: 65536,
			Available:        true,
		},
		{
			ID:               "harnejr.workspace.prepare",
			Name:             "Workspace Prepare",
			Description:      "Prepare Git and local workspace memory before a session starts.",
			Permission:       PermissionWorkspaceSafe,
			MCPSystem:        "workspace-memory",
			OutputLimitBytes: 32768,
			Available:        true,
		},
		{
			ID:               "harnejr.goal.set",
			Name:             "Goal Control",
			Description:      "Set, update, or clear scoped goal state for a workspace session.",
			Permission:       PermissionControl,
			MCPSystem:        "goal-topic-controller",
			OutputLimitBytes: 32768,
			Available:        true,
		},
		{
			ID:               "harnejr.topic.set",
			Name:             "Topic Control",
			Description:      "Set or clear the active topic so agents keep work scoped.",
			Permission:       PermissionControl,
			MCPSystem:        "goal-topic-controller",
			OutputLimitBytes: 32768,
			Available:        true,
		},
		{
			ID:               "harnejr.healing.plan",
			Name:             "Autonomous Healing Plan",
			Description:      "Create deterministic repair steps from doctor and quality findings.",
			Permission:       PermissionReadOnly,
			MCPSystem:        "autonomous-healer",
			OutputLimitBytes: 65536,
			Available:        true,
		},
		{
			ID:               "harnejr.context.pack",
			Name:             "Context Pack",
			Description:      "Summarize the current workspace state into a compact continuation payload.",
			Permission:       PermissionReadOnly,
			MCPSystem:        "context-efficiency",
			OutputLimitBytes: 65536,
			Available:        true,
		},
	}
}
