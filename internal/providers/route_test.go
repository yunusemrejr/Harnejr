package providers

import "testing"

func TestRoutePrefersStepFunForReview(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		{ID: "streamlake-kat-coding-plan", Aliases: []string{"streamlake"}, Enabled: true, AuthMode: AuthNone, Runtime: RuntimeRawHTTP, BillingMode: BillingSubscription, DefaultModel: "kat-coder-pro-v2", Models: []ModelProfile{{ID: "kat-coder-pro-v2", SupportsReasoning: true, SupportsTools: true, SupportsStreaming: true, CostClass: CostMedium, ContextWindow: 128000}}},
		{ID: "stepfun-step-plan", Aliases: []string{"stepfun-ai"}, Enabled: true, AuthMode: AuthNone, Runtime: RuntimeRawHTTP, BillingMode: BillingSubscription, DefaultModel: "step-3.7-flash", Models: []ModelProfile{{ID: "step-3.7-flash", SupportsReasoning: true, SupportsTools: true, SupportsStreaming: true, CostClass: CostCheap, ContextWindow: 32000}}},
	}}
	decision := Route(registry, RouteRequest{TaskType: "review patch", NeedsReasoning: true, NeedsTools: true})
	if decision.Selected == nil || decision.Selected.ProviderID != "stepfun-step-plan" {
		t.Fatalf("expected StepFun review route, got %#v", decision.Selected)
	}
}

func TestRoutePrefersKatForCodingWhenRequested(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		{ID: "streamlake-kat-coding-plan", Aliases: []string{"streamlake"}, Enabled: true, AuthMode: AuthNone, Runtime: RuntimeRawHTTP, BillingMode: BillingSubscription, DefaultModel: "kat-coder-pro-v2", Models: []ModelProfile{{ID: "kat-coder-pro-v2", SupportsReasoning: true, SupportsTools: true, SupportsStreaming: true, CostClass: CostMedium, ContextWindow: 128000}}},
		{ID: "stepfun-step-plan", Aliases: []string{"stepfun-ai"}, Enabled: true, AuthMode: AuthNone, Runtime: RuntimeRawHTTP, BillingMode: BillingSubscription, DefaultModel: "step-3.7-flash", Models: []ModelProfile{{ID: "step-3.7-flash", SupportsReasoning: true, SupportsTools: true, SupportsStreaming: true, CostClass: CostCheap, ContextWindow: 32000}}},
	}}
	decision := Route(registry, RouteRequest{TaskType: "implement backend patch", NeedsReasoning: true, NeedsTools: true})
	if decision.Selected == nil || decision.Selected.ProviderID != "streamlake-kat-coding-plan" {
		t.Fatalf("expected KAT coding route, got %#v", decision.Selected)
	}
}

func TestRouteRejectsMissingAuthAndDeniedProviders(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		{ID: "needs-key", Enabled: true, AuthMode: AuthBearer, APIKeyEnv: "HARNEJR_TEST_MISSING", Runtime: RuntimeRawHTTP, BillingMode: BillingAPI, DefaultModel: "x", Models: []ModelProfile{{ID: "x", CostClass: CostCheap}}},
		{ID: "local", Enabled: true, AuthMode: AuthNone, Runtime: RuntimeGoNative, BillingMode: BillingLocal, DefaultModel: "local", Models: []ModelProfile{{ID: "local", CostClass: CostFree}}},
	}}
	decision := Route(registry, RouteRequest{Denied: []string{"local"}})
	if decision.Selected != nil {
		t.Fatalf("expected no selected route, got %#v", decision.Selected)
	}
	if len(decision.Rejected) != 2 {
		t.Fatalf("expected two rejections, got %#v", decision.Rejected)
	}
}

func TestRouteLongContextSelectsLargeModel(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		{ID: "minimax-token-plan", Enabled: true, AuthMode: AuthNone, Runtime: RuntimeRawHTTP, BillingMode: BillingSubscription, DefaultModel: "MiniMax-M2", Models: []ModelProfile{
			{ID: "MiniMax-M2", SupportsReasoning: true, SupportsStreaming: true, CostClass: CostCheap, ContextWindow: 64000},
			{ID: "MiniMax-M3", SupportsReasoning: true, SupportsStreaming: true, CostClass: CostMedium, ContextWindow: 1000000},
		}},
	}}
	decision := Route(registry, RouteRequest{TaskType: "document analysis", NeedsLongContext: true, NeedsReasoning: true})
	if decision.Selected == nil || decision.Selected.Model != "MiniMax-M3" {
		t.Fatalf("expected long-context model, got %#v", decision.Selected)
	}
}
