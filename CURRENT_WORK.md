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

## November 12, 2026 — administrator-page heartbeat fix (Homepage-only)

- **Production observation:** when Safari remains on `/suite-admin`, Homepage becomes **Stale** after its last `/api/activity/ping` ages past 150 seconds, although the four other app reporters remain fresh. The main Homepage `/` embeds `static/activity.js`; the administrator page previously did not. Its 30-second dashboard data refresh is not a browser-presence heartbeat.
- **Fix:** `static/suite-admin.html` loads the existing same-origin `/static/activity.js` once, deferred. Initial implementation did not work because `/suite-admin` CSP allowed only `script-src 'unsafe-inline'`, blocking the external script. **Follow-up:** `main.go` now permits `script-src 'self' 'unsafe-inline'` while retaining `default-src 'none'`, `connect-src 'self'`, and `frame-ancestors 'none'`. The CSP is a named constant with regression coverage; no authentication or collector changes. The reporter uses the same per-tab `sessionStorage` ID when navigating within Homepage; browser requests carry only ephemeral ID and visibility data to Homepage. No keys in browser code.
- **Regression:** `TestHomepageAdminPageIncludesPresenceHeartbeat` asserts the script is embedded once and targets the same-origin ping route; `TestHomepageAdminCSPPermitsHeartbeatScript` guards that the CSP permits both the external script and same-origin fetch while retaining its restrictive baseline. The previous freshness threshold (150 seconds), rolling-count logic and other services are unchanged.
- **Validation:** exact-head GitHub Actions required after this documentation commit; owner's Mac/Safari verification remains pending. Production binary Build 000045 was running before this fix; source edits do **not** update it. Do not restart Homepage, release, merge, or alter tunnels without owner authorization.
- **Mac handoff:** before pulling, check `git status -sb`: the installed local `homego.sh` has the intentional uncommitted private activity-loader patch and must **not** be overwritten. Pull fast-forward changes if possible, run `go test -count=1 ./...` and `bash ./devbuild.sh`, verify the new `./dist/homepage-darwin-arm64 -version` stamp; only after owner approves, run `./homego.sh`, check local/public `/healthz`, and leave Safari on `/suite-admin` more than three minutes, refreshing until Homepage stays **Reporting** with at least one active tab. Confirm all four reporters continue reporting.

## November 12, 2026 — false “Never reported” after collector restart

- **Observed:** after rebuilding/restarting Homepage, the admin dashboard showed NewsGo/SailingGo/ScraperGo as `Never reported` despite verified earlier reporting. Homepage deliberately persists hourly aggregate history but **not** the live last-report map: the collector's `latest` map starts empty after every process restart. The other reporters publish on their own roughly 60-second clocks; fresh reports repopulate that map without restarting the reporters.
- **Fix:** show `No report since Homepage started` when no last report exists in the current collector process, rather than `Never reported`. Preserve `Not reporting — activity unknown` and blank activity numbers until the next real report. Do **not** infer fresh status, person counts, or exact last report timestamps from hourly history.
- **Regression:** `TestActivityAdminPanelEmbeddedAndReadOnly` now requires the truthful empty-report text in the administrator HTML. After deploying, keep the four reporters running and verify they become Reporting on their next scheduled report; if any stay unknown beyond a couple of report intervals, inspect only that service's logs/configuration, without disclosing keys.
- **Deployment:** no automatic restart, merge, or release. Mac must fetch/pull, test and rebuild Homepage; current live binary does not reflect these changes until separately restarted by owner.

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
