#!/usr/bin/env bash
# Homepage: tested, sequentially numbered macOS binaries.
set -Eeuo pipefail
cd "$(dirname "$0")"
usage() { echo "Usage: ./devbuild.sh [--help]"; }
case "${1:-}" in
  "") [[ $# -eq 0 ]] || { usage; exit 2; } ;;
  --help|-h) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac
command -v go >/dev/null || { echo 'Go required' >&2; exit 1; }
command -v git >/dev/null || { echo 'Git required' >&2; exit 1; }
git rev-parse --git-dir >/dev/null 2>&1 || { echo 'Git checkout required' >&2; exit 1; }
go test -count=1 ./...
# Private per-checkout sequence, shared across branches, never committed.
COUNTER="$(git rev-parse --git-path homepage-build-counter)"
mkdir -p "$(dirname "$COUNTER")" dist
previous=0
if [[ -f "$COUNTER" ]]; then
  previous="$(cat "$COUNTER")"
  [[ "$previous" =~ ^[0-9]+$ ]] || { echo 'Invalid build counter' >&2; exit 1; }
fi
next=$((10#$previous + 1))
(( next <= 999999 )) || { echo 'Build counter exceeded' >&2; exit 1; }
printf '%d\n' "$next" > "${COUNTER}.tmp"
mv "${COUNTER}.tmp" "$COUNTER"
printf -v build 'Build %06d' "$next"
for arch in amd64 arm64; do
  target="dist/homepage-darwin-$arch"
  CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" go build -trimpath -buildvcs=false \
    -ldflags="-s -w -X 'main.appBuild=$build'" -o "$target" .
  chmod 755 "$target"
done
(cd dist && shasum -a 256 homepage-darwin-amd64 homepage-darwin-arm64 > SHA256SUMS)
echo "Homepage $build"
ls -lh dist/homepage-darwin-{amd64,arm64} dist/SHA256SUMS
