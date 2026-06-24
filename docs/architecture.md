# Architecture

Harnejr is split into three layers: the model brain, the harness brain, and the human-facing web GUI.

The model brain proposes plans, edits, tests, and review findings. It does not own permission decisions, provider fallback, completion acceptance, or workspace boundaries.

The harness brain is the Go daemon. It owns local state, HTTP APIs, policy, workspace path resolution, shell execution gates, provider routing, subagent scheduling, judge activation, compaction, MCP and skills registries, and evidence exports.

The web GUI is the control surface. It edits provider profiles, starts sessions, displays logs, shows policy decisions, and exposes goal/yolo/loop/swarm/export controls. It must never store raw provider keys in browser storage.

## Process layout

```text
Browser web GUI
  React/Vite app
  session control
  provider editor
  policy viewer
  logs and exports
        |
        | HTTP and future WebSocket/SSE
        v
harnejrd Go daemon
  SQLite state, planned
  workspace path engine
  policy engine
  provider router
  subagent scheduler
  judge/completion gate
  MCP/skills manager
        |
        | internal HTTP/stdin JSON-RPC, planned
        v
Node provider sidecar
  SDK experiments
  stream normalization
  provider smoke tests
```

## First scaffold boundary

The current scaffold implements:

- Go daemon entrypoint.
- Health/config/policy HTTP endpoints.
- Default provider, policy, agent, MCP, and skill configs.
- Shell policy classifier with tests.
- Workspace path escape prevention with tests.
- Minimal React/Vite UI shell.
- TypeScript provider/session schemas.
- Conservative Ubuntu install script.

The next boundary is persistent state and real session execution.
