# The Mauris — Homepage

Personal Go-powered launchpad published at https://home.themauris.org/ through its dedicated Cloudflare Tunnel. The site links to four **independent** applications:

| Card | Destination |
| --- | --- |
| Sailing Outlook | https://weather.themauris.org/ |
| Stocks | https://stocks.themauris.org/ |
| NewsGo | https://newsgo.themauris.org/ |
| ScraperGo | https://scrapergo.themauris.org/ |

The cards navigate in the current tab. When signed in, launching NewsGo, Stocks or SailingGo passes a session-bound Homepage handoff token; when signed out, the Homepage launcher denies application access and offers personal or authenticated guest sign-in. The direct-visit recovery bridge is also being corrected in Issue #60; consumer application enforcement is a separately tested rollout. Homepage manages the single 12-hour login session. While that login is still valid, Homepage's **Manage tokens → Current session** panel displays the remaining session time and a **Renew** button beside **Disconnect** that extends the session to twelve hours from renewal without asking for the personal approved credential again. The **Disconnect** button in this same panel is the single sign-out control: it revokes the Homepage session and its child handoffs, and clears Homepage-origin browser session caches. Every signed-in user can access Manage tokens, while credential administration remains admin-only. The Homepage renewal operation refreshes the host-only browser cookie and keeps the same opaque session ID and extends only its currently valid child handoffs. It remains subject to credential revocation; expired sessions must sign in again. The previous NewsGo renewal API remains available temporarily for backward compatibility pending coordinated NewsGo UI migration. Homepage validates its stored session ID when returning from NewsGo so renewed expiration is recovered. Override their targets with `SAILING_URL`, `STOCKS_URL`, `NEWSGO_URL`, and `SCRAPERGO_URL` if necessary.

## Direct-visit session recovery

A fresh browser visit to NewsGo, Stocks or SailingGo without an app handoff can temporarily navigate to one of Homepage's strictly allowlisted fixed routes: `/launch/newsgo`, `/launch/stocks`, or `/launch/sailing`. These return a same-origin, no-store, strict-CSP bridge. The bridge validates an existing (not expired or revoked) Homepage login using Homepage's opaque local session ID or host-only cookie via the existing `POST /api/handoff/token` endpoint, then passes only a newly minted child bearer through the destination URL fragment. The bridge never creates or extends a Homepage login; no personal credential is exposed and the destinations are never taken from query parameters. An authoritative unauthenticated response returns the app in explicit browser-only mode, while transient errors remain visible rather than logging the user out. The individual app must implement the one-time redirect to its recovery route; Homepage alone cannot force a direct visit to recover.

## Running Homepage

Homepage is a locally hosted Go application published through its dedicated Cloudflare Tunnel at https://home.themauris.org/. The server listens on `127.0.0.1:8787` by default; its health endpoint is `/healthz`. Its private tunnel and persistent account data remain on the host, never in Git.

For source requirements, exact Mac checkout, test, build, version-check and authorized launch commands, see [DEVELOPMENT.md](DEVELOPMENT.md). Developer commands are deliberately maintained there rather than duplicated in this user-facing README. The owner decides when a service or tunnel may be restarted.

## Repository reference documents

The suite-wide UI component specification and Homepage-owned Go button module are documented in [SUITE_UI_DESIGN_SYSTEM.md](SUITE_UI_DESIGN_SYSTEM.md). This is separate from the development workflow policy.

Before development, read [AGENTS.md](AGENTS.md), the suite-wide [SUITE_DEVELOPMENT_POLICY.md](SUITE_DEVELOPMENT_POLICY.md), and [CURRENT_WORK.md](CURRENT_WORK.md), and verify live GitHub state. [ARCHITECTURE.md](ARCHITECTURE.md) specifies Homepage ownership and security boundaries; [DEVELOPMENT.md](DEVELOPMENT.md) supplies Homepage-specific builds and Mac acceptance. Homepage `main` is the shared workflow authority; consumer adoption and any amendments require separate owner-reviewed PRs.

The homepage footer is rendered by the Go server with the running executable's semantic version and exact build number, matching `-version`. `/api/config` also returns those values. The version is visible even if the browser JavaScript fails or is cached.

The draft Homepage UI integration uses the Homepage-owned shared banner package for the launchpad masthead while retaining the existing hero, application cards and session controls. Operational controls remain in a secondary toolbar; the new banner shows verified identity when the browser receives it from the established Homepage session API.

With `SUITE_DATA_DIR` configured, Homepage stores its 12-hour session state at `SUITE_DATA_DIR/auth/homepage-sessions.json`. It persists hashed session IDs and app handoff references, without writing approved personal credentials or raw session/bearer tokens. Keep that directory on persistent local storage with restrictive access. A Homepage process restart must not shorten an unexpired session, but sign-out and credential revocation still invalidate it. The service fails to start rather than silently discarding an unreadable state file; write errors are surfaced instead of reporting a successful login or logout. Existing sessions created by earlier in-memory-only versions cannot be recovered on first upgrade.\n\nThe browser assets live under `static/`. `main.go` owns HTTP setup, embedded assets, and the JSON card configuration at `/api/config`. The health check is `/healthz`. The homepage intentionally does not proxy or manage the four linked services.

## Opt-in authentication trace

For troubleshooting a premature **Sign in** display, start Homepage with `./homego.sh -trace` **only after authorizing the application restart**. The launcher passes the flag to the Homepage executable and redirects its output to `~/projects/homepage/homepage.log`. This is a temporary diagnostic option, not permanent debug logging. Normal `./homego.sh` starts without trace. The launcher **truncates** `homepage.log` on each restart: copy an important trace before restarting.

With `-trace`, Homepage logs `[AUTH-TRACE]` records for session validation requests, status codes, authentication transport (session header, app bearer or none), presence of the Homepage cookie, selected session rejection reason, and navigation/UI events reported by the Homepage browser. Events include page load, focus, page restoration, network reconnection, sign-in result, validated state, app launch and handoff result. Correlation uses an ephemeral random browser page ID, **not** a credential. Trace messages contain no raw bearer, session ID, approved token, cookie value, personal preferences or identity. The tracing endpoint `POST /api/trace` accepts only strict enumerated browser diagnostics from Homepage's origin and is unavailable without `-trace`.

After reproducing, inspect with `grep '\[AUTH-TRACE\]' ~/projects/homepage/homepage.log | tail -n 120`. Check whether the browser recorded `sessionStored=false`, whether the server returned `status=401`, and whether `reason=session_not_found`, `session_expired` or `credential_unapproved` appears. `status=200` paired with `ui_signed_out` would suggest browser UI state/race rather than server revocation. Copy only redacted trace lines when sharing diagnostics; do not share the complete private log without reviewing it first.

## SailingGo account preferences (coordinated candidate)

The Sailing Outlook card now uses the same signed-in Homepage session-bound handoff as Stocks and NewsGo, with a separate per-user `/api/storage/sailing` namespace. It sends only the temporary bearer in the destination URL fragment; signed-out visits continue in browser-only mode. SailingGo offers explicit account Save and Load actions before automatic account updates are enabled. The Homepage/SailingGo companion changes are merged; the owner controls future deployment and Safari acceptance.

## Suite activity telemetry (Issue #28, merged)

The activity collector is an **unreleased API-only candidate** on the scoped development branch. With `SUITE_DATA_DIR` enabled, Homepage exposes `POST /api/admin/activity/ingest` for server-to-server reports and admin-only `GET /api/admin/activity` for current reports and hourly aggregates. There is no activity dashboard, browser reporter, or sibling-service instrumentation yet; unreported services are shown as `unknown`, not as quiet.

Each instrumented service must use its own private `HOMEPAGE_ACTIVITY_KEY_<SERVICE>` environment variable (HOMEPAGE, NEWSGO, STOCKS, SAILING, SCRAPERGO; use normal ASCII without the displayed zero-width separator). Values must contain at least 32 characters and must never be distributed to browser code or committed. A report requires the service identity, a matching server-side bearer, a timestamp within two minutes, a nonrepeated nonce, and a bounded JSON payload. Set these only after the companion service reporters have been reviewed; collection is not enabled merely by configuring keys.

Aggregated hourly samples are persisted atomically under `SUITE_DATA_DIR/activity/hourly.json` with mode `0600`, retained for about 30 days. Samples indicate the number of accepted aggregate reports, not distinct hourly visitors; `active2Sum / samples` is the observed mean of each report's two-minute active count. Only reported activity is measured, and an unavailable history file marks history unavailable without modifying account sessions. Hourly history is not a substitute for the live freshness status. No automated restart is provided.

### Homepage browser activity

The merged implementation includes a same-origin `POST /api/activity/ping` endpoint and a 60-second browser presence heartbeat for the Homepage launchpad. It counts anonymous visits and, if a valid Homepage cookie is present, authenticated visits. The client stores a random identifier only in origin-scoped `sessionStorage`; no user IDs, IP addresses, clicked content, or credentials are stored in telemetry. A fresh visitor record is considered active for up to two minutes. Browser interaction and tab presence are separate signals; a page hidden in Safari does not send ongoing heartbeats. The companion apps now implement their own optional activity reporting.

Hourly report persistence is batched: the first accepted aggregate is synced immediately, then subsequent aggregates are flushed no more frequently than every five minutes when new telemetry arrives. A process crash can lose the last few minutes of aggregates; in-memory presence remains deliberately nonpersistent. The current design does not install a periodic shutdown flush or guarantee exact hourly totals. This is merged optional diagnostic monitoring, not guaranteed exact-count analytics.

### Admin Session Activity panel (Issue #28)

On the activity branch, the existing `/suite-admin` page displays a separate **Session activity** section only for an authenticated Homepage administrator. It fetches `GET /api/admin/activity` with the current Homepage session and periodically refreshes every 30 seconds while visible. There are per-service status cards for Homepage, NewsGo, MauriStocks, SailingGo and ScraperGo, including two-, fifteen- and sixty-minute counts, anonymous/signed-in status, in-flight requests and the last accepted report. Missing/stale service observations are explicitly labeled **unknown**, and their numeric fields are hidden rather than displayed as zero. The read-only hourly pattern visualizes average two-minute observed active counts by UTC hour of day, for the last seven or thirty days. It is a mean across reported samples, not a unique-visitor count; an hour with no samples is not an hour with zero traffic. The panel contains no restart button or ability to manage services. The existing **Current session** controls remain accessible to ordinary signed-in users.

For testing on a separate loopback port, a newly created, **isolated** test data directory has no approved administrator tokens, so you must explicitly provision a disposable test admin to view the panel. Never copy production credentials or reuse the running application's persistent `SUITE_DATA_DIR` just to test telemetry. Safari and Mac validation are still required; no services are deployed by this PR.

### Developer acceptance

For isolated telemetry test fixtures and authorized Mac/Safari acceptance, consult [DEVELOPMENT.md](DEVELOPMENT.md) and the current scoped PR. Do not use production credentials or restart the live service merely to inspect the dashboard.

## Authenticated suite access — Issue #60 candidate

Every application must require **verified Homepage personal-user or temporary guest access**. An unauthenticated direct visit must never be converted into anonymous/browsing-only access. The Homepage recovery bridge now attempts the existing same-origin handoff and, on missing/expired/rejected parent sign-in, sends the user to Homepage's personal-or-guest choice rather than redirecting to the destination with `#suite_mode=local`. A verified handoff still travels only in the destination URL fragment; a server outage or invalid handoff never authorizes anonymous access. Existing Homepage account-storage authorization, per-user isolation and guest storage exclusion remain unchanged.

Homepage now includes an initial reusable, separately versioned Go module at `shared/auth` with an opt-in server-side access guard and explicit allowed roles/permissions. Consumers must provide a **trusted verifier** backed by Homepage; neither the visual banner nor a client-provided role label grants access. This first Homepage candidate does **not** yet enforce access on consumer app servers. Separate approved consumer PRs must integrate that guard and deny protected APIs to unauthenticated users; until that rollout finishes, direct URLs and consumer API paths are not suite-wide secure.
