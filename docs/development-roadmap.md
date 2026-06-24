# Development roadmap

## Step 1: Scaffold

Status: started.

- Repository layout.
- Go daemon.
- Web GUI shell.
- Shared schemas.
- Default configs.
- Safety and workspace tests.
- Install script.

## Step 2: State and sessions

- SQLite schema.
- Session creation and restoration.
- Workspace snapshot metadata.
- Goal/yolo/loop/swarm state machine.
- JSONL event writer.

## Step 3: Tool execution

- Read/list/search APIs.
- Workspace-safe file edit API.
- Shell runner with policy gate.
- Test command detection.
- Denied-action ledger.

## Step 4: Providers

- Provider profile validation.
- Health probes.
- OpenAI-compatible raw HTTP adapter.
- Ollama native adapter.
- Reasoning adapters.
- Streaming parsers.
- Fallback ledger.

## Step 5: Agents and judges

- Agent registry.
- Subagent scheduler.
- KAT plus StepFun pairing gate.
- Goal completion judge.
- Review evidence schema.

## Step 6: MCP and skills

- Skill scanner for `~/skills`, `~/.agents`, and `~/.codex/skills`.
- MCP config registry.
- MCP process lifecycle and handshake checks.
- Tool smoke tests.

## Step 7: Web GUI completion

- Provider editor.
- Session console.
- Policy viewer.
- Logs and export browser.
- Agent graph.
- Settings import/export.
