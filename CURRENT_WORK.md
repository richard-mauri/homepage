# Current work — Homepage

## Issue #24 — persistent session recovery for NewsGo (November 9, 2026)

- **Merged baseline:** GitHub `main` includes Issue #22 / merged PR #23 at commit `425e461d8ce5e4a0af7b341d5c01af302992aa17`. The previous CURRENT_WORK status on main was stale; PR #23 is already merged and its fix branch has been removed by the owner.
- **Current scope:** [Issue #24](https://github.com/richard-mauri/homepage/issues/24), branch `fix/24-newsgo-session-recovery`, draft PR pending. Coordinated with NewsGo [Issue #79](https://github.com/richard-mauri/newsgo/issues/79) / draft [PR #80](https://github.com/richard-mauri/newsgo/pull/80).
- **Observed defect:** Browser closure clears NewsGo's tab-scoped transfer bearer, while Homepage remains signed in. A direct NewsGo visit reports Not signed in until the owner goes to Homepage and launches NewsGo.
- **Implementation:** Fixed `GET /launch/newsgo` route checks the existing Homepage HttpOnly host-only cookie and approved credential, issues a new parent-bound, durable hashed handoff and redirects only to NewsGo with a fragment. No valid session yields `#suite_mode=local`; no credentials or unrestricted return URLs; no parent extension or new sign-in. NewsGo PR #80 performs the recovery navigation when its bearer is missing and suppresses loops for explicit local mode.
- **Security:** Parent 12-hour expiry, revocation, owner isolation, and atomic storage remain authoritative. Existing Homepage login, Renew and Disconnect do not change. No NewsGo/Sailing/Stocks/ScraperGo process or tunnel modifications.
- **Validation:** Check Go tests and Intel/Apple Silicon cross-build CI on the **exact final PR head**. Mac/Safari, actual browser-close/reopen, persistence and Cloudflare tunnel acceptance are owner-only and unverified.
- **Acceptance:** Deploy Homepage candidate first **only with owner restart approval**, then NewsGo PR #80 candidate. Sign in to Homepage, open NewsGo and close the browser; restart only NewsGo, open NewsGo directly, verify the identity returns without sign-in. Also verify revoked/expired/absent login returns browser-only, preferences persist, Renew/Disconnect remain only on Homepage, and all other apps and tunnels work.
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
