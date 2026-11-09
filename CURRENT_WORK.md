# Current work — Homepage

## Issue #24 — persistent session recovery for NewsGo and Stocks (November 9, 2026)

- **Merged baseline:** GitHub `main` includes Issue #22 / merged PR #23 at commit `425e461d8ce5e4a0af7b341d5c01af302992aa17`. The previous CURRENT_WORK status on main was stale; PR #23 is already merged and its fix branch has been removed by the owner.
- **Current scope:** [Issue #24](https://github.com/richard-mauri/homepage/issues/24), branch `fix/24-newsgo-session-recovery`, draft [PR #25](https://github.com/richard-mauri/homepage/pull/25). Coordinated with NewsGo [Issue #79](https://github.com/richard-mauri/newsgo/issues/79) / draft [PR #80](https://github.com/richard-mauri/newsgo/pull/80) and MauriStocks [Issue #11](https://github.com/richard-mauri/mauristocks/issues/11) / draft [PR #12](https://github.com/richard-mauri/mauristocks/pull/12).
- **Observed defect:** Browser closure clears NewsGo and MauriStocks tab-scoped transfer bearers, while Homepage remains signed in. A direct app visit reports Not signed in until the user goes to Homepage and relaunches.
- **Implementation:** Fixed `GET /launch/newsgo` and `GET /launch/stocks` routes serve the same embedded same-origin, no-store/CSP-protected recovery page with strict fixed destination allowlist. Its JS uses the retained Homepage opaque session ID and falls back to the host-only cookie, calling existing same-origin POST `/api/handoff/token` to issue a durable parent-bound handoff; then it returns only to the selected trusted NewsGo or Stocks origin with a fragment. An authoritative 401 yields `#suite_mode=local`; network failures remain visible as unavailable. No handoff is created by GET, no personal credentials or unrestricted return URLs, no parent extension or new sign-in. NewsGo PR #80 and MauriStocks PR #12 perform the recovery navigation when their bearer is missing and suppress loops for explicit local mode.
- **Security:** Parent 12-hour expiry, revocation, owner isolation, and atomic storage remain authoritative. Existing Homepage login, Renew and Disconnect do not change. No NewsGo/Sailing/Stocks/ScraperGo process or tunnel modifications. Only coordinated draft code changes.
- **Validation:** Go CI [run 42693822238](https://github.com/richard-mauri/homepage/actions/runs/42693822238) passed at implementation head `a5eb35461ab15046d4e04d1e30d264838d88f459` (Go tests and both macOS architecture builds); verify CI again after this documentation-only snapshot update. Mac/Safari, actual browser-close/reopen, persistence and Cloudflare tunnel acceptance are owner-only and unverified.
- **Acceptance:** Deploy Homepage candidate first **only with owner restart approval**, then NewsGo PR #80 and MauriStocks PR #12 candidates. Sign in to Homepage, open each app, close the browser; restart only that app, open it directly, verify identity returns without sign-in. Also verify revoked/expired/absent login returns browser-only, preferences persist, Renew/Disconnect remain only on Homepage, and all other apps and tunnels work.
- **Merge boundary:** Draft PR only; no merge, release, semantic version change, restart or tunnel edits without owner approval.

## Mac checkout and acceptance (Homepage)

```bash
cd ~/projects/homepage
git fetch origin
git checkout fix/24-newsgo-session-recovery
git pull --ff-only origin fix/24-newsgo-session-recovery
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

After authorized restart, validate local Homepage and https://home.themauris.org/ in Safari, session recovery, approved-user roles, all four launch destinations, mobile/desktop UI and independent tunnels.
