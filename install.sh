#!/usr/bin/env bash
set -euo pipefail

repo_url="${HARNEJR_REPO_URL:-https://github.com/yunusemrejr/Harnejr.git}"
branch="${HARNEJR_BRANCH:-main}"
source_default="${HARNEJR_SOURCE_DIR:-$HOME/.local/src/harnejr}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing dependency: $1" >&2
    exit 1
  fi
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" >/dev/null 2>&1 && pwd || pwd)"
if [ ! -f "$script_dir/go.mod" ] || [ ! -d "$script_dir/apps/web" ]; then
  need git
  mkdir -p "$(dirname "$source_default")"
  if [ -d "$source_default/.git" ]; then
    git -C "$source_default" fetch --prune origin
    git -C "$source_default" checkout "$branch"
    git -C "$source_default" pull --ff-only origin "$branch"
  else
    if [ -e "$source_default" ]; then
      mv "$source_default" "${source_default}.backup.$(date +%Y%m%d%H%M%S)"
    fi
    git clone --branch "$branch" "$repo_url" "$source_default"
  fi
  exec bash "$source_default/install.sh" "$@"
fi

repo_root="$script_dir"
install_root="${HARNEJR_INSTALL_ROOT:-$HOME/.local/share/harnejr}"
bin_dir="${HARNEJR_BIN_DIR:-$HOME/.local/bin}"
launcher="$bin_dir/harnejr"
listen_addr="${HARNEJR_LISTEN:-127.0.0.1:8765}"

need git
need curl
need go
need node
need npm
need python3

if ! command -v pnpm >/dev/null 2>&1; then
  if command -v corepack >/dev/null 2>&1; then
    corepack enable
    corepack prepare pnpm@latest --activate
  fi
fi
need pnpm

mkdir -p "$install_root/bin" "$install_root/configs" "$install_root/web" "$bin_dir"

(
  cd "$repo_root"
  go test ./...
  pnpm install
  pnpm build
  go build -o "$install_root/bin/harnejrd" ./cmd/harnejrd
)

python3 - "$install_root/web" <<'PY'
import pathlib
import shutil
import sys
path = pathlib.Path(sys.argv[1]).expanduser().resolve()
home = pathlib.Path.home().resolve()
if path == pathlib.Path('/') or path == home or path == home.parent:
    raise SystemExit(f"refusing unsafe web directory cleanup: {path}")
if path.exists():
    shutil.rmtree(path)
path.mkdir(parents=True, exist_ok=True)
PY

cp -R "$repo_root/apps/web/dist/." "$install_root/web/"
cp -R "$repo_root/configs/." "$install_root/configs/"
commit="unknown"
if git -C "$repo_root" rev-parse --short HEAD >/dev/null 2>&1; then
  commit="$(git -C "$repo_root" rev-parse --short HEAD)"
fi
cat > "$install_root/install.json" <<META
{
  "sourceDir": "$repo_root",
  "repoUrl": "$repo_url",
  "branch": "$branch",
  "commit": "$commit",
  "installedAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
META

cat > "$launcher" <<LAUNCHER
#!/usr/bin/env bash
set -euo pipefail
install_root="$install_root"
source_dir="$repo_root"
repo_url="$repo_url"
branch="$branch"
listen="\${HARNEJR_LISTEN:-$listen_addr}"
url="http://\${listen}"
bin="\$install_root/bin/harnejrd"
config_dir="\$install_root/configs"
web_dir="\$install_root/web"
pid_file="\$install_root/harnejrd.pid"
log_file="\${HARNEJR_LOG_FILE:-\$install_root/harnejrd.log}"
launcher_path="$launcher"

health() {
  command -v curl >/dev/null 2>&1 && curl -fsS "\$url/api/health" >/dev/null 2>&1
}

open_ui() {
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "\$url" >/dev/null 2>&1 || true
  fi
}

start_daemon() {
  if health; then
    open_ui
    printf 'Harnejr is already running at %s\n' "\$url"
    return 0
  fi
  mkdir -p "\$install_root"
  "\$bin" --listen "\$listen" --config-dir "\$config_dir" --web-dir "\$web_dir" >>"\$log_file" 2>&1 &
  daemon_pid=\$!
  printf '%s\n' "\$daemon_pid" > "\$pid_file"
  for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
    if health; then
      open_ui
      printf 'Harnejr daemon started on %s with pid %s\n' "\$url" "\$daemon_pid"
      printf 'Daemon log: %s\n' "\$log_file"
      return 0
    fi
    sleep 0.2
  done
  printf 'Harnejr daemon did not become ready. Log: %s\n' "\$log_file" >&2
  return 1
}

stop_daemon() {
  if [ -f "\$pid_file" ]; then
    pid="\$(cat "\$pid_file" 2>/dev/null || true)"
    if [ -n "\$pid" ] && kill -0 "\$pid" >/dev/null 2>&1; then
      kill "\$pid" >/dev/null 2>&1 || true
      for _ in 1 2 3 4 5; do
        kill -0 "\$pid" >/dev/null 2>&1 || break
        sleep 0.2
      done
    fi
    rm -f "\$pid_file"
  fi
  printf 'Harnejr daemon stopped if it was running.\n'
}

doctor() {
  started=0
  if ! health; then
    "\$bin" --listen "\$listen" --config-dir "\$config_dir" --web-dir "\$web_dir" >>"\$log_file" 2>&1 &
    temp_pid=\$!
    started=1
    for _ in 1 2 3 4 5 6 7 8 9 10; do
      health && break
      sleep 0.2
    done
  fi
  curl -fsS "\$url/api/doctor"
  printf '\n'
  if [ "\$started" = "1" ]; then
    kill "\$temp_pid" >/dev/null 2>&1 || true
  fi
}

update_harnejr() {
  was_running=0
  if health; then
    was_running=1
    stop_daemon
  fi
  if [ ! -d "\$source_dir/.git" ]; then
    mkdir -p "\${source_dir%/*}"
    git clone --branch "\$branch" "\$repo_url" "\$source_dir"
  fi
  git -C "\$source_dir" fetch --prune origin
  git -C "\$source_dir" checkout "\$branch"
  git -C "\$source_dir" pull --ff-only origin "\$branch"
  HARNEJR_INSTALL_ROOT="\$install_root" HARNEJR_BIN_DIR="\${launcher_path%/*}" HARNEJR_LISTEN="\$listen" bash "\$source_dir/install.sh"
  if [ "\$was_running" = "1" ]; then
    exec "\$launcher_path" start
  fi
  printf 'Harnejr updated. Run: harnejr\n'
}

uninstall_harnejr() {
  stop_daemon || true
  python3 - "\$install_root" "\$launcher_path" <<'PY'
import pathlib
import shutil
import sys
install_root = pathlib.Path(sys.argv[1]).expanduser().resolve()
launcher = pathlib.Path(sys.argv[2]).expanduser().resolve()
home = pathlib.Path.home().resolve()
expected = home / '.local' / 'share' / 'harnejr'
if install_root != expected and 'HARNEJR_ALLOW_CUSTOM_UNINSTALL' not in __import__('os').environ:
    raise SystemExit(f"refusing to remove custom install root without HARNEJR_ALLOW_CUSTOM_UNINSTALL=1: {install_root}")
if install_root.exists():
    shutil.rmtree(install_root)
if launcher.exists():
    launcher.unlink()
PY
  printf 'Harnejr uninstalled. Project workspaces and their .harnejr memory folders were not removed.\n'
}

case "\${1:-start}" in
  start|open|run) start_daemon ;;
  stop) stop_daemon ;;
  doctor|status) doctor ;;
  update) update_harnejr ;;
  uninstall) uninstall_harnejr ;;
  version) cat "\$install_root/install.json" ; printf '\n' ;;
  help|-h|--help)
    cat <<'HELP'
Harnejr commands:
  harnejr             start daemon and open web UI
  harnejr start       start daemon and open web UI
  harnejr stop        stop daemon started by Harnejr launcher
  harnejr doctor      run daemon readiness check
  harnejr update      pull latest main branch and reinstall
  harnejr uninstall   remove installed launcher and installed Harnejr files
  harnejr version     print installed metadata
HELP
    ;;
  *)
    printf 'Unknown Harnejr command: %s\n' "\$1" >&2
    printf 'Run: harnejr help\n' >&2
    exit 2
    ;;
esac
LAUNCHER

chmod +x "$launcher"

echo "Installed Harnejr daemon to $install_root"
echo "Installed web UI to $install_root/web"
echo "Installed launcher to $launcher"
echo "Run: harnejr"
echo "Lifecycle commands: harnejr update | harnejr doctor | harnejr uninstall"
