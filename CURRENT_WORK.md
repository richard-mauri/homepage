# Current work — Homepage

## Issue #24 / PR #25 — suite session recovery, reconciled with SailingGo

- The authoritative merged baseline before this PR is `main` at `9728f7b7ca4298ce58fe0f3fe92389782bb115ce`, including merged Homepage SailingGo storage/handoff PR #27 and closed Issue #26. The previous feature head `140bcd1fa18ac47f5a384e6179e76611f2c95710` had a merge conflict against main; this branch merges the two ancestries while preserving the already-merged `sailing` namespace, weather CORS origin and signed-in Homepage card handoff.
- Recovery scope: fixed `GET /launch/newsgo`, `/launch/stocks`, and `/launch/sailing` routes, no-store strict-CSP browser bridge, same-origin POST to existing /api/handoff/token backed by approved existing Homepage 12-hour session; child bearer returned only in URL fragment. Does not extend expired session or revive/rewrite identity.
- Coordinated app consumers: NewsGo PR #80 (draft until merged), MauriStocks PR #12 already merged, and separate SailingGo recovery consumer still required. This Homepage PR by itself does not automatically redirect a direct SailingGo visit.
- Root tests: session_recovery_test.go adds no-GET-credential, fixed destination allowlist, parent preservation/revocation/expiry and durable storage checks. Latest-head CI must be verified following integration; Mac/Safari/tunnel not verified by assistant.
- Owner authorized reconcile and merge on November 10, 2026. Merge only after passing checks. Source merge/branch operations are remote; no Mac checkout, build, restart, Cloudflare modification or release has occurred.

## Mac checkout and checks

```sh
cd ~/projects/homepage
git status --short --branch
git fetch origin
git switch main
git pull --ff-only origin main
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
```

Use `./homego.sh` and `curl -fsS http://127.0.0.1:8787/healthz` only if the owner separately authorizes a restart, then Safari test direct visits to NewsGo, Stocks and SailingGo after closing the browser, and expired/revoked/browser-only modes.
