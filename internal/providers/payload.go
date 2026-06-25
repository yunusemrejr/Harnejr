package providers

import "fmt"

func generationPayload(provider ProviderProfile, req GenerateRequest, model string) (map[string]any, *CacheTelemetry, error) {
	limit := req.MaxTokens
	if limit <= 0 {
		limit = 1024
	}
	body := map[string]any{"model": model}
	for key, value := range provider.RequestDefaults {
		body[key] = value
	}
	for key, value := range provider.ExtraBody {
		body[key] = value
	}
	if defaults, ok := provider.ModelRequestDefaults[model]; ok {
		for key, value := range defaults {
			body[key] = value
		}
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
	return body, cache, nil
}

func chatMessages(req GenerateRequest) []map[string]string {
	messages := []map[string]string{}
	if req.System != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.System})
	}
	return append(messages, map[string]string{"role": "user", "content": req.Prompt})
}
