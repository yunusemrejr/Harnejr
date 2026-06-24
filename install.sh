#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
install_root="${HARNEJR_INSTALL_ROOT:-$HOME/.local/share/harnejr}"
bin_dir="${HARNEJR_BIN_DIR:-$HOME/.local/bin}"
launcher="$bin_dir/harnejr"
listen_addr="${HARNEJR_LISTEN:-127.0.0.1:8765}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing dependency: $1" >&2
    exit 1
  fi
}

need go
need node
need npm

mkdir -p "$install_root/bin" "$install_root/configs" "$install_root/web" "$bin_dir"

(
  cd "$repo_root"
  go test ./...
  go build -o "$install_root/bin/harnejrd" ./cmd/harnejrd
)

cp -R "$repo_root/configs/." "$install_root/configs/"

cat > "$launcher" <<LAUNCHER
#!/usr/bin/env bash
set -euo pipefail
listen="\${HARNEJR_LISTEN:-$listen_addr}"
url="http://\${listen}"
"$install_root/bin/harnejrd" --listen "\$listen" --config-dir "$install_root/configs" >/tmp/harnejr-daemon.log 2>&1 &
daemon_pid=\$!
for _ in 1 2 3 4 5 6 7 8 9 10; do
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "\$url" >/dev/null 2>&1 || true
    break
  fi
  sleep 0.2
done
printf 'Harnejr daemon started on %s with pid %s\n' "\$url" "\$daemon_pid"
printf 'Daemon log: /tmp/harnejr-daemon.log\n'
wait "\$daemon_pid"
LAUNCHER

chmod +x "$launcher"

echo "Installed Harnejr daemon to $install_root"
echo "Installed launcher to $launcher"
echo "Run: harnejr"
