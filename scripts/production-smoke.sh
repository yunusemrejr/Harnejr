#!/usr/bin/env bash
set -euo pipefail

listen="${HARNEJR_LISTEN:-127.0.0.1:8765}"
url="http://${listen}"
config_dir="${HARNEJR_CONFIG_DIR:-configs}"
web_dir="${HARNEJR_WEB_DIR:-apps/web/dist}"
workspace="$(mktemp -d)"
cleanup() {
  if [ -n "${pid:-}" ]; then
    kill "$pid" >/dev/null 2>&1 || true
  fi
  rm -rf "$workspace"
}
trap cleanup EXIT

mkdir -p bin
go test ./...
go build -o ./bin/harnejrd ./cmd/harnejrd
./bin/harnejrd --listen "$listen" --config-dir "$config_dir" --web-dir "$web_dir" >/tmp/harnejr-smoke.log 2>&1 &
pid=$!

for _ in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS "$url/api/health" >/dev/null; then
    break
  fi
  sleep 0.2
done

curl -fsS "$url/api/doctor" >/dev/null
curl -fsS "$url/api/providers/probe" >/dev/null
curl -fsS "$url/api/providers/registry" | grep -q 'providers'
curl -fsS -X POST "$url/api/providers/route" -H 'content-type: application/json' -d '{"taskType":"review patch","needsReasoning":true,"needsTools":true,"preferred":["stepfun-ai"],"maxCostClass":"cheap"}' | grep -q 'stepfun-step-plan'
curl -fsS "$url/api/mcp/check" >/dev/null
curl -fsS -X POST "$url/api/policy/classify-shell" -H 'content-type: application/json' -d '{"command":"sudo rm -rf /"}' | grep -q 'deny'
curl -fsS -X POST "$url/api/workspaces/prepare" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"userRequest\":\"smoke\"}" >/dev/null
curl -fsS -X POST "$url/api/goals/start" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"goal\":\"ship smoke goal\"}" | grep -q 'checkpoints'
curl -fsS -X POST "$url/api/goals/checkpoint" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"checkpointId\":\"scope\",\"status\":\"done\",\"notes\":\"smoke\"}" >/dev/null
curl -fsS -X POST "$url/api/workspace/files/write" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"path\":\"notes.md\",\"content\":\"ok\"}" | grep -q 'backup'
curl -fsS -X POST "$url/api/workspace/files/patch" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"path\":\"notes.md\",\"oldText\":\"ok\",\"newText\":\"patched\"}" | grep -q 'snapshot'
curl -fsS -X POST "$url/api/memory/summary" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\"}" | grep -q 'summary'
curl -fsS -X POST "$url/api/shell/run" -H 'content-type: application/json' -d "{\"workspaceRoot\":\"$workspace\",\"sessionId\":\"smoke\",\"command\":\"pwd\"}" >/dev/null
curl -fsS -X POST "$url/api/agents/plan" -H 'content-type: application/json' -d '{"task":"implement production provider routing fix","mode":"goal","requestedModel":"kat-coder-pro-v2"}' | grep -q 'stepfun-step-plan'
curl -fsS -X POST "$url/api/completion/check" -H 'content-type: application/json' -d '{"goal":"ship","evidence":[],"tests":[],"subagentReviews":0,"qualityGatePass":false,"providerPlanPass":false}' | grep -q 'accepted":false'

echo "Harnejr production smoke passed."
