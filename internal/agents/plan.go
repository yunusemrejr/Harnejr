package agents

import "strings"

type PlannedAgent struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerId"`
	Model      string `json:"model"`
	Role       string `json:"role"`
	Reason     string `json:"reason"`
	Required   bool   `json:"required"`
}

type PlanRequest struct {
	Task              string `json:"task"`
	Mode              string `json:"mode"`
	RequestedProvider string `json:"requestedProvider,omitempty"`
	RequestedModel    string `json:"requestedModel,omitempty"`
}

type Plan struct {
	Serious          bool           `json:"serious"`
	RequiresJudge    bool           `json:"requiresJudge"`
	RequiresSubagents bool           `json:"requiresSubagents"`
	Agents           []PlannedAgent `json:"agents"`
	Warnings         []string       `json:"warnings,omitempty"`
}

func BuildPlan(req PlanRequest) Plan {
	task := strings.ToLower(req.Task)
	serious := isSerious(task) || req.Mode == "goal"
	plan := Plan{Serious: serious, RequiresJudge: serious || req.Mode == "goal", RequiresSubagents: serious}
	if !serious {
		plan.Agents = append(plan.Agents, stepfun("quick-verifier", "cheap sanity review for a non-trivial harness task", false))
		return plan
	}
	plan.Agents = append(plan.Agents,
		stepfun("stepfun-verifier", "mandatory cheap verifier for serious work", true),
		PlannedAgent{ID: "minimax-reader", ProviderID: "minimax-token-plan", Model: "MiniMax-M3", Role: "long-context-reader", Reason: "independent context and regression review", Required: true},
	)
	if needsImplementation(task) || req.RequestedProvider == "streamlake-kat-coding-plan" || strings.Contains(req.RequestedModel, "kat-coder") {
		plan.Agents = append(plan.Agents, PlannedAgent{ID: "kat-coder", ProviderID: "streamlake-kat-coding-plan", Model: "kat-coder-pro-v2", Role: "implementation-reviewer", Reason: "bounded coding pressure", Required: true})
		if !containsStepFun(plan.Agents) {
			plan.Agents = append([]PlannedAgent{stepfun("kat-companion", "KAT usage requires StepFun companion verification", true)}, plan.Agents...)
		}
	}
	if req.Mode == "goal" || strings.Contains(task, "complete") || strings.Contains(task, "production") {
		plan.Agents = append(plan.Agents, PlannedAgent{ID: "nim-skeptic", ProviderID: "nvidia-build-nim", Model: "nvidia/nemotron-3-ultra-550b-a55b", Role: "completion-skeptic", Reason: "goal completion must be challenged by an independent reviewer", Required: true})
	}
	return plan
}

func stepfun(id string, reason string, required bool) PlannedAgent {
	return PlannedAgent{ID: id, ProviderID: "stepfun-step-plan", Model: "step-3.7-flash", Role: "verifier", Reason: reason, Required: required}
}

func containsStepFun(agents []PlannedAgent) bool {
	for _, agent := range agents {
		if agent.ProviderID == "stepfun-step-plan" {
			return true
		}
	}
	return false
}

func isSerious(task string) bool {
	if len(task) > 160 {
		return true
	}
	for _, word := range []string{"production", "release", "autonomous", "provider", "shell", "workspace", "security", "subagent", "judge", "mcp", "database", "migration", "refactor", "install", "update"} {
		if strings.Contains(task, word) {
			return true
		}
	}
	return false
}

func needsImplementation(task string) bool {
	for _, word := range []string{"implement", "fix", "patch", "write", "edit", "code", "refactor", "bug"} {
		if strings.Contains(task, word) {
			return true
		}
	}
	return false
}
