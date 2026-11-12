# Suite activity monitoring — design proposal (Issue #28)

**Status:** design-only draft, not implemented or deployed. **Owner:** Homepage central collector; four sibling apps remain independently deployed.

## Purpose and exclusions

Help Homepage administrators determine low-impact service restart windows across Homepage, NewsGo, MauriStocks, SailingGo and ScraperGo. Count visits with and without Homepage login, including direct application visits and ScraperGo jobs initiated by NewsGo. There is **no automatic restart**, no claim that inactivity guarantees safety, and no behavioral tracking by named user.

The currently deployed applications do **not** become monitored by merging this document. Homepage currently owns twelve-hour authentication, not sibling-service process supervision.

## Metrics model

For each application, show independent rolling unique active **browser instance** counts over 2, 5, 15 and 60 minutes; authenticated/anonymous breakdown only if a service can verify this locally; last *foreground interaction* and last *browser presence* separately; in-flight request count and last completion; last successfully received report; and collection health (fresh/stale/offline/not instrumented). Also maintain hourly aggregates for up to 30 days. Anonymous visitor counts across multiple apps **cannot** be added to yield an accurate suite-wide unique-person count.

A **presence heartbeat** is not a human action. A hidden page with retained application connection should be shown separately from a foreground page; avoid attributing automatic refresh to user interaction. Record an interaction only for meaningful pointer/keyboard action or deliberate user workflow. Throttle interaction telemetry and avoid logging interaction targets, URLs, coordinates or input values. In-flight tracking counts requests at handler entry and completion, with timeout bounds and crash-safe reset to unknown after process restart. The reader and classification paths of ScraperGo count as operations even if called programmatically from NewsGo.

## Trust boundary and API contract (subject to implementation review)

The browser reports minimal activity only to its own application's **same-origin** server endpoint, not directly to Homepage, with HTTP POST, no-store, a small JSON body and restricted content type. Session-scoped anonymous IDs are generated randomly in the browser, kept in `sessionStorage`, rotated at a sensible boundary and not used across origins. Where an application has verified Homepage bearer auth, it can label reports authenticated, but **never** send personal credentials or the bearer to telemetry. The server does not trust client assertions of admin role, application name, or auth status.

Each service aggregates local browser activity and active operations, then reports bounded aggregates to Homepage using service-to-service authentication and strictly pinned app identity. Prefer loopback-only communication where services share the Mac; if not viable, a mutually authenticated, authenticated transport with explicit allowlist and replay protection is required. Do not embed telemetry shared secrets in browser JavaScript, place them in the Git repository, or reuse users' personal tokens. All submitted timestamps must be validated and server receive-time recorded; reject out-of-window and duplicate reports. Limit body sizes, source cardinality, report rate and timestamp skew. Disallow public cross-origin collection APIs and broad CORS.

Provisional endpoints (names can change during code review):
- Same-origin per-app `POST /api/activity/ping` for browser activity; bounded and rate limited.
- Homepage `POST /api/admin/activity/ingest` for authenticated service collectors, never for arbitrary browsers.
- Homepage `GET /api/admin/activity` for server-verified administrators; `Cache-Control: no-store`.

Ingestion responses are not authenticated-session renewal. The collector must never write or extend the existing Homepage auth session or child handoff.

## State, restart and failure handling

Use bounded in-memory recent observations, with hashed/opaque short-lived identifiers if required for deduplication, and a separate aggregate file under `SUITE_DATA_DIR/activity/`. The file must use restrictive permissions, atomic replacement, and versioned schema. Default aggregate retention 30 days; active raw ephemeral identifiers expire within at most 60 minutes after last report. Avoid unbounded file writes from each ping; rate-limit persistence and flush on clean shutdown when practical. Never store IPs, exact article URLs, stock symbols, sailing location, browser fingerprints, raw credentials, session IDs, personal preferences or individual click paths.

If a source fails to report, mark its status **unknown** after a bounded freshness interval, rather than showing zero activity. Restarting an app loses in-flight state; explicitly mark uncertain until the restarted source reports health and outstanding-work counters. Data loss or collector outage cannot break primary app functionality, sign-in, or existing account persistence. A dashboard must report last telemetry time so stale data are not mistaken for idle.

## Restart-readiness semantics

Per service: **busy** when active foreground presence or in-flight operations are observed; **possibly quiet** only when recent service reports are healthy, foreground active count and in-flight operations are both zero for a continuous 10-minute window; **unknown** for any instrumentation or reporting gap. Quiet is advice, not safety proof; it cannot detect all background jobs and is not authority to restart a service. Show each service independently and a suite-wide summary that is unknown if any service is unknown.

## UI placement

Add an **admin-only** Session Activity section to Homepage `/suite-admin`; leave ordinary users' Manage Tokens / Current Session access intact. Cards for per-service current/recent presence, last interaction, in-flight operations, freshness and advisory quiet/busy/unknown; hourly history for the last 7/30 days. The UI must fetch via server-authorized admin API, with appropriate no-store response headers; hiding the panel in JS is not security.

## Implementation sequence

1. Homepage collector and isolated deterministic Go tests (signed service ingress, admin read authorization, freshness, counters, clock abstraction, retention, atomic persistence). Provide a local mock reporter and verify it can safely report unknown for uninstrumented apps.
2. Implement same-origin activity observation in Homepage, NewsGo, MauriStocks, SailingGo, ScraperGo in independent issue-scoped PRs. Preserve browser-only mode. Report real in-flight tasks using deferred completion hooks.
3. Add admin dashboard, request history and conservative quiet-readiness logic. Verify multiple tabs, direct visits, signed-in visits, anon visits, hidden tabs, network loss, application restarts, service crashes, auth expiration and stale telemetry.

## Release safeguards and validation

Every repository follows its own `AGENTS.md`, `DEVELOPMENT.md` and `CURRENT_WORK.md` workflow: branch from latest merged main, draft PR, Go tests, source-specific CI and separate Mac/Safari acceptance. Verify visible local dashboard and external HTTPS tunnel, including browser-only mode, before considering promotion. No merge/release/tag/version change/DNS changes/tunnel changes/service restart without owner authorization.

## Decisions pending source-level review

Confirm whether a single loopback collector is reachable from all app Go processes and whether server-side app telemetry can authenticate per application without touching existing auth configuration. Confirm per-app request paths and background polling patterns before deciding which handlers count as user work. Confirm retention, local state location and acceptable telemetry intervals in Mac runtime testing.
