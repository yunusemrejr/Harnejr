# Safety policy

Harnejr autonomy is deny-and-continue, not ask-and-wait.

When `/yolo` or `/goal` is active, ordinary safe workspace actions should proceed without user prompts. Unsafe actions must be blocked by the daemon and the agent must choose a reversible local alternative.

## Hard blocks

The first shell classifier denies privilege escalation, destructive filesystem commands, raw device writes, shutdown commands, service mutation, firewall mutation, unsafe recursive permission changes, force-push flows, remote script piping, likely secret reads, and risky remote mutation commands.

## Workspace boundary

The workspace path engine resolves real paths and rejects traversal and symlink escapes. Future edit/write APIs must call this engine before touching files.

## Policy principle

Prompts can remind a model about safety. They are not enforcement. The daemon must make the final allow, deny, or ask decision before execution.
