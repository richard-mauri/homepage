# Current work — Homepage

**Merged baseline:** GitHub `main` at `b15057701523f7c0b1e90f51f8a7f636d78ad483` (PR #57 merged). Homepage shared UI foundation PR #55 and consumer PR #57 are merged; the previous `CURRENT_WORK.md` statement describing PR #57 as open was stale. Older PR #53 documentation work is separate; verify current GitHub state before changing it.

**Active scope:** [Issue #58](https://github.com/richard-mauri/homepage/issues/58), branch `fix/58-storage-cors-preflight`, draft PR to `main` (see live GitHub for PR number and latest exact head/CI). A Safari no-store storage `PUT` from Stocks preflights `authorization,content-type,cache-control,pragma`, but the old CORS handler allowed only the first two, returning reproducible HTTP 403 `preflight headers not allowed`. This branch accepts only the four named headers, with order/case-independent matching, for existing trusted origins and GET/PUT. Regression tests cover allowed and denied cases. All actual storage authorization, per-user data isolation, and revision guards are untouched.

**Safety:** GitHub branch changes only. No production accounts, preferences, subscriptions, watchlists, browser state, session tokens, persistent data, Git-private Mac build counter, Cloudflare DNS/tunnel, launchers or processes were read or modified. Owner Mac-local changes are not visible and must be checked before checkout.

**Verification:** Required exact-head Go CI must pass on the final draft PR commit; do not rely on older run results. Mac local build/Safari/production preflight acceptance not run. The prior owner read-only curl returned 403 and is evidence of the original defect, not verification of this fix.

**Owner handoff (only after CI passes):** From `~/projects/homepage`, inspect `git status --short --branch` and stop if dirty; `git fetch origin`, `git checkout fix/58-storage-cors-preflight`, `git pull --ff-only origin fix/58-storage-cors-preflight`, `./devbuild.sh`, `./dist/homepage-darwin-arm64 -version`. Only after separate authorization, `./homego.sh` restarts Homepage and its tunnel; check `curl -fsS http://127.0.0.1:8787/healthz`, then the credential-free preflight in `DEVELOPMENT.md` and Safari account synchronization.

**Promotion:** Draft PR only; do not FINISH, merge, close Issue #58, delete branch, restart, deploy, tag or release without the respective owner authorization. MauriStocks draft PR #34 remains independent and unmerged.
