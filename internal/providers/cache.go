package providers

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const deepSeekCacheStrategy = "stable-prefix-v1"

const deepSeekStablePrefixHeader = `HARNEJR_DEEPSEEK_CACHE_PREFIX_V1
Purpose: preserve byte-stable prompt prefixes for DeepSeek-compatible prompt-cache routes.
Rules:
- Static harness doctrine, provider behavior, tool registry, and role policy stay above the marker.
- Dynamic session state, timestamps, workspace-specific logs, tool output, and user requests stay below the marker.
- Do not rewrite this prefix casually; change the version only when the stable contract intentionally changes.`

var volatilePrefixPattern = regexp.MustCompile(`(?i)(current time|timestamp|session[-_ ]?id|request[-_ ]?id|trace[-_ ]?id|\d{4}-\d{2}-\d{2}t\d{2}:\d{2}:\d{2})`)

type CacheTelemetry struct {
	Eligible          bool     `json:"eligible"`
	Applied           bool     `json:"applied"`
	Strategy          string   `json:"strategy,omitempty"`
	PrefixHash        string   `json:"prefixHash,omitempty"`
	StablePrefixBytes int      `json:"stablePrefixBytes,omitempty"`
	HitTokens         int      `json:"hitTokens,omitempty"`
	MissTokens        int      `json:"missTokens,omitempty"`
	HitRatio          float64  `json:"hitRatio,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
}

func cacheOptimizedChatMessages(provider ProviderProfile, req GenerateRequest, model string) ([]map[string]string, *CacheTelemetry) {
	if !supportsPromptCache(provider, model) {
		return nil, nil
	}
	telemetry := &CacheTelemetry{Eligible: true, Strategy: deepSeekCacheStrategy}
	mode := strings.ToLower(strings.TrimSpace(req.CacheMode))
	if mode == "" {
		mode = "auto"
	}
	if mode == "off" || mode == "disabled" || mode == "none" {
		return nil, telemetry
	}
	if mode != "auto" && mode != "deepseek" && mode != "stable-prefix" {
		telemetry.Warnings = append(telemetry.Warnings, "unknown cacheMode; using stable-prefix optimization")
	}

	stablePrefix := buildStableCachePrefix(req)
	telemetry.Applied = true
	telemetry.PrefixHash = stablePrefixHash(provider.ID, model, stablePrefix)
	telemetry.StablePrefixBytes = len([]byte(stablePrefix))
	telemetry.Warnings = append(telemetry.Warnings, cachePrefixWarnings(stablePrefix)...)

	userContent := buildDynamicCacheContent(req)
	return []map[string]string{
		{"role": "system", "content": stablePrefix},
		{"role": "user", "content": userContent},
	}, telemetry
}

func supportsPromptCache(provider ProviderProfile, model string) bool {
	for _, candidate := range provider.Models {
		if candidate.ID == model && candidate.SupportsPromptCache {
			return true
		}
	}
	id := strings.ToLower(provider.ID + " " + provider.OpenCodeProviderID + " " + strings.Join(provider.Aliases, " ") + " " + model)
	return strings.Contains(id, "deepseek")
}

func buildStableCachePrefix(req GenerateRequest) string {
	stable := strings.TrimSpace(req.CacheStablePrefix)
	if stable == "" {
		stable = strings.TrimSpace(req.System)
	}
	if stable == "" {
		stable = "Harnejr provider worker. Follow daemon policy, workspace boundaries, billing safety, compact memory, and evidence discipline."
	}
	return deepSeekStablePrefixHeader + "\n\n[HARNEJR_STABLE_SYSTEM]\n" + stable + "\n\nCACHE_HIT_OPTIMIZED_STABLE_PREFIX_END"
}

func buildDynamicCacheContent(req GenerateRequest) string {
	var b strings.Builder
	if dynamic := strings.TrimSpace(req.CacheDynamicContext); dynamic != "" {
		b.WriteString("[HARNEJR_DYNAMIC_SESSION_STATE]\n")
		b.WriteString(dynamic)
		b.WriteString("\n\n")
	}
	b.WriteString("[HARNEJR_USER_REQUEST]\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	return b.String()
}

func stablePrefixHash(providerID string, model string, stablePrefix string) string {
	sum := sha256.Sum256([]byte(providerID + "\n" + model + "\n" + stablePrefix))
	return hex.EncodeToString(sum[:])
}

func cachePrefixWarnings(stablePrefix string) []string {
	if volatilePrefixPattern.MatchString(stablePrefix) {
		return []string{"stable prefix appears to contain volatile state; move timestamps/session IDs below CACHE_HIT_OPTIMIZED_STABLE_PREFIX_END"}
	}
	return nil
}

func applyUsageToCache(cache *CacheTelemetry, usage *UsageMetrics) {
	if cache == nil || usage == nil {
		return
	}
	cache.HitTokens = usage.PromptCacheHitTokens
	cache.MissTokens = usage.PromptCacheMissTokens
	denominator := cache.HitTokens + cache.MissTokens
	if denominator > 0 {
		cache.HitRatio = float64(cache.HitTokens) / float64(denominator)
	}
}
