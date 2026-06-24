# Provider settings UI

Harnejr exposes provider setup as a first-class product surface instead of requiring users to edit JSON by hand.

The web UI has a dedicated Providers screen for:

- choosing a provider profile;
- enabling or disabling that profile;
- editing display name, base URL, endpoint, protocol, billing mode, auth mode, auth header, default model, and environment variable name;
- viewing configured models and context/output metadata;
- saving a local credential through the daemon.

Credentials are not stored in browser storage. The browser sends the value once to the local daemon. The daemon writes it under the active config directory and stores only the resulting local file path in the provider registry.

Relevant endpoints:

```text
GET /api/providers/registry
PUT /api/providers/registry
PUT /api/providers/secret
GET /api/providers/probe
POST /api/providers/probe
```

Runtime provider calls use the saved local credential file before falling back to environment variables.

The chat composer must remain visually primary: a message text area and one primary Send button. Provider setup, diagnostics, workers, review, yolo, and goal controls must live in separate panels so they do not compete with the main send action.
