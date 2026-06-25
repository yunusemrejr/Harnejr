# Web UI remake

Date: 2026-06-25

This pass replaces the previous panel shell with a purpose-driven local control surface. The interface is dark, gray-and-cream only, and built around daemon-backed work instead of decorative status language.

## Design rules

- No decorative gradients, glows, emojis, blinking indicators, or fake status badges.
- No browser-side secret storage. API keys are submitted to the daemon credential endpoint.
- All configuration belongs where it is used: provider contracts in Providers, workspace and goals in Runtime, additive prompt in Prompts, runtime proof in Diagnostics, and task execution in Session.
- The selected provider route is always visible as a contract: provider ID, model, protocol, and billing mode.
- The main composer has one primary action. Worker passes, review, memory, and goal controls remain explicit secondary actions.

## Functional layout

### Session

The Session surface is the normal working area. It contains the transcript, selected model route, billing fallback control, task composer, memory save action, worker action, review action, and Send.

### Providers

The Providers surface is now a full editor for inference contracts. It supports adding providers, duplicating providers, removing providers, editing IDs, editing endpoint and protocol fields, saving local API keys through the daemon, editing auth metadata, editing default models, and maintaining the model roster.

### Runtime

The Runtime surface owns workspace root, session ID, topic, yolo state, checkpointed goal text, workspace preparation, control persistence, goal start, compact memory, worker pass, and completion review.

### Diagnostics

Diagnostics runs daemon proof checks rather than relying on interface labels. It exposes Doctor, provider probe, MCP check, and skills/agents discovery.

### Prompts

Prompts contains only the user-level additive prompt. It explicitly stays below daemon policy and cannot override safety, provider routing, or workspace boundaries.

## Implementation notes

- `apps/web/src/App.tsx` was remade around five task-oriented surfaces: Session, Providers, Runtime, Diagnostics, and Prompts.
- `apps/web/src/styles.css` was replaced with a dark gray/cream system, responsive three-column layout, and form patterns for provider and model editing.
- The UI continues to use the existing daemon APIs for provider registry, credential saving, provider calls, workspace preparation, checkpointed goals, compact memory, provider probes, MCP checks, skill discovery, and user prompts.
