#!/usr/bin/env bash
set -euo pipefail

listen="${HARNEJR_LISTEN:-127.0.0.1:8765}"
url="http://${listen}/api/doctor"
config_dir="${HARNEJR_CONFIG_DIR:-configs}"
web_dir="${HARNEJR_WEB_DIR:-apps/web/dist}"

if ! command -v curl >/dev/null 2>&1; then
  printf 'curl is required to query %s\n' "$url" >&2
  exit 1
fi

if curl -fsS "$url"; then
  printf '\n'
  exit 0
fi

printf 'No running Harnejr daemon at %s. Starting temporary daemon for doctor check.\n' "$url" >&2
if [ ! -x ./bin/harnejrd ]; then
  go build -o ./bin/harnejrd ./cmd/harnejrd
fi
./bin/harnejrd --listen "$listen" --config-dir "$config_dir" --web-dir "$web_dir" >/tmp/harnejr-doctor-daemon.log 2>&1 &
pid=$!
trap 'kill "$pid" >/dev/null 2>&1 || true' EXIT
for _ in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS "$url"; then
    printf '\n'
    exit 0
  fi
  sleep 0.2
done
printf 'Harnejr doctor failed. Temporary daemon log: /tmp/harnejr-doctor-daemon.log\n' >&2
exit 1
