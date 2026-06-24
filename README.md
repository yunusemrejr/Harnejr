# Harnejr

Harnejr is an open-source, MIT-licensed, Ubuntu-native agentic coding harness with a local web GUI. It is designed as a real harness, not a terminal prompt wrapper: the daemon owns policy, provider routing, workspace safety, subagents, judges, MCP and skill discovery, compaction, logs, and completion proof.

The current repository is an initial scaffold. It establishes the architecture and safe defaults needed before implementation grows into autonomous coding behavior.

## Non-negotiable shape

- Web GUI only. No TUI, no Electron, no editor extension dependency.
- Go daemon for local execution, policy, state, and filesystem safety.
- TypeScript web frontend for configuration, session control, logs, and provider editing.
- TypeScript provider sidecar for fast-moving SDK integrations where Go-native clients are not enough.
- Ubuntu-first install path through `install.sh`, ending with a `harnejr` command that opens the local web UI.
- Provider configs are editable and secret-safe. API keys are stored locally by the user and referenced through environment variables or future secret refs, never committed.
- Autonomous modes must deny unsafe actions and continue with safe alternatives instead of asking the user for confirmation.

## Planned engineered commands

Harnejr's web session command layer will support these runtime commands as deterministic harness actions, not as prompt-only suggestions:

| Command | Required behavior |
| --- | --- |
| `/goal` | Starts a judged autonomous goal loop. It cannot run at the same time as `/loop`. Completion requires evidence and judge approval. |
| `/yolo` | Removes ordinary confirmation prompts while keeping hard safety blocks active. It is compatible with all other modes. |
| `/loop` | Runs a task for a fixed number of iterations. It cannot run while a goal is active. |
| `/swarm` | Spawns five bounded subagents from available providers/models while the main agent retains control. |
| `/export` | Writes a full JSONL session export into the active workspace, including harness actions, model usage, token accounting, errors, and verification evidence. |

## Repository layout

```text
apps/web/                 Local web GUI
packages/shared/          TypeScript schemas shared by the UI and Node sidecars
packages/provider-node/   Provider SDK sidecar scaffold
cmd/harnejrd/             Go daemon entrypoint
internal/config/          Local config loading and defaults
internal/policy/          Deterministic shell/action policy
internal/providers/       Provider contracts, capability metadata, and routing types
internal/server/          HTTP API served by the daemon
internal/workspace/       Workspace path resolution and escape prevention
configs/                  Default provider, policy, agent, MCP, and skill config
docs/                     Architecture and engineering notes
scripts/                  Development helpers
```

## Development

Requirements on Ubuntu:

- Go 1.22+
- Node.js 20+
- pnpm 9+

Run the daemon tests:

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

## Install flow

The `install.sh` script is intentionally conservative in this scaffold. It checks dependencies, builds the Go daemon, installs files under `~/.local/share/harnejr`, writes a launcher to `~/.local/bin/harnejr`, and does not install or expose provider keys.

```bash
./install.sh
harnejr
```

## Safety baseline

The first implemented policy surface is deterministic shell classification. It denies commands that can damage the OS, mutate privileged state, expose secrets, or perform risky remote/deployment operations. The model layer will not be trusted to remember these rules; the daemon must enforce them before execution.
