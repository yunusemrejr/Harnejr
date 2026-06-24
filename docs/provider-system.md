# Provider system

Harnejr treats provider setup as a transport contract, not a base URL plus model string.

A provider profile records provider ID, aliases, OpenCode-compatible provider ID when relevant, protocol, runtime, billing mode, base URL, endpoint path, auth mode, model namespace, reasoning adapter, streaming parser, timeout, retry policy, and fallback notes.

This prevents hidden billing changes and endpoint drift. Subscription-backed providers must not silently fall back to pay-as-you-go routes.

## Default provider classes

| Provider ID | OpenCode-compatible alias | Role |
| --- | --- | --- |
| `openai-api` | `openai` | Direct OpenAI Responses API route. API billing only. |
| `deepseek-api` | `deepseek` | Cache-sensitive large-context reasoning and coding. |
| `stepfun-step-plan` | `stepfun-ai` | Cheap mandatory verifier and frequent subagent worker. |
| `streamlake-kat-coding-plan` | `streamlake` | KAT Coder implementation specialist. Must be paired with StepFun. |
| `minimax-token-plan` | `minimax` | Long-context research, document reading, and repository analysis. |
| `kimi-code-subscription` | `kimi-code` | Kimi Code subscription coding path. |
| `nvidia-build-nim` | `nvidia-nim` | Independent deep review, skill/MCP exploration, and completion skepticism. |
| `ollama-local` | none | Local fallback and secret-adjacent work. |
| `ollama-cloud` | `ollama-cloud` | OpenAI-compatible Ollama Cloud route used by OpenCode-style configs. |
| `ollama-cloud-native` | none | Native Ollama Cloud `/api` route. |
| `xiaomi-mimo-token-plan` | `xiaomi` | Secondary coding and reasoning provider with explicit auth header mode. |
| `xai-supergrok-oauth` | `supergrok` | OAuth-backed Responses-shaped xAI route. |
| `openrouter-api` | `openrouter` | Aggregator fallback and experimentation. |
| `custom-openai-compatible` | none | Disabled editable placeholder for local gateways and corporate routes. |

## OpenCode guide alignment

The default registry intentionally includes OpenCode-compatible aliases and model rosters for StreamLake, Ollama Cloud, NVIDIA NIM, StepFun AI, and MiniMax. These aliases are compatibility metadata, not an excuse to collapse billing paths. The canonical Harnejr provider IDs remain explicit so the daemon can distinguish subscription, API-billed, local, and OAuth-backed routes.

## Startup validation target

Every enabled provider must eventually pass auth presence checks, endpoint reachability checks, tiny non-streaming calls, streaming parser checks, reasoning field checks when configured, optional tool-call checks, and billing-mode logging before automatic routing.

The current daemon Doctor performs static provider-registry validation: duplicate IDs, duplicate aliases, unknown enum values, malformed base URLs, missing auth references, invalid endpoints, enabled placeholders, missing default models, duplicate model IDs, and impossible output/context limits.

## Error classes

Provider errors should be classified into stable buckets: auth, billing or quota, rate limit, wrong endpoint, wrong model, unsupported field, context too large, output too large, retryable server error, network, stream parser, unsupported tool call, and unknown.
