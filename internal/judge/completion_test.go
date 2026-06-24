package judge

import "testing"

func TestEvaluateBlocksWeakCompletion(t *testing.T) {
	decision := Evaluate(CompletionInput{Goal: "ship production harness"})
	if decision.Accepted {
		t.Fatal("expected weak completion to be blocked")
	}
	if len(decision.Blockers) == 0 {
		t.Fatal("expected blockers")
	}
}

func TestEvaluateAcceptsEvidenceBackedCompletion(t *testing.T) {
	decision := Evaluate(CompletionInput{
		Goal:             "ship production harness",
		Evidence:         []string{"diff reviewed", "doctor ready"},
		Tests:            []string{"go test ./..."},
		SubagentReviews:  2,
		QualityGatePass:  true,
		ProviderPlanPass: true,
	})
	if !decision.Accepted {
		t.Fatalf("expected completion acceptance, got %#v", decision)
	}
}
