package providers

import (
	"sort"
	"strings"
)

type RouteCandidate struct {
	ProviderID  string    `json:"providerId"`
	Model       string    `json:"model"`
	Runtime     Runtime   `json:"runtime"`
	BillingMode string    `json:"billingMode"`
	CostClass   CostClass `json:"costClass"`
	Score       int       `json:"score"`
	Reason      string    `json:"reason"`
}

type RouteRejection struct {
	ProviderID string `json:"providerId"`
	Reason     string `json:"reason"`
}

type RouteDecision struct {
	Selected   *RouteCandidate  `json:"selected,omitempty"`
	Candidates []RouteCandidate `json:"candidates"`
	Rejected   []RouteRejection `json:"rejected,omitempty"`
}

func Route(registry Registry, req RouteRequest) RouteDecision {
	decision := RouteDecision{}
	preferred := preferenceRank(req.Preferred)
	for _, provider := range registry.Providers {
		candidate, rejection, ok := routeCandidate(provider, req, preferred)
		if ok {
			decision.Candidates = append(decision.Candidates, candidate)
		} else {
			decision.Rejected = append(decision.Rejected, rejection)
		}
	}
	sort.SliceStable(decision.Candidates, func(i, j int) bool {
		left := decision.Candidates[i]
		right := decision.Candidates[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if costRank(left.CostClass) != costRank(right.CostClass) {
			return costRank(left.CostClass) < costRank(right.CostClass)
		}
		return left.ProviderID < right.ProviderID
	})
	if len(decision.Candidates) > 0 {
		selected := decision.Candidates[0]
		decision.Selected = &selected
	}
	return decision
}

func routeCandidate(provider ProviderProfile, req RouteRequest, preferred map[string]int) (RouteCandidate, RouteRejection, bool) {
	if !provider.Enabled {
		return RouteCandidate{}, RouteRejection{ProviderID: provider.ID, Reason: "provider disabled"}, false
	}
	for _, id := range req.Denied {
		if providerMatches(provider, id) {
			return RouteCandidate{}, RouteRejection{ProviderID: provider.ID, Reason: "provider excluded by request"}, false
		}
	}
	if provider.AuthMode != AuthNone {
		if _, ok := authValue(provider); !ok {
			return RouteCandidate{}, RouteRejection{ProviderID: provider.ID, Reason: "auth not configured"}, false
		}
	}
	model, modelReason, ok := routeModel(provider, req)
	if !ok {
		return RouteCandidate{}, RouteRejection{ProviderID: provider.ID, Reason: modelReason}, false
	}
	if req.MaxCostClass != "" && req.MaxCostClass != CostUnknown && costRank(model.CostClass) > costRank(req.MaxCostClass) {
		return RouteCandidate{}, RouteRejection{ProviderID: provider.ID, Reason: "model cost exceeds max cost class"}, false
	}
	score, reason := routeScore(provider, model, req, preferred)
	return RouteCandidate{ProviderID: provider.ID, Model: model.ID, Runtime: provider.Runtime, BillingMode: string(provider.BillingMode), CostClass: model.CostClass, Score: score, Reason: strings.TrimSpace(modelReason + "; " + reason)}, RouteRejection{}, true
}

func preferenceRank(values []string) map[string]int {
	out := map[string]int{}
	for index, value := range values {
		out[strings.TrimSpace(value)] = len(values) - index
	}
	return out
}

func providerMatches(provider ProviderProfile, id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	if provider.ID == id || provider.OpenCodeProviderID == id {
		return true
	}
	for _, alias := range provider.Aliases {
		if alias == id {
			return true
		}
	}
	return false
}

func routeModel(provider ProviderProfile, req RouteRequest) (ModelProfile, string, bool) {
	if len(provider.Models) == 0 {
		if req.NeedsReasoning || req.NeedsTools || req.NeedsLongContext {
			return ModelProfile{}, "provider has no model capability metadata", false
		}
		return ModelProfile{ID: provider.DefaultModel, SupportsStreaming: true, CostClass: CostUnknown}, "using default model", true
	}
	models := make([]ModelProfile, 0, len(provider.Models))
	for _, model := range provider.Models {
		if model.ID == "" {
			continue
		}
		if req.NeedsTools && !model.SupportsTools {
			continue
		}
		if req.NeedsReasoning && !model.SupportsReasoning {
			continue
		}
		if req.NeedsLongContext && model.ContextWindow < 128000 {
			continue
		}
		models = append(models, model)
	}
	if len(models) == 0 {
		return ModelProfile{}, "no model satisfies requested capabilities", false
	}
	sort.SliceStable(models, func(i, j int) bool {
		if req.NeedsLongContext && models[i].ContextWindow != models[j].ContextWindow {
			return models[i].ContextWindow > models[j].ContextWindow
		}
		if models[i].ID == provider.DefaultModel {
			return true
		}
		if models[j].ID == provider.DefaultModel {
			return false
		}
		if costRank(models[i].CostClass) != costRank(models[j].CostClass) {
			return costRank(models[i].CostClass) < costRank(models[j].CostClass)
		}
		return models[i].ID < models[j].ID
	})
	return models[0], "model satisfies requested capabilities", true
}

func routeScore(provider ProviderProfile, model ModelProfile, req RouteRequest, preferred map[string]int) (int, string) {
	score := 100
	var reasons []string
	for id, rank := range preferred {
		if providerMatches(provider, id) {
			score += 500 + rank*10
			reasons = append(reasons, "preferred provider")
			break
		}
	}
	task := strings.ToLower(req.TaskType)
	if strings.Contains(task, "review") || strings.Contains(task, "verify") || strings.Contains(task, "judge") || strings.Contains(task, "test") {
		score += familyScore(provider, map[string]int{"stepfun": 80, "nvidia": 55, "minimax": 45, "deepseek": 35, "kimi": 25})
		reasons = append(reasons, "review profile")
	}
	if strings.Contains(task, "implement") || strings.Contains(task, "code") || strings.Contains(task, "executor") || strings.Contains(task, "refactor") {
		score += familyScore(provider, map[string]int{"streamlake": 90, "kat": 90, "kimi": 60, "deepseek": 45, "minimax": 35, "stepfun": 20})
		reasons = append(reasons, "coding profile")
	}
	if strings.Contains(task, "summar") || strings.Contains(task, "research") || strings.Contains(task, "document") {
		score += familyScore(provider, map[string]int{"minimax": 75, "deepseek": 45, "kimi": 40, "ollama": 30, "stepfun": 25})
		reasons = append(reasons, "reading profile")
	}
	if req.NeedsLongContext {
		score += familyScore(provider, map[string]int{"minimax": 80, "deepseek": 70, "nvidia": 45, "kimi": 40}) + model.ContextWindow/20000
		reasons = append(reasons, "long-context need")
	}
	if req.NeedsReasoning {
		score += 30
		reasons = append(reasons, "reasoning need")
	}
	if req.NeedsTools {
		score += 20
		reasons = append(reasons, "tool-call need")
	}
	score += map[CostClass]int{CostFree: 25, CostCheap: 20, CostMedium: 8, CostExpensive: -10, CostUnknown: 0}[model.CostClass]
	return score, strings.Join(reasons, ", ")
}

func familyScore(provider ProviderProfile, weights map[string]int) int {
	ids := append([]string{provider.ID, provider.OpenCodeProviderID, provider.DisplayName}, provider.Aliases...)
	best := 0
	for _, raw := range ids {
		value := strings.ToLower(raw)
		for key, score := range weights {
			if strings.Contains(value, key) && score > best {
				best = score
			}
		}
	}
	return best
}

func costRank(cost CostClass) int {
	switch cost {
	case CostFree:
		return 0
	case CostCheap:
		return 1
	case CostMedium:
		return 2
	case CostExpensive:
		return 3
	default:
		return 4
	}
}
