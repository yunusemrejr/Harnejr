# DeepSeek cache-hit discipline

Date: 2026-06-25

Harnejr treats DeepSeek prompt caching as runtime economics, not a prompt-only reminder. DeepSeek-compatible provider calls now get a deterministic stable-prefix layout when the selected provider/model advertises prompt-cache support.

## Behavior

For OpenAI-compatible chat providers with prompt-cache support, Harnejr builds messages as:

```text
system:
  HARNEJR_DEEPSEEK_CACHE_PREFIX_V1
  static cache contract
  HARNEJR_STABLE_SYSTEM
  caller-provided stable prefix or stable system prompt
  CACHE_HIT_OPTIMIZED_STABLE_PREFIX_END

user:
  HARNEJR_DYNAMIC_SESSION_STATE, when supplied
  HARNEJR_USER_REQUEST
```

The stable prefix intentionally excludes timestamps, session IDs, changing workspace logs, raw tool output, and one-off request text. Those belong below `CACHE_HIT_OPTIMIZED_STABLE_PREFIX_END`.

## Request controls

`POST /api/llm/generate` and `POST /api/llm/stream` accept:

```json
{
  "cacheMode": "auto",
  "cacheStablePrefix": "stable project / harness / tool policy text",
  "cacheDynamicContext": "current task state, recent evidence, latest diff summary"
}
```

Values for `cacheMode`:

- `auto`: default; applies stable-prefix optimization to prompt-cache-capable providers.
- `deepseek` or `stable-prefix`: explicit stable-prefix behavior.
- `off`, `disabled`, or `none`: keeps the original message layout but still reports cache eligibility.

## Response telemetry

Provider responses now include cache telemetry when applicable:

```json
{
  "cache": {
    "eligible": true,
    "applied": true,
    "strategy": "stable-prefix-v1",
    "prefixHash": "...",
    "stablePrefixBytes": 512,
    "hitTokens": 9000,
    "missTokens": 1200,
    "hitRatio": 0.8823,
    "warnings": []
  },
  "usage": {
    "promptTokens": 10200,
    "completionTokens": 120,
    "totalTokens": 10320,
    "promptCacheHitTokens": 9000,
    "promptCacheMissTokens": 1200
  }
}
```

`prefixHash` is namespaced by provider and model so Harnejr can detect drift across similar calls. The warning list flags likely volatile text in the stable prefix.

## Reasonix-inspired rules carried into Harnejr

- Keep stable harness doctrine first.
- Keep stable provider and tool behavior text before dynamic state.
- Keep user requests and latest session facts below the stable-prefix marker.
- Preserve the same prefix across subagents when the role contract is the same.
- Measure cache hits and misses from provider usage instead of claiming cache behavior by status text.

## Verification

The provider tests now cover:

- stable prefix hash remains unchanged when only the dynamic prompt changes;
- `cacheMode: "off"` preserves the original messages;
- DeepSeek `prompt_cache_hit_tokens` and `prompt_cache_miss_tokens` are parsed into provider telemetry;
- volatile stable-prefix state emits a warning.
