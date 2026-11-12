# Current work — Homepage

## Issue #28 — suite activity collector (runtime candidate)

- Authoritative merged base: `main` commit `f3535b3166dd2a64876cef607860e102f138d598` (November 10, 2026).
- Active development branch: `issue-28-activity-collector` (draft PR #30); paired design-only draft PR #29 uses independent branch `issue-28-suite-activity-design`.
- Scope: Homepage-only first collector checkpoint. `activity.go` exposes service-authenticated aggregate intake and server-verified admin read. Routes are mounted only when `SUITE_DATA_DIR` enables suite storage. Separate service keys are configured through `HOMEPAGE_ACTIVITY_KEY_HOMEPAGE`, `_NEWSGO`, `_STOCKS`, `_SAILING`, `_SCRAPERGO`. Each configured key must be at least 32 characters. Keys are local secrets and must never enter Git or browser code.
- Presently no sibling service sends reports, so the admin API must show `unknown` for uninstrumented apps. This checkpoint includes **30-day batched hourly history and Homepage-origin browser presence reporting (anonymous and signed-in)**, and an **admin-only read-only Session Activity panel**, but has **no instrumentation in NewsGo/Stocks/SailingGo/ScraperGo, and no restart logic**. Counts are bounded aggregate snapshots supplied by trusted service reporters, not independently verified individual sessions.
- Regression tests in `activity_test.go`, `activity_history_test.go`, `activity_home_test.go`, and `activity_admin_test.go` cover invalid counts, unauthorized browser and service calls, replay, fresh/stale/uninstrumented status, anonymous admin denial. GitHub Actions Go CI #215 passed on the admin-panel head `aa10d91`. New local Safari Session-header fix and disposable test-admin helper need exact-head CI; this latest admin UI has **not yet been visually validated in Safari**. No production rollout.
- Isolated UI testing: run `sh scripts/prepare-activity-ui-test.sh "$TEST_ROOT/data"` only with a temporary `homepage-activity.*` test root. Local loopback `Secure` cookies may be unavailable in Safari; the admin UI now uses its existing short-lived `Session` header in memory after Connect, and does not reuse approved personal tokens. Never point this test at production data.
- Preserve existing personal authentication and all app handoffs; do not modify linked services or tunnels. Never merge/release/restart without owner authorization.
- Follow-up: have the owner test admin UI with the guarded disposable test credential on isolated port 8788, validate batching/failure semantics, add per-app cohort/operation reporting; instrument each application in separately scoped PRs before promoting this as a useful live monitor.

## Mac acceptance (only when owner authorizes runtime restart)

```sh
cd ~/projects/homepage
git fetch origin
git checkout issue-28-activity-collector
git pull --ff-only origin issue-28-activity-collector
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

Review actual checkout diff/status and test output before starting the launcher. Safari and live-tunnel acceptance follow only after separately approved restart; a Go build does not establish browser acceptance.
