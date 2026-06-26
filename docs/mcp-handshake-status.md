# MCP handshake status

Harnejr exposes two MCP readiness surfaces:

```text
GET  /api/mcp/check
POST /api/mcp/handshake
```

`/api/mcp/check` reads the configured MCP registry and verifies static readiness such as command presence and required environment variables.

`/api/mcp/handshake` is intentionally explicit about what it has and has not proven. At this stage it returns the same command and environment readiness plus:

```json
{
  "handshakeAttempted": false,
  "reason": "external MCP process handshake requires the daemon stdio probe implementation"
}
```

This is not a cosmetic endpoint. It prevents Harnejr from claiming `initialize`, `tools/list`, or tool-smoke success before the daemon has a safe stdio probe implementation in place.

## Production rule

MCP status must never be reported as healthy from configuration alone. The final production state still requires:

1. process start proof;
2. MCP initialize response proof;
3. `tools/list` proof;
4. optional selected tool smoke proof;
5. redacted event-ledger evidence for each stage.
