package providers

import (
	"encoding/json"
	"fmt"
)

func ParseGeneration(protocol Protocol, data []byte) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", err
	}
	if text, ok := raw["output_text"].(string); ok {
		return text, nil
	}
	if response, ok := raw["response"].(string); ok {
		return response, nil
	}
	choices, ok := raw["choices"].([]any)
	if ok && len(choices) > 0 {
		choice, ok := choices[0].(map[string]any)
		if ok {
			msg, ok := choice["message"].(map[string]any)
			if ok {
				if text, ok := msg["content"].(string); ok {
					return text, nil
				}
			}
		}
	}
	message, ok := raw["message"].(map[string]any)
	if ok {
		if text, ok := message["content"].(string); ok {
			return text, nil
		}
	}
	return "", fmt.Errorf("could not parse %s response text", protocol)
}
