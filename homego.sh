#!/usr/bin/env bash

# Start only Homepage and its dedicated Cloudflare Tunnel on this Intel Mac.
set -Eeuo pipefail
HOME=/Users/richardmauri
APP_DIR="$HOME/projects/homepage"
APP="$APP_DIR/dist/homepage-darwin-amd64"
export SUITE_DATA_DIR="$APP_DIR/private"
TOKEN="$APP_DIR/.cloudflared/token"
CLOUDFLARED="$HOME/Applications/cloudflared"
cd "$APP_DIR"
for file in "$APP" "$TOKEN" "$CLOUDFLARED"; do
  [[ -f "$file" ]] || { echo "Missing required file: $file" >&2; exit 1; }
done
[[ -x "$APP" && -x "$CLOUDFLARED" ]] || { echo 'Executable missing permissions' >&2; exit 1; }
version_output="$("$APP" -version)"
[[ "$version_output" =~ ^Homepage\ v[0-9]+\.[0-9]+\.[0-9]+\ ·\ Build\ [0-9]{6}$ ]] || {
  echo "Invalid Homepage version output: $version_output" >&2; exit 1;
}
[[ "$version_output" != *"Build 000000" ]] || { echo 'Unstamped build refused' >&2; exit 1; }
echo "Starting $version_output with arguments: ${*:-default settings}"
# Stop only exact Homepage dist executable and this tunnel's token-file process.
ps -axo pid=,command= | while read -r pid command; do
  if [[ "$command" == "$APP" || "$command" == "$APP "* ]]; then
    kill "$pid" 2>/dev/null || true
  fi
done
ps -axo pid=,command= | while read -r pid command; do
  if [[ "$command" == *"cloudflared tunnel run"* && "$command" == *"$TOKEN"* ]]; then
    kill "$pid" 2>/dev/null || true
  fi
done
nohup "$APP" "$@" > "$APP_DIR/homepage.log" 2>&1 < /dev/null &
nohup "$CLOUDFLARED" tunnel run --token-file "$TOKEN" \
  > "$APP_DIR/.cloudflared/tunnel.log" 2>&1 < /dev/null &
disown -a || true
echo 'Homepage and its dedicated tunnel launched; check logs and /healthz.'
