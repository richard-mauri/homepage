# Current work — Homepage

## Issue #39 — Immediate verified suite launch observations

- **Baseline:** GitHub `main`, including merged Homepage PR #35 (NewsGo/ScraperGo user activity aliases). Homepage PR #37 (admin guest revocation) is a separate unmerged draft and is not included here. NewsGo PR #94 and ScraperGo PR #137 are separate draft popup UX corrections.
- **Branch:** `fix/39-launch-observations` from `main`; one focused draft PR for Issue #39.
- **Owner's reproduced problem:** A guest can sign into Homepage and immediately launch NewsGo, yet the administrator sees “No observation in past hour” for both because a tab's next heartbeat/report has not arrived.
- **Change:** A successful verified Homepage suite launch sends a bounded, same-origin `POST /api/activity/launch` acknowledgment before navigation. Homepage authenticates the parent session and records separate, ephemeral last-seen entries for Homepage and the selected destination. Admin Users & Access overlays these observations with per-app heartbeat reports without fabricating active sessions. Missing, expired or revoked identity is rejected, and telemetry failure never blocks launch.
- **Security/data:** No raw tokens persisted in activity history; guest observations stay in memory and disappear when the guest is revoked or expires. No token renewal, navigation redirect, change to guest one-hour lifetime or 60-second heartbeat.
- **Validation:** Go tests cover guest verification, origin/service/missing-identity rejection, revocation, expiry, no fabricated session, source integration. Exact PR head CI and owner Apple Silicon Safari acceptance are required. Linux CI does not establish live Safari behavior.
- **Owner Safari acceptance:** Sign in as a one-hour guest, immediately click NewsGo, and inspect Users & Access from an admin session. Homepage and NewsGo should show recent **Last seen** observations even if no heartbeat has arrived; **Current sessions** remains heartbeat-derived. Check ScraperGo launch and revoked guest rejection. No automatic redirects.
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
