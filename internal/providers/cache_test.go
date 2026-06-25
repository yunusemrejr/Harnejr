package providers

import (
	"strings"
	"testing"
)

func TestDeepSeekCacheStablePrefixHashSurvivesPromptChange(t *testing.T) {
	provider := ProviderProfile{
		ID:       "deepseek-api",
		Protocol: ProtocolOpenAIChat,
		Models: []ModelProfile{{
			ID:                  "deepseek-v4-flash",
			SupportsPromptCache: true,
		}},
	}
	reqA := GenerateRequest{System: "Stable harness policy", Prompt: "First dynamic task"}
	reqB := GenerateRequest{System: "Stable harness policy", Prompt: "Second dynamic task"}

	messagesA, cacheA := cacheOptimizedChatMessages(provider, reqA, "deepseek-v4-flash")
	messagesB, cacheB := cacheOptimizedChatMessages(provider, reqB, "deepseek-v4-flash")

	if cacheA == nil || cacheB == nil || !cacheA.Applied || !cacheB.Applied {
		t.Fatalf("expected cache optimization to apply: %#v %#v", cacheA, cacheB)
	}
	if cacheA.PrefixHash != cacheB.PrefixHash {
		t.Fatalf("stable prefix hash drifted when only dynamic prompt changed: %s vs %s", cacheA.PrefixHash, cacheB.PrefixHash)
	}
	if messagesA[0]["content"] != messagesB[0]["content"] {
		t.Fatalf("system cache prefix changed when only dynamic prompt changed")
	}
	if messagesA[1]["content"] == messagesB[1]["content"] {
		t.Fatalf("dynamic message should still carry the different user prompt")
	}
	if !strings.Contains(messagesA[0]["content"], "CACHE_HIT_OPTIMIZED_STABLE_PREFIX_END") {
		t.Fatalf("stable prefix marker missing: %s", messagesA[0]["content"])
	}
}

func TestDeepSeekCacheModeOffKeepsOriginalMessages(t *testing.T) {
	provider := ProviderProfile{
		ID:       "deepseek-api",
		Protocol: ProtocolOpenAIChat,
		Models: []ModelProfile{{
			ID:                  "deepseek-v4-flash",
			SupportsPromptCache: true,
		}},
	}
	req := GenerateRequest{System: "System", Prompt: "Prompt", CacheMode: "off"}
	body, cache, err := generationPayload(provider, req, "deepseek-v4-flash")
	if err != nil {
		t.Fatal(err)
	}
	if cache == nil || !cache.Eligible || cache.Applied {
		t.Fatalf("expected eligible but unapplied cache telemetry: %#v", cache)
	}
	messages := body["messages"].([]map[string]string)
	if len(messages) != 2 || messages[0]["content"] != "System" || messages[1]["content"] != "Prompt" {
		t.Fatalf("cacheMode=off should preserve original messages: %#v", messages)
	}
}

func TestParseDeepSeekPromptCacheUsage(t *testing.T) {
	payload := []byte(`{
		"choices": [{"message": {"content": "ok"}}],
		"usage": {
			"prompt_tokens": 120,
			"completion_tokens": 8,
			"total_tokens": 128,
			"prompt_cache_hit_tokens": 90,
			"prompt_cache_miss_tokens": 30
		}
	}`)
	parsed, err := ParseGenerationPayload(ProtocolOpenAIChat, payload)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Text != "ok" {
		t.Fatalf("unexpected text: %q", parsed.Text)
	}
	if parsed.Usage == nil || parsed.Usage.PromptCacheHitTokens != 90 || parsed.Usage.PromptCacheMissTokens != 30 {
		t.Fatalf("cache usage not parsed: %#v", parsed.Usage)
	}
	cache := &CacheTelemetry{Eligible: true, Applied: true}
	applyUsageToCache(cache, parsed.Usage)
	if cache.HitTokens != 90 || cache.MissTokens != 30 || cache.HitRatio != 0.75 {
		t.Fatalf("cache telemetry not updated: %#v", cache)
	}
}

func TestStablePrefixWarnsOnVolatileState(t *testing.T) {
	provider := ProviderProfile{ID: "deepseek-api", Protocol: ProtocolOpenAIChat, Models: []ModelProfile{{ID: "deepseek-v4-flash", SupportsPromptCache: true}}}
	_, cache := cacheOptimizedChatMessages(provider, GenerateRequest{System: "timestamp=2026-06-25T12:00:00", Prompt: "task"}, "deepseek-v4-flash")
	if cache == nil || len(cache.Warnings) == 0 {
		t.Fatalf("expected volatile stable-prefix warning: %#v", cache)
	}
}
