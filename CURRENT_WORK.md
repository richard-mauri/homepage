# Current work — Homepage

## Issue #39 — Immediate verified suite launch observations

- **Baseline:** GitHub `main`, including merged Homepage PR #35 (NewsGo/ScraperGo user activity aliases). Homepage PR #37 (admin guest revocation) is a separate unmerged draft and is not included here. NewsGo PR #94 and ScraperGo PR #137 are separate draft popup UX corrections.
- **Branch:** `fix/39-launch-observations` from `main`; one focused draft PR for Issue #39.
- **Owner's reproduced problem:** A guest can sign into Homepage and immediately launch NewsGo, yet the administrator sees “No observation in past hour” for both because a tab's next heartbeat/report has not arrived.
- **Change:** A successful verified Homepage suite launch sends a bounded, same-origin `POST /api/activity/launch` acknowledgment before navigation. On any verified return to Homepage, the landing page likewise acknowledges the Homepage arrival. Homepage authenticates the parent session and records ephemeral last-seen entries for Homepage and the selected destination. Admin Users & Access overlays these observations with per-app heartbeat reports without fabricating active sessions. Missing, expired or revoked identity is rejected, and telemetry failure never blocks launch.
- **NewsGo→ScraperGo:** The companion ScraperGo PR #137 now immediately sends its already verified reporter snapshot on a confirmed arrival/ping, instead of waiting for the next 60-second aggregate report. No new cross-origin browser API or NewsGo link rewrite is necessary. Homepage remains the central validation/collector authority.
- **Security/data:** No raw tokens persisted in activity history; guest observations stay in memory and disappear when the guest is revoked or expires. No token renewal, navigation redirect, change to guest one-hour lifetime or 60-second heartbeat.
- **Validation:** Go tests cover guest verification, origin/service/missing-identity rejection, revocation, expiry, no fabricated session, source integration. Exact PR head CI and owner Apple Silicon Safari acceptance are required. Linux CI does not establish live Safari behavior.
- **Owner Safari acceptance:** Sign in as a one-hour guest and click Homepage→each of the four suite cards; verify immediate Homepage/destination **Last seen** observations for the apps presently shown in per-user activity, while other apps await a per-user UI. Click a NewsGo headline to open ScraperGo and verify ScraperGo reports before the next periodic minute. Click Back to Homepage from each app and verify a fresh Homepage observation. **Current sessions** remains heartbeat-derived. Check revoked guest denial and no automatic redirects.
- **Boundary:** No merge, release, tag, tunnel/credential change or service restart without owner authorization. Broader protected-API audit remains Homepage Issue #38.

Mac checkout/test/build/launch sequence for owner acceptance:

```bash
cd ~/projects/homepage
git status -sb
git fetch origin
git checkout fix/39-launch-observations
git pull --ff-only origin fix/39-launch-observations
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-arm64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```
