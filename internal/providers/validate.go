package providers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Registry struct {
	Version   int               `json:"version"`
	Providers []ProviderProfile `json:"providers"`
}

type ValidationIssue struct {
	ProviderID string `json:"providerId,omitempty"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
}

func LoadRegistry(path string) (Registry, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, err
	}
	var registry Registry
	if err := json.Unmarshal(bytes, &registry); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func ValidateRegistry(registry Registry) []ValidationIssue {
	var issues []ValidationIssue
	if registry.Version < 1 {
		issues = append(issues, ValidationIssue{Field: "version", Message: "registry version must be positive"})
	}
	if len(registry.Providers) == 0 {
		issues = append(issues, ValidationIssue{Field: "providers", Message: "provider registry is empty"})
	}
	seenIDs := map[string]bool{}
	seenAliases := map[string]string{}
	for _, provider := range registry.Providers {
		id := strings.TrimSpace(provider.ID)
		if id == "" {
			issues = append(issues, ValidationIssue{Field: "id", Message: "provider id is required"})
			continue
		}
		if seenIDs[id] {
			issues = append(issues, ValidationIssue{ProviderID: id, Field: "id", Message: "duplicate provider id"})
		}
		seenIDs[id] = true
		for _, alias := range provider.Aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				issues = append(issues, ValidationIssue{ProviderID: id, Field: "aliases", Message: "empty alias"})
				continue
			}
			if owner, ok := seenAliases[alias]; ok && owner != id {
				issues = append(issues, ValidationIssue{ProviderID: id, Field: "aliases", Message: fmt.Sprintf("alias %q already belongs to %s", alias, owner)})
			}
			seenAliases[alias] = id
		}
		issues = append(issues, validateProvider(provider)...)
	}
	return issues
}

func validateProvider(provider ProviderProfile) []ValidationIssue {
	id := provider.ID
	var issues []ValidationIssue
	if strings.TrimSpace(provider.DisplayName) == "" {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "displayName", Message: "display name is required"})
	}
	if !validProtocol(provider.Protocol) {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "protocol", Message: "unknown protocol"})
	}
	if !validRuntime(provider.Runtime) {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "runtime", Message: "unknown runtime"})
	}
	if !validBilling(provider.BillingMode) {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "billingMode", Message: "unknown billing mode"})
	}
	if !validAuth(provider.AuthMode) {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "authMode", Message: "unknown auth mode"})
	}
	if provider.AuthMode != AuthNone && provider.Enabled && provider.APIKeyEnv == "" && provider.APIKeySecretRef == "" {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "apiKeyEnv", Message: "enabled authenticated provider needs apiKeyEnv or apiKeySecretRef"})
	}
	if (provider.AuthMode == AuthAPIKeyHeader || provider.AuthMode == AuthXAPIKey) && strings.TrimSpace(provider.AuthHeaderName) == "" {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "authHeaderName", Message: "header auth mode requires authHeaderName"})
	}
	if provider.Runtime != RuntimeSubprocess && provider.Protocol != ProtocolCLIBacked {
		parsed, err := url.Parse(provider.BaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			issues = append(issues, ValidationIssue{ProviderID: id, Field: "baseURL", Message: "baseURL must be an absolute URL"})
		}
	}
	if !strings.HasPrefix(provider.Endpoint, "/") {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "endpoint", Message: "endpoint must start with /"})
	}
	if provider.Enabled && strings.Contains(provider.BaseURL, "example.invalid") {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "baseURL", Message: "placeholder provider cannot be enabled"})
	}
	if strings.TrimSpace(provider.DefaultModel) == "" {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "defaultModel", Message: "default model is required"})
	}
	seenModels := map[string]bool{}
	defaultFound := len(provider.Models) == 0
	for _, model := range provider.Models {
		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			issues = append(issues, ValidationIssue{ProviderID: id, Field: "models", Message: "model id is required"})
			continue
		}
		if seenModels[modelID] {
			issues = append(issues, ValidationIssue{ProviderID: id, Field: "models", Message: "duplicate model id: " + modelID})
		}
		seenModels[modelID] = true
		if modelID == provider.DefaultModel {
			defaultFound = true
		}
		if model.ContextWindow > 0 && model.MaxOutputTokens > model.ContextWindow {
			issues = append(issues, ValidationIssue{ProviderID: id, Field: "models", Message: "model output limit exceeds context window: " + modelID})
		}
	}
	if !defaultFound {
		issues = append(issues, ValidationIssue{ProviderID: id, Field: "defaultModel", Message: "default model is not present in models list"})
	}
	return issues
}

func validProtocol(value Protocol) bool {
	switch value {
	case ProtocolOpenAIChat, ProtocolOpenAIResponses, ProtocolAnthropicMessages, ProtocolOllamaNative, ProtocolCLIBacked, ProtocolOAuthBacked, ProtocolCustomHTTP:
		return true
	default:
		return false
	}
}

func validRuntime(value Runtime) bool {
	switch value {
	case RuntimeGoNative, RuntimeNodeAISDK, RuntimeNodeOpenAI, RuntimeRawHTTP, RuntimeSubprocess:
		return true
	default:
		return false
	}
}

func validBilling(value BillingMode) bool {
	switch value {
	case BillingAPI, BillingSubscription, BillingLocal, BillingOAuth, BillingUnknown:
		return true
	default:
		return false
	}
}

func validAuth(value AuthMode) bool {
	switch value {
	case AuthBearer, AuthAPIKeyHeader, AuthXAPIKey, AuthCustomHeaders, AuthNone:
		return true
	default:
		return false
	}
}
