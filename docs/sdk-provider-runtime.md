# SDK provider runtime

Harnejr provider execution has two backend paths: Go raw HTTP and an optional Node SDK sidecar. The Go daemon remains the authority for sessions, workspace safety, shell policy, file mutation, provider routing, billing fallback, and completion evidence.

The sidecar package `packages/provider-node` is a narrow provider adapter. It uses Vercel AI SDK, OpenAI-compatible provider adapters, Anthropic-compatible provider adapters, and Zod validation. It does not own permissions, memory, goals, tools, or files.

Provider profiles select this path with `runtime: "node_ai_sdk"`. If a profile asks for an SDK runtime and `HARNEJR_PROVIDER_NODE_PATH` is not set, the daemon must return an explicit SDK runtime error instead of pretending the provider worked.

Current installed SDK packages are `ai`, `@ai-sdk/openai-compatible`, `@ai-sdk/anthropic`, and `zod`. The Go daemon can still use raw HTTP for OpenAI Responses-shaped providers until a provider profile is explicitly switched to a sidecar runtime.
