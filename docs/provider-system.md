# Provider system

Harnejr treats provider setup as a transport contract, not a base URL plus model string.

A provider profile records provider ID, protocol, runtime, billing mode, base URL, endpoint path, auth mode, model namespace, reasoning adapter, streaming parser, timeout, retry policy, and fallback notes.

This prevents hidden billing changes and endpoint drift. Subscription-backed providers must not silently fall back to pay-as-you-go routes.

## Required provider classes

| Provider ID | Role |
| --- | --- |
| `stepfun-step-plan` | Cheap mandatory verifier and frequent subagent worker. |
| `streamlake-kat-coding-plan` | Strong bounded coding or implementation subagent. Must be paired with StepFun. |
| `minimax-token-plan` | Long-context reading and repository analysis. |
| `kimi-code-subscription` | Coding and review subagent through the subscription coding path. |
| `deepseek-api` | Cache-sensitive large-context reasoning and coding. |
| `nvidia-build-nim` | Independent deep review and completion skepticism. |
| `ollama-local` | Local fallback and secret-adjacent work. |
| `ollama-cloud-native` | Cloud open-model diversity. |
| `xiaomi-mimo-token-plan` | Secondary coding and reasoning provider with explicit auth header mode. |
| `openrouter-api` | Aggregator fallback and experimentation. |

## Startup validation target

Every enabled provider must eventually pass auth presence checks, endpoint reachability checks, tiny non-streaming calls, streaming parser checks, reasoning field checks when configured, optional tool-call checks, and billing-mode logging before automatic routing.

## Error classes

Provider errors should be classified into stable buckets: auth, billing or quota, rate limit, wrong endpoint, wrong model, unsupported field, context too large, output too large, retryable server error, network, stream parser, unsupported tool call, and unknown.
