package providers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuthValueReadsLocalKeyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "provider.key")
	if err := os.WriteFile(path, []byte("test-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, ok := authValue(ProviderProfile{APIKeySecretRef: path})
	if !ok || value != "test-value" {
		t.Fatalf("local key file was not read: %q %v", value, ok)
	}
}

func TestSaveAndLoadRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "providers.default.json")
	registry := Registry{Version: 1, Providers: []ProviderProfile{{ID: "local", DisplayName: "Local", Enabled: true, Protocol: ProtocolOllamaNative, Runtime: RuntimeGoNative, BillingMode: BillingLocal, BaseURL: "http://localhost:11434/api", Endpoint: "/chat", AuthMode: AuthNone, DefaultModel: "model", StreamingParser: "ollama_jsonl"}}}
	if err := SaveRegistry(path, registry); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Providers) != 1 || loaded.Providers[0].ID != "local" {
		t.Fatalf("unexpected loaded registry: %#v", loaded)
	}
}
