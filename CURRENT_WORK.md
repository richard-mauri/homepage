# Current work

**Authoritative source:** GitHub `main` is the merged baseline as of 2026-11-05. Documentation reconciliation PR #16 has merged. There is no active release or deployment authorization.

**Active regression:** [Issue #17](https://github.com/richard-mauri/homepage/issues/17) reports that Homepage displayed **Sign in** about ten minutes after a valid sign-in and same-tab visit to MauriStocks. The owner confirms they did **not** manually restart Homepage. Root cause has **not** been established from live HTTP logs or Safari network traces.

**Scoped branch:** `fix/homepage-session-resilience-20261105` from `main`. The proposed client-only mitigation separates authentication rejection (401/403) from transient network/5xx errors, exposes a retryable connection state rather than an inaccurate **Sign in** label, retries with Homepage's host-only cookie when its stored session ID is rejected, and validates on focus/pageshow/online independently of any app bearer. A source-level regression test covers these invariants.

**Remaining risk / investigation:** Server-side `sessionService.sessions` and `handoffs` are currently volatile maps. Unexpected process failure, service restart, or crash could still prematurely destroy otherwise unexpired 12-hour sessions. A durable store with explicit revocation and expiry semantics is a separate, necessary reliability improvement before claiming full 12-hour survivability across process restarts. Capture actual `/api/session` response codes and check whether the live server is restarting unexpectedly. Do not interpret MauriStocks' cached signed-in badge as evidence of a live parent-session validation.

**Next steps:** Review the scoped draft PR, exact-head CI, and syntax/tests. For runtime acceptance, the owner must run the documented Mac checkout/test/build/launch sequence and verify Safari via both localhost and the live Cloudflare tunnel. Test signed-in navigation to Stocks and back, brief network interruption, invalid/expired session, explicit sign-out and browser page restoration. Do not merge, release, tag, restart or deploy without the owner's explicit approval.
