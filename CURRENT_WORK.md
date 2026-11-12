# Current work — Homepage

## Issue #28 — suite activity collector (runtime candidate)

- Authoritative merged base: `main` commit `f3535b3166dd2a64876cef607860e102f138d598` (November 10, 2026).
- Active development branch: `issue-28-activity-collector` (draft PR #30); paired design-only draft PR #29 uses independent branch `issue-28-suite-activity-design`.
- Scope: Homepage-only first collector checkpoint. `activity.go` exposes service-authenticated aggregate intake and server-verified admin read. Routes are mounted only when `SUITE_DATA_DIR` enables suite storage. Separate service keys are configured through `HOMEPAGE_ACTIVITY_KEY_HOMEPAGE`, `_NEWSGO`, `_STOCKS`, `_SAILING`, `_SCRAPERGO`. Each configured key must be at least 32 characters. Keys are local secrets and must never enter Git or browser code.
- Presently no sibling service sends reports, so the admin API must show `unknown` for uninstrumented apps. This checkpoint includes **30-day aggregated hourly history with atomic persistence**, but has **no dashboard UI, no service-side browser activity tracking, and no restart logic**. Counts are bounded aggregate snapshots supplied by trusted service reporters, not independently verified individual sessions.
- Regression tests in `activity_test.go` and `activity_history_test.go` cover invalid counts, unauthorized browser and service calls, replay, fresh/stale/uninstrumented status, anonymous admin denial. GitHub Actions Go CI run #196 passed on head `86755b29f5793d79864df06452fb5a121c8dd0f2`: Go tests and both macOS architecture builds succeeded. Mac/Safari/tunnel checks are **not run**; this documentation edit requires a new head CI check.
- Preserve existing personal authentication and all app handoffs; do not modify linked services or tunnels. Never merge/release/restart without owner authorization.
- Follow-up: add complete cohort/operation reporting, ingestion batching/performance checks and the admin dashboard; instrument each application in separately scoped PRs before promoting this as a useful live monitor.

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
