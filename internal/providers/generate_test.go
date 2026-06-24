package providers

import "testing"

func TestFindProviderUsesAlias(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{{ID: "canonical", Aliases: []string{"alias"}}}}
	provider, ok := FindProvider(registry, "alias")
	if !ok || provider.ID != "canonical" {
		t.Fatalf("alias did not resolve: %#v %v", provider, ok)
	}
}

func TestGenerateFallbackCandidatesIncludePreferredFirst(t *testing.T) {
	registry := Registry{Version: 1, Providers: []ProviderProfile{
		{ID: "a", Enabled: true, AuthMode: AuthNone},
		{ID: "b", Enabled: true, AuthMode: AuthNone},
	}}
	candidates := generationCandidates(registry, []string{"b"})
	if len(candidates) != 2 || candidates[0].ID != "b" || candidates[1].ID != "a" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
}
