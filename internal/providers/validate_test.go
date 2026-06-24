package providers

import (
	"path/filepath"
	"testing"
)

func TestDefaultProviderRegistryValidates(t *testing.T) {
	registry, err := LoadRegistry(filepath.Join("..", "..", "configs", "providers.default.json"))
	if err != nil {
		t.Fatal(err)
	}
	if issues := ValidateRegistry(registry); len(issues) > 0 {
		t.Fatalf("default provider registry has validation issues: %#v", issues)
	}
}

func TestValidateRegistryRejectsDuplicateProviderIDs(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		minimalProvider("same"),
		minimalProvider("same"),
	}}
	issues := ValidateRegistry(registry)
	if len(issues) == 0 {
		t.Fatal("expected duplicate provider id issue")
	}
}

func TestValidateRegistryRejectsMissingAuth(t *testing.T) {
	provider := minimalProvider("bad-auth")
	provider.APIKeyEnv = ""
	registry := Registry{Version: 1, Providers: []ProviderProfile{provider}}
	issues := ValidateRegistry(registry)
	if len(issues) == 0 {
		t.Fatal("expected missing auth issue")
	}
}

func minimalProvider(id string) ProviderProfile {
	return ProviderProfile{
		ID:              id,
		DisplayName:     id,
		Enabled:         true,
		Protocol:        ProtocolOpenAIChat,
		Runtime:         RuntimeRawHTTP,
		BillingMode:     BillingAPI,
		BaseURL:         "https://example.com/v1",
		Endpoint:        "/chat/completions",
		APIKeyEnv:       "EXAMPLE_API_KEY",
		AuthMode:        AuthBearer,
		DefaultModel:    "model-a",
		StreamingParser: "openai_sse",
		TimeoutMs:       1,
		MaxRetries:      0,
		Models: []ModelProfile{{
			ID:                "model-a",
			SupportsStreaming: true,
			CostClass:         CostCheap,
		}},
	}
}
