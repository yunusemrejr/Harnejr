package providers

import (
	"os"
	"strings"
)

type ProbeResult struct {
	ProviderID  string   `json:"providerId"`
	Mode        string   `json:"mode"`
	Ready       bool     `json:"ready"`
	AuthPresent bool     `json:"authPresent"`
	Endpoint    string   `json:"endpoint"`
	Model       string   `json:"model"`
	Issues      []string `json:"issues,omitempty"`
}

func ProbeStatic(registry Registry) []ProbeResult {
	issuesByProvider := map[string][]string{}
	for _, issue := range ValidateRegistry(registry) {
		issuesByProvider[issue.ProviderID] = append(issuesByProvider[issue.ProviderID], issue.Message)
	}
	results := make([]ProbeResult, 0, len(registry.Providers))
	for _, provider := range registry.Providers {
		result := ProbeResult{
			ProviderID: provider.ID,
			Mode:       string(provider.BillingMode),
			Endpoint:   strings.TrimRight(provider.BaseURL, "/") + provider.Endpoint,
			Model:      provider.DefaultModel,
			Issues:     issuesByProvider[provider.ID],
		}
		if provider.AuthMode == AuthNone || provider.APIKeySecretRef != "" || (provider.APIKeyEnv != "" && os.Getenv(provider.APIKeyEnv) != "") {
			result.AuthPresent = true
		} else if provider.Enabled {
			result.Issues = append(result.Issues, "auth environment variable is not set: "+provider.APIKeyEnv)
		}
		result.Ready = len(result.Issues) == 0
		results = append(results, result)
	}
	return results
}
