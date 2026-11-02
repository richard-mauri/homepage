# Homepage current work

**Date:** 2026-11-02. **Status:** Draft backend integration for Issue #2; not merged or deployed.

- Authoritative merged baseline: `main`; scoped working branch: `feat/central-suite-storage`, draft PR #3.
- The previous `feat/versioned-build-workflow` candidate is separate and not silently merged.
- Implemented: `internal/suitestore` validates NewsGo-style hashed tokens (ordinary and admin users), stores per-user/per-app revisioned JSON, performs atomic writes, checks write conflicts, and has auth/isolation tests.
- Added `configureSuiteStorage` to Homepage's Go server. **Default disabled**; with operator-set `SUITE_DATA_DIR`, GET/PUT `/api/storage/{newsgo|mauristocks}` endpoints are registered and require bearer tokens. No token registry or NewsGo data is migrated automatically. Without a populated `SUITE_DATA_DIR/auth/approved-tokens.json`, authentication fails closed. Use a private Mac directory outside source control.
- Added opt-in session endpoints at `/api/session`: same-origin POST using existing bearer token establishes a random 12-hour Secure, HttpOnly, SameSite=Lax cookie scoped to themauris.org; GET reports current session; DELETE logs out. Sessions are in process memory and do not persist across restart. **Storage API still requires bearer tokens, not session cookies.** Token revocation does not immediately invalidate an existing session; rotation/revocation semantics must be completed before deployment. No browser UI and no cross-subdomain app integration exist yet. Do not enable this on the live tunnel.
- Added `.github/workflows/ci.yml` for Go tests and macOS cross-builds. Prior run 41473486537 passed; check CI again on the exact latest PR head. No visual diffs (no HTML/CSS/JS changes).
- **Remaining:** session authorization for the storage API, immediate revocation and session invalidation, CSRF defenses for cookie-authenticated writes, cross-subdomain client behavior, security review, persistent root provisioning, backup/recovery and migration, Homepage admin token UI, NewsGo and MauriStocks integration, final Mac/Safari/Cloudflare acceptance.
- No applications, credentials, tunnels, DNS, semantic versions, or runtime processes were changed.
- Mac acceptance still uses `go test -count=1 ./...`, `./devbuild.sh`, stamped Intel binary, Homepage-only launch and Safari/Cloudflare checks as applicable. No merge, release, or deployment without owner authorization.
