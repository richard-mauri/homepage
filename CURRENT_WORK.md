# Current work — Homepage

- **Latest UX adjustment:** Homepage's guest control now reads `Sign in as Guest` throughout the launcher and account UI. A verified handoff remains required to navigate from Homepage; invalid sessions stay on Homepage. CI #320 passed before this note; owner Safari acceptance is pending.

## Consolidated draft PR #37 — Guest revocation and verified navigation activity

- **Authoritative baseline:** GitHub merged `main` includes Homepage PR #35 (per-user activity aliases). Draft PR #37 now combines Issue #36's admin Revoke Guest feature with Issue #39's immediate verified suite navigation activity. Redundant draft PR #40 will be closed only after the combined PR is verified. None of these changes has merged into `main`.
- **Branch:** `feat/36-admin-revoke-guest` (draft PR #37). The working integration uses a merge commit with parents from this branch and `fix/39-launch-observations`; history preserves both lines of work.
- **Guest revocation:** Admin-only `revoke-guest` action on existing `/api/admin/users` invalidates the memory-only guest parent and associated handoffs. Dark-themed confirmation dialog has Cancel/Revoke guest and supports Esc cancellation; guest persists until confirmed. Named users and unrelated guests are unaffected. Expired guests remain subject to memory GC.
- **Verified navigation activity:** Homepage's launch click records bounded, memory-only Last seen for Homepage and all four destinations. Returning from any app to Homepage records a verified Homepage arrival. The Users & Access per-user view includes Homepage, NewsGo, MauriStocks, SailingGo, ScraperGo. Stocks and Sailing currently have launch Last seen only, not their own identity-aware current-session telemetry. ScraperGo companion draft PR #137 immediately reports accepted verified arrival/ping so NewsGo→ScraperGo is recognized without waiting for the next 60-second aggregate report.
- **Semantics:** Last seen from navigation does not create a Current session. Browser heartbeat remains 60 seconds; guest token stays one hour, cannot save/restore preferences, and guest identity/activity records never become durable. No automatic redirects or token renewal. Invalid/missing/revoked identity cannot generate verified launch event.
- **Checks:** Go CI must pass on the final combined head, then owner tests Apple Silicon Mac Safari with named and temporary guests, quick transitions, guest revoke confirmation and invalidated handoffs. CI is not Safari acceptance. Homepage Issue #38 tracks later protected-API review.
- **Boundary:** Do not merge, tag, release, modify tunnels/credentials or restart owner's Mac processes without explicit owner authorization. No sibling repo source changes are included in Homepage PR #37.

Owner Mac acceptance commands:

```bash
cd ~/projects/homepage
git status -sb
git fetch origin
git checkout feat/36-admin-revoke-guest
git pull --ff-only origin feat/36-admin-revoke-guest
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-arm64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```
