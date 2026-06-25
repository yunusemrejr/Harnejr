package providers

import (
	"encoding/json"
	"fmt"
)

type ParsedGeneration struct {
	Text  string
	Usage *UsageMetrics
}

type UsageMetrics struct {
	PromptTokens           int `json:"promptTokens,omitempty"`
	CompletionTokens       int `json:"completionTokens,omitempty"`
	InputTokens            int `json:"inputTokens,omitempty"`
	OutputTokens           int `json:"outputTokens,omitempty"`
	TotalTokens            int `json:"totalTokens,omitempty"`
	PromptCacheHitTokens   int `json:"promptCacheHitTokens,omitempty"`
	PromptCacheMissTokens  int `json:"promptCacheMissTokens,omitempty"`
}

func ParseGeneration(protocol Protocol, data []byte) (string, error) {
	parsed, err := ParseGenerationPayload(protocol, data)
	if err != nil {
		return "", err
	}
	return parsed.Text, nil
}

func ParseGenerationPayload(protocol Protocol, data []byte) (ParsedGeneration, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return ParsedGeneration{}, err
	}
	usage := parseUsage(raw["usage"])
	if text, ok := raw["output_text"].(string); ok {
		return ParsedGeneration{Text: text, Usage: usage}, nil
	}
	if response, ok := raw["response"].(string); ok {
		return ParsedGeneration{Text: response, Usage: usage}, nil
	}
	choices, ok := raw["choices"].([]any)
	if ok && len(choices) > 0 {
		choice, ok := choices[0].(map[string]any)
		if ok {
			msg, ok := choice["message"].(map[string]any)
			if ok {
				if text, ok := msg["content"].(string); ok {
					return ParsedGeneration{Text: text, Usage: usage}, nil
				}
			}
		}
	}
	message, ok := raw["message"].(map[string]any)
	if ok {
		if text, ok := message["content"].(string); ok {
			return ParsedGeneration{Text: text, Usage: usage}, nil
		}
	}
	return ParsedGeneration{}, fmt.Errorf("could not parse %s response text", protocol)
}

func parseUsage(value any) *UsageMetrics {
	usageMap, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	usage := &UsageMetrics{
		PromptTokens:          intFromAny(usageMap["prompt_tokens"]),
		CompletionTokens:      intFromAny(usageMap["completion_tokens"]),
		InputTokens:           intFromAny(usageMap["input_tokens"]),
		OutputTokens:          intFromAny(usageMap["output_tokens"]),
		TotalTokens:           intFromAny(usageMap["total_tokens"]),
		PromptCacheHitTokens:  intFromAny(usageMap["prompt_cache_hit_tokens"]),
		PromptCacheMissTokens: intFromAny(usageMap["prompt_cache_miss_tokens"]),
	}
	if usage.InputTokens == 0 {
		usage.InputTokens = usage.PromptTokens
	}
	if usage.OutputTokens == 0 {
		usage.OutputTokens = usage.CompletionTokens
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}
	return usage
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		parsed, _ := v.Int64()
		return int(parsed)
	default:
		return 0
	}
}
