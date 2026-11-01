#!/bin/bash
# Start homepage (Intel/amd64) and its own Cloudflare Tunnel.
# ScraperGo and all other services/tunnels remain untouched.
# Requires homepage/.cloudflared/token (do not commit this file).

HOME=/Users/richardmauri
APP_DIR="$HOME/projects/homepage"
APP="$APP_DIR/dist/homepage-darwin-amd64"
TOKEN="$APP_DIR/.cloudflared/token"
CLOUDFLARED="$HOME/Applications/cloudflared"
PORT=8787

cd "$APP_DIR" || exit 1
for file in "$APP" "$TOKEN" "$CLOUDFLARED"; do
    if [[ ! -f "$file" ]]; then
        echo "Missing required file: $file" >&2
        exit 1
    fi
done
if [[ ! -x "$APP" || ! -x "$CLOUDFLARED" ]]; then
    echo "Homepage binary and cloudflared must be executable." >&2
    exit 1
fi

# Verify the *exact* executable before stopping any existing homepage process.
# An old dist binary without -version must not silently replace a working app.
#if ! version_output="$("$APP" -version 2>&1)"; then
#    printf 'Error: Homepage version check failed: %s\n' "$version_output" >&2
#    exit 1
#fi
#if [[ ! "$version_output" =~ ^Homepage\ v[0-9]+\.[0-9]+\.[0-9]+\ ·\ Build\ [0-9]{6}$ ]]; then
#    printf 'Error: unexpected Homepage executable version: %s\n' "$version_output" >&2
#    exit 1
#fi
#printf 'Homepage executable version: %s\n' "$version_output"

# Stop only the Homepage instance launched from this project's dist directory.
ps -axo pid=,command= | while read -r pid command; do
    if [[ "$command" == "$APP"* ]]; then
        kill "$pid" 2>/dev/null || true
    fi
done

# Stop only Homepage's Cloudflare Tunnel (identified by Homepage's token path).
ps -axo pid=,command= | while read -r pid command; do
    if [[ "$command" == *"cloudflared tunnel run"* &&
          "$command" == *"$TOKEN"* ]]; then
        kill "$pid" 2>/dev/null || true
    fi
done

nohup "$APP" "$APP_DIR/homepage.log" 2>&1 < /dev/null &

nohup "$CLOUDFLARED" tunnel run \
    --token-file "$TOKEN" \
    > "$APP_DIR/.cloudflared/tunnel.log" 2>&1 < /dev/null &

disown -a

