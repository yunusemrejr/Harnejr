# Harnejr

Harnejr is an open-source, MIT-licensed, Ubuntu-native agentic coding harness with a local web interface. It is built around a Go daemon, a TypeScript web application, editable provider profiles, workspace safety primitives, built-in local harness systems, and project-local memory.

Harnejr is not a prompt wrapper. The daemon owns policy, workspace preparation, persistent controls, quality checks, readiness reporting, provider contracts, event logs, and completion evidence.

## Project status

Harnejr is in active scaffold-to-runtime development. The repository includes a runnable daemon, installable web control surface, default provider registry, policy config, workspace preparation, local memory, shell classification, policy-gated shell execution, safe workspace file APIs, path-boundary checks, built-in local harness systems, Doctor reporting, LoC quality scanning, repair planning, scoped goal/topic controls, deterministic subagent planning, completion evidence checks, provider probes, skills discovery, a session prompt console, configured model selector, command entry, and a permanent user-level system prompt editor.

Harnejr is ready for serious local harness development and controlled workspace testing. Full autonomous provider-backed coding, provider fallback during live tasks, external MCP handshakes, subagent execution, judge execution, and sandbox/container execution are still under implementation. See `docs/production-readiness.md` for the current confirmed/not-confirmed matrix.

## Install

Single-command install on Ubuntu:

```bash
curl -fsSL https://raw.githubusercontent.com/yunusemrejr/Harnejr/main/install.sh -o /tmp/harnejr-install.sh && bash /tmp/harnejr-install.sh
```

Then launch:

```bash
harnejr
```

The installer clones or updates the source checkout, runs Go tests, installs pnpm workspace dependencies, builds the web UI, builds the daemon, copies `apps/web/dist` into the install directory, copies default configs, and writes the `harnejr` launcher.

Requirements: Ubuntu Linux, Git, curl, Go 1.22 or newer, Node.js 20 or newer, npm, Python 3, and pnpm. If pnpm is missing and Corepack is available, the installer attempts to activate pnpm automatically.

## Lifecycle commands

```bash
harnejr             # start daemon and open the web UI
harnejr doctor      # run daemon readiness check
harnejr update      # pull latest main branch and reinstall
harnejr stop        # stop daemon started by the Harnejr launcher
harnejr uninstall   # remove installed launcher and installed Harnejr files
harnejr version     # print installed metadata
```

`harnejr uninstall` removes the active installed launcher and install directory. It does not delete project workspaces or their `.harnejr` memory folders.

## Design goals

- Local web interface only. No TUI, Electron app, editor extension, or remote-hosted control plane.
- Go daemon owns local execution, filesystem access, policy decisions, workspace state, and safety gates.
- TypeScript web UI owns configuration, visibility, prompt editing, session controls, readiness views, model selection, commands, and provider editing.
- Provider configuration must be explicit, editable, and billing-path aware.
- Autonomy must remove unnecessary confirmation loops without bypassing hard safety rules.
- Completion claims must be supported by evidence, tests, logs, quality gates, subagent plans, or independent review.

## Built-in local systems

| System | Purpose |
| --- | --- |
| Harnejr Doctor | Readiness and configuration inspection. |
| LoC Controller | Scans source files and flags oversized files before completion. |
| Goal and Topic Controller | Stores scoped goal, topic, loop, and yolo state for a workspace session. |
| Autonomous Healer | Builds deterministic repair plans from doctor and quality findings. |
| Workspace Memory | Prepares Git state and `.harnejr` project memory. |
| Context Efficiency | Provides compact state packaging for efficient session continuation. |

## Runtime safety primitives

Current daemon-owned primitives include:

- safe workspace preparation with guarded Git initialization;
- `.harnejr` Markdown memory plus redacted `events.jsonl` ledger;
- real-path workspace boundary checks;
- safe file list/read/write APIs with secret-path denial;
- deterministic shell classifier;
- policy-gated shell runner that only executes `allow` decisions;
- provider registry validation and static/opt-in-live provider probes;
- deterministic subagent plan generation;
- evidence-based completion gate;
- global skill/agent discovery.

## Web control surface

The installed web UI includes:

- daemon readiness and Doctor status;
- built-in MCP/local system visibility;
- configured model selection from `configs/providers.default.json`;
- engineered command entry;
- session prompt entry;
- local transcript view;
- goal, topic, workspace, session ID, and yolo controls;
- permanent additive user-level system prompt editor.

Prompt submissions are currently stored into workspace memory through the daemon. Live provider execution is a roadmap item.

## User-level system prompt

The web UI includes a permanent user-level system prompt editor. The prompt is stored by the daemon and appended to Harnejr's fundamental harness prompt for every session. It does not replace core safety, policy, or runtime control instructions.

```text
GET /api/prompts/user
GET /api/prompts/composed
PUT /api/prompts/user
```

Stored file:

```text
<config-dir>/user.system.md
```

## Engineered command model

Harnejr commands are intended to be runtime state transitions, not prompt-only suggestions.

| Command | Intended behavior |
| --- | --- |
| `/goal` | Start an autonomous goal loop with completion review. It is mutually exclusive with `/loop`. |
| `/yolo` | Continue safe workspace work without ordinary confirmation prompts while keeping hard safety blocks active. |
| `/loop` | Run a fixed-iteration task loop. It is mutually exclusive with `/goal`. |
| `/swarm` | Spawn five bounded subagents from available providers/models while retaining main-session control. |
| `/export` | Write a JSONL session export into the active workspace with actions, errors, providers, token data, and evidence. |

Current daemon controls persist goal, topic, loop, and yolo state through `POST /api/control/apply`. The full autonomous execution loop is still being implemented.

## Workspace lifecycle

Every session prepares its workspace before agent work begins. Harnejr searches upward from the selected workspace for an existing local Git repository. If one exists, that repository root becomes the session project root. If no repository exists, Harnejr initializes one only when the selected folder is narrow enough to be treated as a project.

Harnejr refuses broad locations such as the filesystem root, the user's home folder, Desktop, Documents, Downloads, Pictures, Music, Videos, Public, Templates, and system-level folders. It also refuses to initialize a parent folder when child folders already contain Git repositories.

For safe project roots, Harnejr creates a hidden `.harnejr` directory containing Markdown memory files, scoped control state, and redacted JSONL events.

## Repository layout

```text
apps/web/                 Local web application
packages/shared/          TypeScript schemas shared by UI and sidecars
packages/provider-node/   Provider SDK sidecar scaffold
cmd/harnejrd/             Go daemon entrypoint
internal/agents/          Subagent planning primitives
internal/config/          Local config loading and defaults
internal/doctor/          Readiness report generator
internal/events/          Redacted event ledger
internal/healing/         Deterministic repair planner
internal/judge/           Completion evidence gate
internal/mcp/             Built-in local harness systems
internal/policy/          Shell/action policy primitives
internal/prompts/         Persistent user-level prompt storage
internal/providers/       Provider contracts, probes, and validation
internal/quality/         LoC and quality gate primitives
internal/server/          HTTP API served by the daemon
internal/session/         Goal, topic, yolo, and loop control state
internal/shell/           Policy-gated shell runner
internal/skills/          Global skill and agent discovery
internal/tools/           Built-in harness tool registry
internal/workspace/       Workspace path, Git, memory, files, and boundary logic
configs/                  Default provider, policy, agent, MCP, and skill configs
docs/                     Architecture and engineering notes
scripts/                  Development helpers
```

## Development

```bash
go test ./...
pnpm install
pnpm build
go run ./cmd/harnejrd --listen 127.0.0.1:8765 --web-dir apps/web/dist
pnpm --filter @harnejr/web dev
go build -o bin/harnejrd ./cmd/harnejrd
```

Query the daemon doctor:

```bash
scripts/doctor.sh
```

## Current daemon API

| Endpoint | Purpose |
| --- | --- |
| `GET /api/health` | Daemon health check. |
| `GET /api/config/defaults` | Load default provider, policy, agent, MCP, and skill config. |
| `GET /api/doctor` | Return readiness checks, built-in tools, and local systems. |
| `GET /api/tools` | List built-in harness tools. |
| `GET /api/mcp/systems` | List built-in local harness systems. |
| `GET /api/providers/probe` | Static provider readiness and auth-reference probe. |
| `POST /api/providers/probe` | Optional live provider probe when `{ "live": true }` is supplied. |
| `GET /api/prompts/user` | Read the permanent user-level system prompt. |
| `GET /api/prompts/composed` | Read the core prompt plus the saved user-level prompt. |
| `PUT /api/prompts/user` | Save the permanent user-level system prompt. |
| `POST /api/session/message` | Store a web prompt or command into workspace memory. |
| `POST /api/session/export` | Return the workspace event JSONL ledger. |
| `POST /api/control/apply` | Persist goal, topic, loop, and yolo state for a workspace session. |
| `POST /api/policy/classify-shell` | Classify a shell command as allow, ask, or deny. |
| `POST /api/shell/run` | Execute only policy-allowed shell commands in the prepared workspace. |
| `POST /api/workspaces/prepare` | Prepare a workspace by resolving Git state and local memory. |
| `POST /api/workspace/files/list` | List workspace files within path boundaries. |
| `POST /api/workspace/files/read` | Read non-secret workspace files within path boundaries. |
| `POST /api/workspace/files/write` | Write non-secret workspace files within path boundaries. |
| `POST /api/quality/loc` | Scan source line counts and flag oversized files. |
| `POST /api/healing/plan` | Build a repair plan from doctor and quality findings. |
| `POST /api/agents/plan` | Build a deterministic subagent plan for a task. |
| `POST /api/completion/check` | Block or accept completion based on evidence requirements. |
| `POST /api/skills/discover` | Discover global/workspace skills and agents. |

## Safety model

Harnejr treats prompts as guidance, not enforcement. Current implemented safety primitives include shell command classification, policy-gated shell execution, workspace path boundaries, symlink escape prevention, guarded Git initialization, project-local memory, scoped control state, safe file APIs, LoC quality gates, repair-plan output, provider validation, event redaction, and additive user prompt storage.

## Provider model

Harnejr models a provider as a transport contract, not just a base URL and model name. Provider profiles track protocol, runtime, billing mode, endpoint path, authentication mode, model namespace, reasoning adapter, stream parser, timeout, retry policy, aliases, OpenCode-compatible IDs, request defaults, model-specific defaults, and notes.

## Roadmap

Near-term work: SQLite-backed session history, structured patch application and file locks, live provider execution, provider fallback during active tasks, OpenAI-compatible and Ollama adapters, command dispatcher, provider-backed subagent scheduler, provider-backed judge loop, external MCP process handshakes, provider editor, logs, policy, and export screens.

## License

Harnejr is licensed under the MIT License. See `LICENSE` for details.
