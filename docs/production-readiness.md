# Production readiness

Harnejr is being moved from scaffold to enforceable runtime behavior. This document records which claims are backed by daemon code and which remain open.

## Backed by daemon code

| Area | Current enforcement |
| --- | --- |
| Install lifecycle | Installer builds daemon and web UI and creates `harnejr update` and `harnejr uninstall`. |
| Workspace preparation | Sessions prepare a Git-rooted workspace and `.harnejr` memory while refusing broad roots and nested-repo parent pollution. |
| Workspace boundaries | File APIs use real-path resolution and reject traversal or symlink escapes. |
| Secret-adjacent file safety | File read, write, and patch APIs reject common secret-adjacent paths. |
| Rollback snapshots | File write and patch endpoints create `.harnejr/rollback/<timestamp>/` backups before mutation. |
| Structured edits | A locked text-patch API replaces one expected block and records the result. |
| Shell execution | Shell commands only run after deterministic `allow` classification. `ask` and `deny` do not run. |
| Ubuntu shell sandbox | Shell execution uses Bubblewrap when `bwrap` is available and reports the sandbox mode. Doctor degrades when Bubblewrap is missing. |
| Event ledger | Runtime actions append redacted JSONL events under `.harnejr/events.jsonl`. |
| Provider registry | Provider defaults are statically validated for IDs, aliases, auth refs, endpoints, model defaults, duplicate models, and impossible limits. |
| Provider probes | Static probes exist; opt-in live probes require one explicit provider. |
| Provider generation | `/api/llm/generate` can call configured providers and blocks silent billing-mode fallback unless explicitly allowed. |
| DeepSeek cache discipline | Prompt-cache-capable DeepSeek routes get a stable-prefix message layout, prefix hashes, volatile-prefix warnings, and parsed `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens` telemetry. |
| Streaming generation | `/api/llm/stream` streams one explicit provider as SSE and normalizes common OpenAI and Ollama chunks. |
| Checkpointed goals | `/api/goals/start`, `/api/goals/status`, and `/api/goals/checkpoint` persist a goal as verifiable checkpoints before completion review. |
| Goal/topic state | Goal, topic, loop, and yolo state persist into workspace memory. |
| Lean memory | `/api/memory/summary` writes a compact `summary.md` from useful workspace memory sources instead of raw history bloat. |
| Subagent planning | Serious tasks produce deterministic plans; KAT usage requires StepFun companion planning. |
| Provider-backed workers | `/api/workers/run` executes planned provider workers with bounded concurrency. |
| Completion gate | Completion checks can block weak done-claims lacking evidence, tests, reviews, quality gates, or provider-plan validation. |
| Provider-backed review | `/api/review/run` combines the deterministic completion gate with an independent provider review. |
| Skills discovery | Daemon discovers global and workspace skills and agents. |
| MCP checks | Daemon parses MCP config and reports missing commands or environment variables. |
| Web UI coverage | The installed UI exposes provider generation, workers, review, provider probe, MCP check, Doctor, memory save, checkpointed goal start, memory compaction, goal/topic/yolo controls, and user prompt editing. |
| CI | CI runs Go tests/build, production smoke tests, and TypeScript install/build/typecheck. |

## Still open

| Area | Remaining gap |
| --- | --- |
| Streaming parser depth | SSE exists and emits usage/cache data when providers send it, but provider-specific tool-call and final-event parsers need more coverage. |
| Provider fallback policy | Billing-mode protection exists, but cooldown ledgers, quota budgets, and retry policies need more depth. |
| Shell sandboxing | Bubblewrap is preferred when present, but full seccomp/container hardening is not complete. |
| Edit engine | Locked text replacement exists; multi-hunk diff application and dry-run previews need more work. |
| Worker execution | Provider-backed workers exist, but result grading, retries, and long-run ledgers are still first-pass. |
| Judge execution | Provider-backed review exists, but rotating multi-judge execution is not complete. |
| External MCP handshakes | Command/env checks exist; process initialize, list-tools, and tool-smoke calls still need implementation. |
| SQLite session state | Current durable state is JSONL and project files, not SQLite. |

## Production rule

Do not call a feature complete because a prompt, README, config key, or UI label says it exists. It is complete only when daemon code enforces it, tests cover it, Doctor reports it, and a session export proves it.
