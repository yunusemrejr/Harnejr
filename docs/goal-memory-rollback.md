# Goals, memory, and rollback

Harnejr now treats goals, memory, and file mutation as daemon-controlled runtime surfaces.

## Checkpointed goals

A goal is persisted into the active workspace memory as `.harnejr/goal.json`. Starting a goal creates deterministic checkpoints:

1. understand scope;
2. plan reversible steps;
3. apply bounded changes;
4. run verification;
5. perform independent review;
6. make a completion decision.

Endpoints:

```text
POST /api/goals/start
POST /api/goals/status
POST /api/goals/checkpoint
```

The goal is not complete merely because the model says it is complete. The completion gate still requires evidence.

## Rollback snapshots

Before file-write and text-patch mutations, Harnejr creates a snapshot under:

```text
.harnejr/rollback/<timestamp>/
```

Each snapshot writes a `manifest.json` and, when the target file already exists, a `.before` backup of the previous file content. This makes ordinary file mutations audit-friendly and reversible.

## Lean memory summaries

Harnejr can build a compact memory file at:

```text
.harnejr/summary.md
```

Endpoint:

```text
POST /api/memory/summary
```

The summary keeps recent useful workspace memory from request logs, decisions, notices, goal state, and control state. It is not raw chat history. It exists to prevent long-session bloat while preserving what matters.
