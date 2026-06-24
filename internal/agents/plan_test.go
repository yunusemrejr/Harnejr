package agents

import "testing"

func TestBuildPlanRequiresStepFunWithKAT(t *testing.T) {
	plan := BuildPlan(PlanRequest{Task: "implement production provider routing fix", RequestedModel: "kat-coder-pro-v2"})
	if !plan.RequiresSubagents || !plan.RequiresJudge {
		t.Fatalf("expected serious plan with judge and subagents: %#v", plan)
	}
	hasKAT := false
	hasStepFun := false
	for _, agent := range plan.Agents {
		if agent.ProviderID == "streamlake-kat-coding-plan" {
			hasKAT = true
		}
		if agent.ProviderID == "stepfun-step-plan" {
			hasStepFun = true
		}
	}
	if !hasKAT || !hasStepFun {
		t.Fatalf("expected KAT and StepFun in plan: %#v", plan.Agents)
	}
}
