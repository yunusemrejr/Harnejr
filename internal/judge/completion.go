package judge

import "strings"

type CompletionInput struct {
	Goal             string   `json:"goal"`
	Evidence         []string `json:"evidence"`
	Tests            []string `json:"tests"`
	SubagentReviews  int      `json:"subagentReviews"`
	QualityGatePass  bool     `json:"qualityGatePass"`
	ProviderPlanPass bool     `json:"providerPlanPass"`
}

type CompletionDecision struct {
	Accepted bool     `json:"accepted"`
	Blockers []string `json:"blockers"`
}

func Evaluate(input CompletionInput) CompletionDecision {
	var blockers []string
	if strings.TrimSpace(input.Goal) == "" {
		blockers = append(blockers, "goal is required for completion evaluation")
	}
	if len(input.Evidence) == 0 {
		blockers = append(blockers, "completion requires evidence")
	}
	if len(input.Tests) == 0 {
		blockers = append(blockers, "completion requires at least one test, build, lint, or explicit verification command")
	}
	if input.SubagentReviews < 2 {
		blockers = append(blockers, "serious completion requires at least two independent subagent or judge reviews")
	}
	if !input.QualityGatePass {
		blockers = append(blockers, "quality gate has not passed")
	}
	if !input.ProviderPlanPass {
		blockers = append(blockers, "provider/subagent plan has not been validated")
	}
	return CompletionDecision{Accepted: len(blockers) == 0, Blockers: blockers}
}
