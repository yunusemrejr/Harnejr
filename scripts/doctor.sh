#!/usr/bin/env bash
set -euo pipefail

listen="${HARNEJR_LISTEN:-127.0.0.1:8765}"
url="http://${listen}/api/doctor"

if command -v curl >/dev/null 2>&1; then
  curl -fsS "$url"
  printf '\n'
else
  printf 'curl is required to query %s\n' "$url" >&2
  exit 1
fi
