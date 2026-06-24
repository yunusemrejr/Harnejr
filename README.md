# Harnejr

Harnejr is an open-source, MIT-licensed, Ubuntu-native agentic coding harness with a local web interface. It is built around a Go daemon, a TypeScript web application, and editable provider profiles for modern LLM infrastructure.

Harnejr is not intended to be a prompt wrapper. The daemon is responsible for policy, workspace safety, provider routing, session state, subagent orchestration, logs, and completion evidence.

## Project status

Harnejr is in early scaffold development. The repository currently includes the daemon entrypoint, web application shell, provider/profile schemas, default configuration files, workspace preparation logic, a deterministic shell-policy classifier, path-boundary checks, tests, and installation scaffolding.

It is not production-ready yet. The next major work is persistent session state, provider adapters, real command dispatch, subagent scheduling, MCP and skills integration, and the full web control surface.

## Design goals

- Local web interface only. No TUI, Electron app, editor extension, or remote-hosted control plane.
- Go daemon owns local execution, filesystem access, policy decisions, workspace state, and safety gates.
- TypeScript web UI owns configuration, visibility, session controls, and provider editing.
- TypeScript sidecars may be used for fast-moving provider SDKs and stream normalization.
- Provider configuration must be explicit, editable, and billing-path aware.
- Autonomy must remove unnecessary confirmation loops without bypassing hard safety rules.
- Completion claims must be supported by evidence, tests, logs, or independent review.

## Architecture

```text
Browser web UI
  session controls
  provider editor
  policy viewer
  logs and exports
        |
        | HTTP / future streaming API
        v
harnejrd Go daemon
  workspace preparation
  shell and filesystem policy
  provider routing
  session state
  subagent scheduling
  completion review
        |
        | optional local sidecar boundary
        v
Node provider runtime
  SDK-backed provider experiments
  schema validation
  stream normalization
```

The browser must never call model providers directly. Provider calls go through the local daemon or an explicitly managed local sidecar.

## Engineered command model

Harnejr commands are intended to be runtime state transitions, not prompt-only suggestions.

| Command | Intended behavior |
| --- | --- |
| `/goal` | Start an autonomous goal loop with completion review. It is mutually exclusive with `/loop`. |
| `/yolo` | Continue safe workspace work without ordinary confirmation prompts while keeping hard safety blocks active. |
| `/loop` | Run a fixed-iteration task loop. It is mutually exclusive with `/goal`. |
| `/swarm` | Spawn five bounded subagents from available providers/models while retaining main-session control. |
| `/export` | Write a JSONL session export into the active workspace with actions, errors, providers, token data, and evidence. |

These commands are not fully implemented yet. The repository currently defines the intended behavior and lower-level primitives needed to build them correctly.

## Workspace lifecycle

Every session prepares its workspace before agent work begins.

Harnejr searches upward from the selected workspace for an existing local Git repository. If one exists, that repository root becomes the session project root. If no repository exists, Harnejr initializes one only when the selected folder is narrow enough to be treated as a project.

Harnejr refuses broad locations such as the filesystem root, the user's home folder, Desktop, Documents, Downloads, Pictures, Music, Videos, Public, Templates, and system-level folders. It also refuses to initialize a parent folder when child folders already contain Git repositories.

For safe project roots, Harnejr creates a hidden `.harnejr` directory containing compact Markdown memory files:

```text
.harnejr/
  README.md
  session-log.md
  requests.md
  decisions.md
  errors.md
  notices.md
```

These files are for future-session context: what was requested, what changed, why it changed, what failed, what was noticed, and what should be checked next.

## Repository layout

```text
apps/web/                 Local web application
packages/shared/          TypeScript schemas shared by UI and sidecars
packages/provider-node/   Provider SDK sidecar scaffold
cmd/harnejrd/             Go daemon entrypoint
internal/config/          Local config loading and defaults
internal/policy/          Shell/action policy primitives
internal/providers/       Provider contracts and routing types
internal/server/          HTTP API served by the daemon
internal/workspace/       Workspace path, Git, memory, and boundary logic
configs/                  Default provider, policy, agent, MCP, and skill configs
docs/                     Architecture and engineering notes
scripts/                  Development helpers
```

## Installation

Requirements:

- Ubuntu Linux
- Git
- Go 1.22 or newer
- Node.js 20 or newer
- pnpm 9 or newer for web development

Install from the repository root:

```bash
bash install.sh
harnejr
```

The installer builds the daemon, copies default configuration files, writes a `harnejr` launcher under `~/.local/bin`, starts the local daemon, and opens the web interface in the default browser.

## Development

Run Go tests:

```bash
go test ./...
```

Run the daemon locally:

```bash
go run ./cmd/harnejrd --listen 127.0.0.1:8765
```

Run the web UI in development mode:

```bash
pnpm install
pnpm --filter @harnejr/web dev
```

Build the daemon:

```bash
go build -o bin/harnejrd ./cmd/harnejrd
```

## Current daemon API

| Endpoint | Purpose |
| --- | --- |
| `GET /api/health` | Daemon health check. |
| `GET /api/config/defaults` | Load default provider, policy, agent, MCP, and skill config. |
| `POST /api/policy/classify-shell` | Classify a shell command as allow, ask, or deny. |
| `POST /api/workspaces/prepare` | Prepare a workspace by resolving Git state and local memory. |

## Safety model

Harnejr treats prompts as guidance, not enforcement. Safety-relevant decisions must be made by daemon code before execution.

Current implemented safety primitives include:

- deterministic shell command classification;
- workspace path boundary resolution;
- symlink escape prevention for workspace paths;
- guarded Git initialization for session workspaces;
- local Markdown memory that is scoped to safe project roots.

Future safety work will extend this into file-edit policy, command execution policy, provider policy, MCP policy, and goal-completion policy.

## Provider model

Harnejr models a provider as a transport contract, not just a base URL and model name. Provider profiles track protocol, runtime, billing mode, base URL, endpoint path, authentication mode, model namespace, reasoning adapter, stream parser, timeout, retry policy, and notes.

Default provider profiles are stored in `configs/providers.default.json` and are intended to be editable through the web UI once provider management is implemented.

## Roadmap

Near-term work:

- persistent SQLite-backed session state;
- workspace read/write/edit APIs;
- shell runner behind policy gates;
- provider health probes and OpenAI-compatible adapter;
- Ollama native adapter;
- real command dispatcher for `/goal`, `/yolo`, `/loop`, `/swarm`, and `/export`;
- subagent scheduler and judge loop;
- MCP and skills discovery;
- web UI screens for providers, sessions, logs, policy, and exports.

## License

Harnejr is licensed under the MIT License. See `LICENSE` for details.
