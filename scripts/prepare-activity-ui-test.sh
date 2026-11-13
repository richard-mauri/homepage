#!/bin/sh
# Create a disposable administrator ONLY inside the isolated Homepage activity
# test directory created with: mktemp -d "${TMPDIR:-/tmp}/homepage-activity.XXXXXX"
# Never use this helper with the live SUITE_DATA_DIR.
set -eu
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 TEST_ROOT/data" >&2
  exit 2
fi
root=$1
case "$root" in
  */homepage-activity.*/data) ;;
  *) echo "Refusing non-temporary or unrecognized test data directory" >&2; exit 2 ;;
esac
if [ ! -d "$root" ] || [ -L "$root" ]; then
  echo "Test data directory must exist and must not be a symlink" >&2
  exit 2
fi
auth="$root/auth"
target="$auth/approved-tokens.json"
if [ -e "$target" ] || [ -L "$target" ]; then
  echo "Refusing to overwrite an existing credential registry" >&2
  exit 2
fi
command -v openssl >/dev/null || { echo "openssl is required" >&2; exit 1; }
command -v shasum >/dev/null || { echo "shasum is required" >&2; exit 1; }
umask 077
mkdir -p "$auth"
chmod 700 "$auth"
token=$(openssl rand -hex 32)
digest=$(printf %s "$token" | shasum -a 256 | awk '{print $1}')
# A 64-character lowercase hex token is the only value displayed to the
# operator. The file contains its SHA-256 digest and no reusable raw token.
printf '[{"user":"activity-test-admin","digest":"%s","role":"admin","label":"Disposable isolated activity UI test"}]\n' "$digest" > "$target"
chmod 600 "$target"
printf '\nDisposable administrator token (copy it once into the local test UI):\n%s\n\n' "$token"
echo "This token works ONLY with the isolated test data directory."
echo "Do not paste it in chat, commit it, or use it against production Homepage."
