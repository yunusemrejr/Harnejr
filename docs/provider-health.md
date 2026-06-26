# Provider health ledger

Harnejr now stores provider call health as daemon state instead of treating every fallback decision as a fresh guess.

## State file

The daemon writes provider health under:

```text
<config-dir>/state/provider-health.json
```

The file records:

- provider id;
- health state;
- success and failure counters;
- latest error class;
- latest HTTP status;
- latest latency;
- cooldown expiry when a provider should be skipped temporarily.

## Runtime behavior

`/api/llm/generate` now uses the health-aware generation path. Before trying a provider, the daemon checks whether the provider is marked down or is still inside a cooldown window. Skipped providers are included in the `healthSkipped` field of the generation result.

A successful provider response resets the failure count and cooldown. Auth and endpoint/model errors mark the provider down until health is reset. Temporary provider failures can move a provider into cooldown instead of repeatedly burning calls.

## Endpoints

```text
GET  /api/providers/health
POST /api/providers/health/reset
```

Reset request body:

```json
{ "providerId": "stepfun-step-plan" }
```

Omit `providerId` to clear all provider health state.

## Why this matters

Harnejr is supposed to spare the user from manually switching providers after quota, endpoint, auth, and transient provider errors. A persisted health ledger gives the daemon evidence for routing and fallback decisions instead of relying on prompt reminders or UI labels.
