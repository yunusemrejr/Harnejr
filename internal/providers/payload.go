package providers

import "fmt"

func generationPayload(provider ProviderProfile, req GenerateRequest, model string) (map[string]any, *CacheTelemetry, error) {
	limit := effectiveMaxTokens(provider, req, model)
	body := map[string]any{"model": model}
	mergeMap(body, provider.RequestDefaults)
	mergeMap(body, provider.ExtraBody)
	if defaults, ok := provider.ModelRequestDefaults[model]; ok {
		mergeMap(body, defaults)
	}
	var cache *CacheTelemetry
	switch provider.Protocol {
	case ProtocolOpenAIChat:
		messages := chatMessages(req)
		if optimized, report := cacheOptimizedChatMessages(provider, req, model); report != nil {
			cache = report
			if report.Applied {
				messages = optimized
			}
		}
		body["messages"] = messages
		body["stream"] = false
		body["max_tokens"] = limit
	case ProtocolOpenAIResponses:
		if req.System != "" {
			body["instructions"] = req.System
		}
		body["input"] = req.Prompt
		body["store"] = false
		body["max_output_tokens"] = limit
	case ProtocolOllamaNative:
		body["messages"] = chatMessages(req)
		body["stream"] = false
	case ProtocolAnthropicMessages:
		if req.System != "" {
			body["system"] = req.System
		}
		body["messages"] = []map[string]string{{"role": "user", "content": req.Prompt}}
		body["max_tokens"] = limit
	default:
		return nil, nil, fmt.Errorf("generation unsupported for protocol %s", provider.Protocol)
	}
	mergeMap(body, provider.PayloadOverrides)
	return body, cache, nil
}

func effectiveMaxTokens(provider ProviderProfile, req GenerateRequest, model string) int {
	limit := req.MaxTokens
	if limit <= 0 {
		limit = 1024
	}
	for _, candidate := range provider.Models {
		if candidate.ID == model && candidate.MaxOutputTokens > 0 && limit > candidate.MaxOutputTokens {
			return candidate.MaxOutputTokens
		}
	}
	return limit
}

func mergeMap(dst map[string]any, src map[string]any) {
	for key, value := range src {
		dst[key] = value
	}
}

func chatMessages(req GenerateRequest) []map[string]string {
	messages := []map[string]string{}
	if req.System != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.System})
	}
	return append(messages, map[string]string{"role": "user", "content": req.Prompt})
}
