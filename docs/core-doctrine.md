# Harnejr core doctrine

Harnejr succeeds only if the harness behaves like an engineering control system, not a chat box with tools.

## 1. Hardened daemon first

The Go daemon is the authority. It owns workspace resolution, policy, containment, rollback safety, provider calls, event ledgers, goal state, completion gates, and runtime health. Prompts may guide agents, but prompts do not enforce anything.

A feature is not real until daemon code enforces it, tests cover it, Doctor reports it, and the session ledger can prove it.

## 2. Containment and rollback safety

Autonomous work must be scoped to a prepared workspace. Harnejr must refuse broad system roots, protect secret-adjacent files, prefer sandboxed shell execution, record mutating actions, and keep changes reversible.

Destructive or ambiguous actions must be denied or transformed into safer alternatives. Controlled autonomy means the agent can iterate without asking the user, not that it can bypass safety.

## 3. Focused tool layer

The tool layer must stay small, precise, and inspectable. File read, file write, search, patch, shell, provider, MCP, skill, worker, and judge tools should be separate and narrow. Extra tool surface area is risk.

Tools should return structured evidence, not vague prose. Errors should be classified so the harness can decide whether to retry, reroute, deny, or stop.

## 4. Goal system with checkpoints

A goal is not a sticky prompt. A goal must decompose into checkpoints, evidence requirements, verification commands, quality gates, review gates, and a final completion decision.

The harness must not accept a done-claim from the main model unless evidence exists. Serious goals require independent review and a skeptical completion check.

## 5. Controlled autonomy

YOLO and goal mode should remove unnecessary confirmation loops for safe workspace work. They must not weaken hard guards. If an action is blocked, the agent should reason again and choose a non-destructive local path.

The user should not be asked to babysit ordinary work, but the harness must still fail closed on unsafe actions.

## 6. Smart model routing

Provider selection must match the task. Cheap fast models should handle routine checks. Coding specialists should handle bounded implementation. Long-context models should handle repository digestion. Strong reviewers should challenge completion.

Providers are transport contracts: protocol, endpoint, auth, billing mode, model namespace, parser, quota behavior, and fallback policy. Harnejr must not silently switch billing paths.

## 7. Simple multi-provider wiring

Multi-provider support must reduce lock-in without becoming untraceable. Every provider route should have an explicit ID, health state, billing mode, model list, auth source, and parser.

Fallback must be logged. If fallback would change billing mode, it must require explicit permission.

## 8. Tight prompts and no slop

System prompts should be short, stable, and evidence-oriented. They should demand concrete steps, files inspected, commands run, risks, and verification. They must forbid unsupported completion claims and generic slop.

Dynamic timestamps, noisy reminders, and volatile telemetry do not belong at the top of cache-sensitive prompts.

## 9. Lean session memory

Session memory should carry forward only what matters: current goal, workspace, decisions, errors, evidence, provider failures, files touched, tests run, and unresolved risks.

Raw chat history bloat should be summarized or represented as structured context. The memory system should help the next step, not preserve everything indiscriminately.

## 10. Production acceptance rule

Harnejr is not production-ready because a screen looks good or a config key exists. It is production-ready only when the daemon enforces safety, the UI exposes necessary controls clearly, provider configuration is editable, secrets are handled locally, tests and smoke checks pass, and completion reports contain enough evidence to audit the session.
