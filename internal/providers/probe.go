package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ProbeResult struct {
	ProviderID  string   `json:"providerId"`
	Mode        string   `json:"mode"`
	Ready       bool     `json:"ready"`
	AuthPresent bool     `json:"authPresent"`
	Endpoint    string   `json:"endpoint"`
	Model       string   `json:"model"`
	Live        bool     `json:"live"`
	StatusCode  int      `json:"statusCode,omitempty"`
	LatencyMs   int64    `json:"latencyMs,omitempty"`
	Issues      []string `json:"issues,omitempty"`
}

func ProbeStatic(registry Registry) []ProbeResult {
	issuesByProvider := map[string][]string{}
	for _, issue := range ValidateRegistry(registry) {
		issuesByProvider[issue.ProviderID] = append(issuesByProvider[issue.ProviderID], issue.Message)
	}
	results := make([]ProbeResult, 0, len(registry.Providers))
	for _, provider := range registry.Providers {
		result := baseProbeResult(provider)
		result.Issues = append(result.Issues, issuesByProvider[provider.ID]...)
		_, present := authValue(provider)
		result.AuthPresent = provider.AuthMode == AuthNone || provider.APIKeySecretRef != "" || present
		if !result.AuthPresent && provider.Enabled {
			result.Issues = append(result.Issues, "auth environment variable is not set: "+provider.APIKeyEnv)
		}
		result.Ready = len(result.Issues) == 0
		results = append(results, result)
	}
	return results
}

func ProbeLive(ctx context.Context, registry Registry, providerID string) []ProbeResult {
	var results []ProbeResult
	for _, provider := range registry.Providers {
		if providerID != "" && provider.ID != providerID {
			continue
		}
		result := LiveProbeProvider(ctx, provider)
		results = append(results, result)
	}
	return results
}

func LiveProbeProvider(ctx context.Context, provider ProviderProfile) ProbeResult {
	result := baseProbeResult(provider)
	result.Live = true
	if !provider.Enabled {
		result.Issues = append(result.Issues, "provider is disabled")
		result.Ready = false
		return result
	}
	key, ok := authValue(provider)
	result.AuthPresent = provider.AuthMode == AuthNone || provider.APIKeySecretRef != "" || ok
	if provider.AuthMode != AuthNone && !ok {
		result.Issues = append(result.Issues, "auth environment variable is not set: "+provider.APIKeyEnv)
		result.Ready = false
		return result
	}
	body, err := probePayload(provider)
	if err != nil {
		result.Issues = append(result.Issues, err.Error())
		return result
	}
	payload, err := json.Marshal(body)
	if err != nil {
		result.Issues = append(result.Issues, err.Error())
		return result
	}
	probeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodPost, result.Endpoint, bytes.NewReader(payload))
	if err != nil {
		result.Issues = append(result.Issues, err.Error())
		return result
	}
	req.Header.Set("content-type", "application/json")
	applyAuth(req, provider, key)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Issues = append(result.Issues, err.Error())
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		result.Issues = append(result.Issues, fmt.Sprintf("provider returned HTTP %d", resp.StatusCode))
		return result
	}
	result.Ready = true
	return result
}

func baseProbeResult(provider ProviderProfile) ProbeResult {
	return ProbeResult{ProviderID: provider.ID, Mode: string(provider.BillingMode), Endpoint: strings.TrimRight(provider.BaseURL, "/") + provider.Endpoint, Model: provider.DefaultModel}
}

func authValue(provider ProviderProfile) (string, bool) {
	if provider.APIKeyEnv == "" {
		return "", false
	}
	value := os.Getenv(provider.APIKeyEnv)
	return value, value != ""
}

func applyAuth(req *http.Request, provider ProviderProfile, key string) {
	for name, value := range provider.CustomHeaders {
		req.Header.Set(name, value)
	}
	switch provider.AuthMode {
	case AuthBearer:
		req.Header.Set("authorization", "Bearer "+key)
	case AuthAPIKeyHeader, AuthXAPIKey:
		req.Header.Set(provider.AuthHeaderName, key)
	}
}

func probePayload(provider ProviderProfile) (map[string]any, error) {
	switch provider.Protocol {
	case ProtocolOpenAIChat:
		return map[string]any{"model": provider.DefaultModel, "messages": []map[string]string{{"role": "user", "content": "Reply with ok."}}, "stream": false, "max_tokens": 8}, nil
	case ProtocolOpenAIResponses:
		return map[string]any{"model": provider.DefaultModel, "input": "Reply with ok.", "store": false, "max_output_tokens": 8}, nil
	case ProtocolOllamaNative:
		return map[string]any{"model": provider.DefaultModel, "messages": []map[string]string{{"role": "user", "content": "Reply with ok."}}, "stream": false}, nil
	case ProtocolAnthropicMessages:
		return map[string]any{"model": provider.DefaultModel, "max_tokens": 8, "messages": []map[string]string{{"role": "user", "content": "Reply with ok."}}}, nil
	default:
		return nil, fmt.Errorf("live probe unsupported for protocol %s", provider.Protocol)
	}
}
