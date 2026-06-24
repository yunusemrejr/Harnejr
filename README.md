# Harnejr

Harnejr is an open-source, MIT-licensed, Ubuntu-native agentic coding harness with a local web interface. It is built around a Go daemon, a TypeScript web application, editable provider profiles, workspace safety primitives, built-in local harness systems, and project-local memory.

Harnejr is not a prompt wrapper. The daemon owns policy, workspace preparation, persistent controls, quality checks, readiness reporting, provider contracts, logs, and completion evidence.

## Project status

Harnejr is in active scaffold development. The repository includes a runnable daemon, installable web control surface, default provider registry, policy config, workspace preparation, local memory, shell classification, path-boundary checks, built-in local harness systems, doctor reporting, LoC quality scanning, repair planning, scoped goal/topic controls, a session prompt console, configured model selector, command entry, and a permanent user-level system prompt editor.

The project is ready for serious harness development and local control-surface testing. Full autonomous coding execution, live provider calls, subagent execution, external MCP handshakes, and judge enforcement are still under implementation.

## Design goals

- Local web interface only. No TUI, Electron app, editor extension, or remote-hosted control plane.
- Go daemon owns local execution, filesystem access, policy decisions, workspace state, and safety gates.
- TypeScript web UI owns configuration, visibility, prompt editing, session controls, readiness views, model selection, commands, and provider editing.
- Provider configuration must be explicit, editable, and billing-path aware.
- Autonomy must remove unnecessary confirmation loops without bypassing hard safety rules.
- Completion claims must be supported by evidence, tests, logs, quality gates, or independent review.

## Built-in local systems

| System | Purpose |
| --- | --- |
| Harnejr Doctor | Read-only readiness and configuration inspection. |
| LoC Controller | Scans source files and flags oversized files before completion. |
| Goal and Topic Controller | Stores scoped goal, topic, loop, and yolo state for a workspace session. |
| Autonomous Healer | Builds deterministic repair plans from doctor and quality findings. |
| Workspace Memory | Prepares Git state and `.harnejr` project memory. |
| Context Efficiency | Provides compact state packaging for efficient session continuation. |

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

For safe project roots, Harnejr creates a hidden `.harnejr` directory containing Markdown memory files and scoped control state.

## Repository layout

```text
apps/web/                 Local web application
packages/shared/          TypeScript schemas shared by UI and sidecars
packages/provider-node/   Provider SDK sidecar scaffold
cmd/harnejrd/             Go daemon entrypoint
internal/config/          Local config loading and defaults
internal/doctor/          Readiness report generator
internal/healing/         Deterministic repair planner
internal/mcp/             Built-in local harness systems
internal/policy/          Shell/action policy primitives
internal/prompts/         Persistent user-level prompt storage
internal/providers/       Provider contracts and routing types
internal/quality/         LoC and quality gate primitives
internal/server/          HTTP API served by the daemon
internal/session/         Goal, topic, yolo, and loop control state
internal/tools/           Built-in harness tool registry
internal/workspace/       Workspace path, Git, memory, and boundary logic
configs/                  Default provider, policy, agent, MCP, and skill configs
docs/                     Architecture and engineering notes
scripts/                  Development helpers
```

## Installation

Requirements: Ubuntu Linux, Git, Go 1.22 or newer, Node.js 20 or newer, npm, and pnpm.

```bash
bash install.sh
harnejr
```

The installer runs Go tests, installs pnpm workspace dependencies, builds the web UI, builds the daemon, copies `apps/web/dist` into the install directory, copies default configs, and writes the `harnejr` launcher. The launcher starts the daemon with the installed web UI directory so the browser opens the full React control surface rather than the daemon fallback page.

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
| `GET /api/prompts/user` | Read the permanent user-level system prompt. |
| `GET /api/prompts/composed` | Read the core prompt plus the saved user-level prompt. |
| `PUT /api/prompts/user` | Save the permanent user-level system prompt. |
| `POST /api/session/message` | Store a web prompt or command into workspace memory. |
| `POST /api/control/apply` | Persist goal, topic, loop, and yolo state for a workspace session. |
| `POST /api/policy/classify-shell` | Classify a shell command as allow, ask, or deny. |
| `POST /api/workspaces/prepare` | Prepare a workspace by resolving Git state and local memory. |
| `POST /api/quality/loc` | Scan source line counts and flag oversized files. |
| `POST /api/healing/plan` | Build a repair plan from doctor and quality findings. |

## Safety model

Harnejr treats prompts as guidance, not enforcement. Current implemented safety primitives include shell command classification, workspace path boundaries, symlink escape prevention, guarded Git initialization, project-local memory, scoped control state, LoC quality gates, repair-plan output, and additive user prompt storage.

## Provider model

Harnejr models a provider as a transport contract, not just a base URL and model name. Provider profiles track protocol, runtime, billing mode, endpoint path, authentication mode, model namespace, reasoning adapter, stream parser, timeout, retry policy, and notes.

## Roadmap

Near-term work: SQLite-backed session history, workspace edit APIs, policy-gated shell execution, live provider execution, provider health probes, OpenAI-compatible and Ollama adapters, command dispatcher, subagent scheduler, judge loop, external MCP process handshakes, skills discovery, provider editor, logs, policy, and export screens.

## License

Harnejr is licensed under the MIT License. See `LICENSE` for details.
