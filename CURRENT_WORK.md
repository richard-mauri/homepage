# Current work — Homepage

## Verified development state (November 8, 2026)

- **Merged baseline:** GitHub `main` at `4e94ace432b82bd0e3f26944e99cc30ca9361fb9`. Previous Issue #17 auth-trace draft PR #19 status was checked; no Homepage PR was open at this work session's start.
- **Current scope:** [Issue #20](https://github.com/richard-mauri/homepage/issues/20), branch `feat/20-session-renewal`, draft [PR #21](https://github.com/richard-mauri/homepage/pull/21); coordinated with [NewsGo draft PR #69](https://github.com/richard-mauri/newsgo/pull/69) for NewsGo Issue #57.
- **Feature:** A NewsGo-origin-only authenticated POST `/api/handoff/extend-session` accepts an unexpired, parent-bound NewsGo bearer, revalidates approved credential, extends parent session for twelve hours from renewal, rotates the child bearer and durably commits both with rollback on error. Homepage's own browser revalidates retained opaque session IDs after cross-app extension, even if its cached expiry passed.
- **Invariants:** No durable approved credential, token printing, cross-origin cookie fallback, renewal of expired/revoked/unbound bearers, change to user records or other apps' settings. Old bearer invalid after rotation. Existing same-parent handoffs remain subject to their original expiry; sign-out revokes them all. No tunnel/DNS changes, semantic version change, release or merge.
- **Validation:** Added session-extension lifecycle, origin/CORS, expiry/revocation and restart regression tests. [GitHub Actions run 42494535545](https://github.com/richard-mauri/homepage/actions/runs/42494535545) passed at code/documentation head `253c069d365c1ea7b1b291c79e9834fdb7285d75`; recheck after this current-status update. Mac/Safari and live tunnel remain owner-only/unverified.
- **Owner acceptance:** After deploying Homepage and NewsGo candidate branches, sign in through Homepage and open NewsGo. Click Renew before expiry; confirm the timer resets near 12 hours without leaving NewsGo or entering a personal token. Navigate Back to Homepage and verify signed-in state persists; relaunch NewsGo and check extended expiry. Confirm sign-out/revocation and expired sessions cannot renew. Verify sibling apps and settings remain intact.
- **Merge boundary:** Draft PR only. No merge, release, restart, or main changes without owner approval.

## Mac commands

```bash
cd ~/projects/homepage
git fetch origin
git checkout feat/20-session-renewal
git pull --ff-only origin feat/20-session-renewal
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```
