package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GenerateRequest struct {
	ProviderID          string `json:"providerId"`
	Model               string `json:"model"`
	System              string `json:"system"`
	Prompt              string `json:"prompt"`
	MaxTokens           int    `json:"maxTokens"`
	AllowBillingChange  bool   `json:"allowBillingChange"`
	CacheMode           string `json:"cacheMode,omitempty"`
	CacheStablePrefix   string `json:"cacheStablePrefix,omitempty"`
	CacheDynamicContext string `json:"cacheDynamicContext,omitempty"`
}

type GenerateResult struct {
	ProviderID            string          `json:"providerId"`
	Model                 string          `json:"model"`
	BillingMode           string          `json:"billingMode,omitempty"`
	Text                  string          `json:"text"`
	StatusCode            int             `json:"statusCode,omitempty"`
	LatencyMs             int64           `json:"latencyMs,omitempty"`
	Tried                 []string        `json:"tried,omitempty"`
	Usage                 *UsageMetrics   `json:"usage,omitempty"`
	Cache                 *CacheTelemetry `json:"cache,omitempty"`
	ErrorClass            string          `json:"errorClass,omitempty"`
	Error                 string          `json:"error,omitempty"`
	BillingChangeRequired bool            `json:"billingChangeRequired,omitempty"`
}

func FindProvider(registry Registry, id string) (ProviderProfile, bool) {
	for _, provider := range registry.Providers {
		if provider.ID == id || provider.OpenCodeProviderID == id {
			return provider, true
		}
		for _, alias := range provider.Aliases {
			if alias == id {
				return provider, true
			}
		}
	}
	return ProviderProfile{}, false
}

func GenerateWithFallback(ctx context.Context, registry Registry, preferred []string, req GenerateRequest) GenerateResult {
	candidates := generationCandidates(registry, preferred)
	if len(candidates) == 0 {
		return GenerateResult{Error: "no provider candidate available", ErrorClass: "no-provider"}
	}
	baseBilling := candidates[0].BillingMode
	var tried []string
	var last GenerateResult
	for index, provider := range candidates {
		if index > 0 && !req.AllowBillingChange && provider.BillingMode != baseBilling {
			last = GenerateResult{ProviderID: provider.ID, Model: provider.DefaultModel, BillingMode: string(provider.BillingMode), Tried: tried, Error: "fallback would change billing mode from " + string(baseBilling) + " to " + string(provider.BillingMode), ErrorClass: "billing-change", BillingChangeRequired: true}
			continue
		}
		tried = append(tried, provider.ID)
		result := Generate(ctx, provider, req)
		result.Tried = append([]string{}, tried...)
		if result.Error == "" && strings.TrimSpace(result.Text) != "" {
			return result
		}
		last = result
		if !retryableGenerationError(result.ErrorClass) {
			continue
		}
	}
	if last.Error != "" {
		last.Tried = tried
		return last
	}
	return GenerateResult{Tried: tried, Error: "no provider candidate succeeded", ErrorClass: "unknown"}
}

func generationCandidates(registry Registry, preferred []string) []ProviderProfile {
	seen := map[string]bool{}
	var out []ProviderProfile
	for _, id := range preferred {
		if provider, ok := FindProvider(registry, id); ok && provider.Enabled && !seen[provider.ID] {
			out = append(out, provider)
			seen[provider.ID] = true
		}
	}
	for _, provider := range registry.Providers {
		if provider.Enabled && !seen[provider.ID] && provider.AuthMode != AuthNone {
			if _, ok := authValue(provider); ok {
				out = append(out, provider)
				seen[provider.ID] = true
			}
		}
	}
	for _, provider := range registry.Providers {
		if provider.Enabled && !seen[provider.ID] && provider.AuthMode == AuthNone {
			out = append(out, provider)
			seen[provider.ID] = true
		}
	}
	return out
}

func Generate(ctx context.Context, provider ProviderProfile, req GenerateRequest) GenerateResult {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = provider.DefaultModel
	}
	if provider.Runtime == RuntimeNodeAISDK || provider.Runtime == RuntimeNodeOpenAI {
		return GenerateViaNode(ctx, provider, req, model)
	}
	result := GenerateResult{ProviderID: provider.ID, Model: model, BillingMode: string(provider.BillingMode)}
	key, ok := authValue(provider)
	if provider.AuthMode != AuthNone && !ok {
		result.Error = "missing auth environment variable: " + provider.APIKeyEnv
		result.ErrorClass = "auth"
		return result
	}
	body, cache, err := generationPayload(provider, req, model)
	result.Cache = cache
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "unsupported-protocol"
		return result
	}
	payload, err := json.Marshal(body)
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "payload"
		return result
	}
	timeout := time.Duration(provider.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, strings.TrimRight(provider.BaseURL, "/")+provider.Endpoint, bytes.NewReader(payload))
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "network"
		return result
	}
	httpReq.Header.Set("content-type", "application/json")
	applyAuth(httpReq, provider, key)
	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "network"
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		result.Error = fmt.Sprintf("provider returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
		result.ErrorClass = ClassifyHTTPStatus(resp.StatusCode)
		return result
	}
	parsed, err := ParseGenerationPayload(provider.Protocol, data)
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "parser"
		return result
	}
	result.Text = strings.TrimSpace(parsed.Text)
	result.Usage = parsed.Usage
	applyUsageToCache(result.Cache, parsed.Usage)
	return result
}

func retryableGenerationError(errorClass string) bool {
	switch errorClass {
	case "network", "timeout", "server-retryable", "rate-limit-or-quota", "sdk-runtime", "stream":
		return true
	default:
		return false
	}
}

func ClassifyHTTPStatus(status int) string {
	switch status {
	case 401, 403:
		return "auth"
	case 404:
		return "wrong-endpoint-or-model"
	case 408:
		return "timeout"
	case 413:
		return "context-too-large"
	case 422:
		return "unsupported-field"
	case 429:
		return "rate-limit-or-quota"
	case 500, 502, 503, 504:
		return "server-retryable"
	default:
		return "http-error"
	}
}
