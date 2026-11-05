# Current work

**Authoritative source:** GitHub `main` is the merged baseline; documentation PR #16 merged 2026-11-05. No release/deployment authorized.

**Open regression:** [Issue #17](https://github.com/richard-mauri/homepage/issues/17). User saw **Sign in** about ten minutes after visiting MauriStocks, with no intentional Homepage restart. Actual live HTTP error/status is unconfirmed.

**Scoped draft:** [PR #18](https://github.com/richard-mauri/homepage/pull/18), branch `fix/homepage-session-resilience-20261105`. Browser-side refresh now distinguishes 401/403 from transient connectivity/5xx errors, provides a retry, and rechecks on focus, pageshow, and online; server-side Homepage session and parent-linked handoff maps are now durably serialized to `SUITE_DATA_DIR/auth/homepage-sessions.json` with 0600 permissions, file sync and atomic rename. Session IDs and handoff bearer values are SHA-256-hashed in persisted keys; approved personal tokens are not stored. Persisted expiration remains fixed, and sign-out revocation is persisted. Unreadable state fails startup; failed writes do not acknowledge successful state changes.

**Test coverage:** New Go restart-simulation tests cover 10-minute navigation gap, persistence after recreating the service, per-browser isolation, revocation after restart, absolute expiration, raw-token exclusion, file permissions, corrupt state and persistence write errors. Source-level UI regression test also added. Exact-head CI and Mac/Safari acceptance must be checked and reported separately.

**Deployment caution:** Moving from the old memory-only build to this build cannot recover any sessions the old process kept only in memory. The first rollout may require a new sign-in; subsequent process restarts should preserve these new durable sessions. Do not move or delete `SUITE_DATA_DIR`; keep it on durable storage. Do not print or commit private session state.

**Remaining acceptance:** Review PR and exact-head CI, then owner-side Mac acceptance per [DEVELOPMENT.md](DEVELOPMENT.md): checkout the actual branch, `go test -count=1 ./...`, `./devbuild.sh`, inspect `./dist/homepage-darwin-amd64 -version`, owner-approved `./homego.sh`, `curl -fsS http://127.0.0.1:8787/healthz`, and verify local/Safari/live tunnel. Exercise repeated Homepage→Stocks→Homepage transitions, 10+ minute return, outage recovery, explicit signout, and restart survival. CI does not validate Safari or a live Mac deployment. No merge, tag, release, deployment, DNS/tunnel changes or restart without explicit owner approval.
