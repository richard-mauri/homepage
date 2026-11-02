# Homepage current work

**Date:** 2026-11-02. **Status:** Draft foundation for Issue #2; not merged or deployed.

- Authoritative merged source: `main`; scoped working branch: `feat/central-suite-storage`.
- Previous version/build workflow candidate `feat/versioned-build-workflow` still requires its own review and Mac acceptance; do not silently merge or deploy it.
- Issue #2: central suite authentication and per-user persistent settings for NewsGo, MauriStocks and future clients.
- Current changes: isolated `internal/suitestore` Go package (not registered on Homepage's HTTP mux). It can validate legacy NewsGo SHA-256 approved-token records, isolate JSON documents by user/app namespace, and enforce revisions and atomic writes; tests include non-admin and admin user isolation but have not been executed in the Mac or CI environment yet.
- **Not yet implemented:** Homepage session cookie/login UI, token issuance/admin management UI, central auth registry migration, configured persistent data root, CORS/cross-subdomain integration, legacy settings migration, NewsGo/MauriStocks clients, backup/restore and deployment. No existing application behavior changes in this branch.
- Security review remains necessary before exposing these handlers: session/cookie design, credential storage, origin rules, permissions and directory configuration.
- Validation note: attempted to clone the branch for local Go tests on 2026-11-02, but this environment could not resolve github.com. No tests were run; no visual diffs are applicable because no UI assets were changed.\n- Required acceptance before merge: run `go test -count=1 ./...`, verify application behavior on Mac/Safari and Cloudflare Tunnel where relevant, review file permissions and data migration, and obtain explicit owner approval for any merge/deployment.
- Do not change semantic `appVersion`, launch destinations, tunnel settings, tokens or running processes without owner authorization.
