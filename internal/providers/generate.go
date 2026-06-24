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
	ProviderID string `json:"providerId"`
	Model      string `json:"model"`
	System     string `json:"system"`
	Prompt     string `json:"prompt"`
	MaxTokens  int    `json:"maxTokens"`
}

type GenerateResult struct {
	ProviderID string   `json:"providerId"`
	Model      string   `json:"model"`
	Text       string   `json:"text"`
	StatusCode int      `json:"statusCode,omitempty"`
	LatencyMs  int64    `json:"latencyMs,omitempty"`
	Tried      []string `json:"tried,omitempty"`
	Error      string   `json:"error,omitempty"`
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
	var tried []string
	for _, provider := range generationCandidates(registry, preferred) {
		tried = append(tried, provider.ID)
		result := Generate(ctx, provider, req)
		result.Tried = append([]string{}, tried...)
		if result.Error == "" && strings.TrimSpace(result.Text) != "" {
			return result
		}
		if len(preferred) == 1 {
			return result
		}
	}
	return GenerateResult{Tried: tried, Error: "no provider candidate succeeded"}
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
	result := GenerateResult{ProviderID: provider.ID, Model: model}
	key, ok := authValue(provider)
	if provider.AuthMode != AuthNone && !ok {
		result.Error = "missing auth environment variable: " + provider.APIKeyEnv
		return result
	}
	body, err := generationPayload(provider, req, model)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	payload, err := json.Marshal(body)
	if err != nil {
		result.Error = err.Error()
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
		return result
	}
	httpReq.Header.Set("content-type", "application/json")
	applyAuth(httpReq, provider, key)
	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		result.Error = fmt.Sprintf("provider returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
		return result
	}
	text, err := ParseGeneration(provider.Protocol, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Text = strings.TrimSpace(text)
	return result
}
