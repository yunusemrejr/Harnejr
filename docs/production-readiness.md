# Production readiness

Harnejr is being moved from scaffold to enforceable runtime behavior. This document records which claims are currently backed by daemon code and which remain roadmap items.

## Confirmed by daemon primitives

| Area | Current enforcement |
| --- | --- |
| Install lifecycle | `install.sh` bootstraps source, builds daemon and web UI, installs launcher, and generates `harnejr update` / `harnejr uninstall`. |
| Workspace preparation | Each session can prepare a Git-rooted workspace and `.harnejr` memory while refusing broad roots and parent folders with nested repos. |
| Workspace boundaries | File APIs use real-path workspace resolution and reject traversal/symlink escapes. |
| Secret-adjacent file safety | File read/write/patch APIs reject common secret paths such as `.env`, SSH keys, credential/token files, and secret-named files. |
| Structured edits | A locked text-patch API can replace one expected text block and records the change in the workspace ledger. |
| Shell execution | `/api/shell/run` only runs commands classified as deterministic `allow`; `ask` and `deny` do not execute. |
| Event ledger | Mutating/runtime actions append redacted JSONL events under `.harnejr/events.jsonl`. |
| Provider registry | Default providers are validated for IDs, aliases, auth references, endpoints, model defaults, duplicate models, and impossible output/context limits. |
| Provider probes | Static provider probes are available by default; opt-in live probes can issue tiny provider calls for one explicit provider. |
| Provider generation | `/api/llm/generate` can call configured providers in a non-streaming mode and use fallback candidates when available. |
| Goal/topic state | Session controls persist into workspace memory. |
| Subagent planning | Serious tasks produce deterministic subagent plans; KAT usage requires StepFun companion planning. |
| Provider-backed workers | `/api/workers/run` uses the subagent plan and attempts provider-backed worker calls. |
| Completion gate | Completion checks can block claims lacking evidence, tests, reviews, quality gates, or provider-plan validation. |
| Provider-backed review | `/api/review/run` runs the deterministic completion gate and attempts an independent provider review. |
| Skills discovery | Daemon can discover global skills and agents from `~/skills`, `~/.agents`, `~/.codex/skills`, and workspace `.harnejr/skills`. |
| MCP checks | Daemon can parse external MCP config and report missing commands or env vars. |
| Doctor | Doctor validates core config files, built-in systems, tools, and provider registry health. |
| CI | CI runs Go tests/build, production smoke tests, and TypeScript install/build/typecheck. |

## Still not production-complete

| Area | Missing proof or implementation |
| --- | --- |
| Streaming provider runtime | Non-streaming provider calls exist, but SSE/JSONL streaming normalization is not complete. |
| Provider fallback policy | Fallback exists, but quota-aware billing-change policy and cooldown ledgers are still minimal. |
| Shell sandboxing | Shell execution is policy-gated and workspace-scoped, but not yet containerized/seccomp sandboxed. |
| Advanced edit engine | Locked text replacement exists; multi-hunk diff application, dry-run previews, and file locks across processes need more work. |
| Subagent execution | Provider-backed worker calls exist, but concurrency, result validity grading, retries, and long-run ledgers are still first-pass. |
| Judge execution | Provider-backed review exists, but it is not yet a rotating multi-judge system. |
| External MCP handshakes | Config command/env checks exist; full process start, initialize, list-tools, and tool-smoke calls are still needed. |
| SQLite session state | Current durable state is local files/JSONL. SQLite event history remains blocked/not implemented. |
| Web UI coverage | The UI exposes core controls but does not yet expose every new production endpoint. |

## Production rule

Do not call a feature complete because a prompt, README, config key, or UI label says it exists. It is complete only when daemon code enforces it, tests cover it, Doctor reports it, and a session export proves it.
