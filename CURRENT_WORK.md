# Current work — Homepage

## Issue #26 / draft PR #27 — SailingGo authenticated preferences (November 10, 2026)

- Authoritative merged baseline: GitHub main at `ada914b2cbb6068183e1af52de672fc27e661ffe`.
- Branch: `feat/26-sailing-account-storage`, draft PR #27, related SailingGo issue #17 / draft PR #18.
- Scope: register `sailing` per-user account namespace, authorize the fixed `https://weather.themauris.org` browser origin for profile/storage CORS, and issue the existing session-bound handoff on signed-in SailingGo card launch; signed-out launch uses explicit local mode. No change to login life, token durability, legacy cookies or other applications.
- Regression tests: `sailing_storage_test.go` exercises CORS/storage authorization, namespace and card handoff. GitHub CI on latest head must be verified after this documentation update. No Mac, Safari or tunnel acceptance yet.
- Another unrelated open draft PR #25 (`fix/24-newsgo-session-recovery`) exists; do not mix its direct-visit recovery changes into this PR. SailingGo direct-visit bridge is not included.
- No merge, release/version, launch or tunnel changes authorized. Deploy this Homepage dependency only in coordination with SailingGo PR #18 after owner acceptance.

## Mac checkout, build and acceptance

```sh
cd ~/projects/homepage
git status --short --branch
git fetch origin
git checkout feat/26-sailing-account-storage
git pull --ff-only origin feat/26-sailing-account-storage
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

Do not execute `homego.sh` until the owner approves a restart. Confirm existing local changes before checkout. Safari should verify both normal and admin identities, signed-in SailingGo handoff, signed-out browser-only mode, existing Stocks/NewsGo launches, storage 409 isolation, session expiry and sign-out revocation.
