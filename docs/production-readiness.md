# Production readiness

Harnejr is being moved from scaffold to enforceable runtime behavior. This document records which claims are currently backed by daemon code and which remain roadmap items.

## Confirmed by daemon primitives

| Area | Current enforcement |
| --- | --- |
| Install lifecycle | `install.sh` bootstraps source, builds daemon and web UI, installs launcher, and generates `harnejr update` / `harnejr uninstall`. |
| Workspace preparation | Each session can prepare a Git-rooted workspace and `.harnejr` memory while refusing broad roots and parent folders with nested repos. |
| Workspace boundaries | File APIs use real-path workspace resolution and reject traversal/symlink escapes. |
| Secret-adjacent file safety | File read/write APIs reject common secret paths such as `.env`, SSH keys, credential/token files, and secret-named files. |
| Shell execution | `/api/shell/run` only runs commands classified as deterministic `allow`; `ask` and `deny` do not execute. |
| Event ledger | Mutating/runtime actions append redacted JSONL events under `.harnejr/events.jsonl`. |
| Provider registry | Default providers are validated for IDs, aliases, auth references, endpoints, model defaults, duplicate models, and impossible output/context limits. |
| Provider probes | Static provider probes are available by default; opt-in live probes can issue tiny provider calls when the user explicitly requests them. |
| Goal/topic state | Session controls persist into workspace memory. |
| Subagent planning | Serious tasks produce deterministic subagent plans; KAT usage requires StepFun companion planning. |
| Completion gate | Completion checks can block claims lacking evidence, tests, reviews, quality gates, or provider-plan validation. |
| Skills discovery | Daemon can discover global skills and agents from `~/skills`, `~/.agents`, `~/.codex/skills`, and workspace `.harnejr/skills`. |
| Doctor | Doctor validates core config files, built-in systems, tools, and provider registry health. |
| CI | CI runs Go tests/build and TypeScript install/build/typecheck. |

## Still not production-complete

| Area | Missing proof or implementation |
| --- | --- |
| Live provider execution | Harnejr can probe providers, but the full streaming provider runtime and normalized model-call loop are not complete. |
| Provider fallback runtime | No runtime scheduler yet replaces failed providers during an active LLM task. |
| Shell sandboxing | Shell execution is policy-gated and workspace-scoped, but not yet containerized/seccomp sandboxed. |
| Workspace edit engine | Safe file write exists; structured patch/diff application and file locks are still needed. |
| Subagent execution | Plans exist, but independent provider-backed subagent execution and result ledgers are not complete. |
| Judge execution | Completion gate exists, but provider-backed judge calls are not yet wired. |
| External MCP handshakes | Built-in local systems exist; external MCP process start, initialize, list-tools, and smoke-call checks are still needed. |
| SQLite session state | Current durable state is local files/JSONL. SQLite event history remains roadmap. |
| Web UI coverage | The UI exposes core controls but does not yet expose every new production endpoint. |

## Production rule

Do not call a feature complete because a prompt, README, config key, or UI label says it exists. It is complete only when daemon code enforces it, tests cover it, Doctor reports it, and a session export proves it.
