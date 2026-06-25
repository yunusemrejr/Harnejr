package providers

import "testing"

func TestGenerationPayloadCapsOutputTokens(t *testing.T) {
	provider := ProviderProfile{Protocol: ProtocolOpenAIChat, DefaultModel: "small", Models: []ModelProfile{{ID: "small", MaxOutputTokens: 128}}}
	payload, _, err := generationPayload(provider, GenerateRequest{Prompt: "x", MaxTokens: 999}, "small")
	if err != nil {
		t.Fatal(err)
	}
	if payload["max_tokens"] != 128 {
		t.Fatalf("expected capped max_tokens, got %#v", payload["max_tokens"])
	}
}

func TestGenerationPayloadAppliesOverridesLast(t *testing.T) {
	provider := ProviderProfile{Protocol: ProtocolOpenAIChat, RequestDefaults: map[string]any{"temperature": 0.7}, PayloadOverrides: map[string]any{"temperature": 0.1}}
	payload, _, err := generationPayload(provider, GenerateRequest{Prompt: "x", MaxTokens: 10}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if payload["temperature"] != 0.1 {
		t.Fatalf("override not applied last: %#v", payload)
	}
}
