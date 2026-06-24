package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type StreamEvent struct {
	ProviderID string `json:"providerId"`
	Model      string `json:"model"`
	Text       string `json:"text,omitempty"`
	Reasoning  string `json:"reasoning,omitempty"`
	Done       bool   `json:"done"`
	Error      string `json:"error,omitempty"`
}

func Stream(ctx context.Context, provider ProviderProfile, req GenerateRequest, emit func(StreamEvent) error) GenerateResult {
	model := strings.TrimSpace(req.Model)
	if model == "" { model = provider.DefaultModel }
	result := GenerateResult{ProviderID: provider.ID, Model: model, BillingMode: string(provider.BillingMode)}
	key, ok := authValue(provider)
	if provider.AuthMode != AuthNone && !ok {
		result.Error = "missing auth environment variable: " + provider.APIKeyEnv
		result.ErrorClass = "auth"
		_ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true})
		return result
	}
	body, err := generationPayload(provider, req, model)
	if err != nil { result.Error = err.Error(); result.ErrorClass = "unsupported-protocol"; _ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true}); return result }
	body["stream"] = true
	payload, err := json.Marshal(body)
	if err != nil { result.Error = err.Error(); result.ErrorClass = "payload"; _ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true}); return result }
	timeout := time.Duration(provider.TimeoutMs) * time.Millisecond
	if timeout <= 0 { timeout = 120 * time.Second }
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, strings.TrimRight(provider.BaseURL, "/")+provider.Endpoint, bytes.NewReader(payload))
	if err != nil { result.Error = err.Error(); result.ErrorClass = "network"; _ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true}); return result }
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "text/event-stream")
	applyAuth(httpReq, provider, key)
	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil { result.Error = err.Error(); result.ErrorClass = "network"; _ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true}); return result }
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		result.Error = strings.TrimSpace(string(data))
		result.ErrorClass = ClassifyHTTPStatus(resp.StatusCode)
		_ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Error: result.Error, Done: true})
		return result
	}
	var text strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		chunk, done := normalizeStreamLine(provider.Protocol, line)
		if chunk.Text != "" || chunk.Reasoning != "" {
			chunk.ProviderID = provider.ID
			chunk.Model = model
			text.WriteString(chunk.Text)
			if err := emit(chunk); err != nil { result.Error = err.Error(); result.ErrorClass = "stream"; return result }
		}
		if done { break }
	}
	if err := scanner.Err(); err != nil { result.Error = err.Error(); result.ErrorClass = "stream"; return result }
	result.Text = text.String()
	_ = emit(StreamEvent{ProviderID: provider.ID, Model: model, Done: true})
	return result
}

func normalizeStreamLine(protocol Protocol, line string) (StreamEvent, bool) {
	if strings.HasPrefix(line, "data:") { line = strings.TrimSpace(strings.TrimPrefix(line, "data:")) }
	if line == "[DONE]" { return StreamEvent{Done: true}, true }
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil { return StreamEvent{}, false }
	if protocol == ProtocolOllamaNative {
		if msg, ok := raw["message"].(map[string]any); ok {
			if content, ok := msg["content"].(string); ok { return StreamEvent{Text: content}, false }
		}
		if done, _ := raw["done"].(bool); done { return StreamEvent{Done: true}, true }
	}
	if choices, ok := raw["choices"].([]any); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]any); ok {
			if delta, ok := choice["delta"].(map[string]any); ok {
				text, _ := delta["content"].(string)
				reasoning, _ := delta["reasoning_content"].(string)
				return StreamEvent{Text: text, Reasoning: reasoning}, false
			}
		}
	}
	return StreamEvent{}, false
}
