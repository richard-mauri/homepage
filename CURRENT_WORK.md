# Current work — Homepage

## Issue #22 — Homepage-owned renewal (November 9, 2026)

- **Authoritative baseline:** GitHub main at `b62f1e68556b049167bd62330fa52899e5b3df77` (Issue #20 / PR #21 merged).
- **Scoped work:** Issue #22, branch `fix/22-homepage-session-renewal`, draft PR #23. No changes to main, production tunnels, other apps, semantic version, or Mac runtime.
- **Behavior:** Manage Tokens (`/suite-admin`) Current session panel gets a Renew action and HH:MM:SS session countdown beside Disconnect, available to all authenticated users (the Homepage header has no Renew, timer or redundant Sign out). The Manage tokens link is visible to ordinary signed-in users, while Approved users remains admin-only. Disconnect is the sole sign-out action; after successful server revocation, it clears the Homepage browser's session and handoff caches. Renewal also refreshes the host-only HttpOnly session cookie's Max-Age. POST `/api/session/renew` checks the unexpired Homepage Session header and approved credential, extends the durable parent expiry to 12 hours from renewal, and extends only live parent-bound child handoffs. No cookie-only or cross-origin renewal.
- **Compatibility:** Existing NewsGo-origin `/api/handoff/extend-session` remains temporarily available so deployed NewsGo does not break. Fully removing NewsGo's Renew control requires a coordinated NewsGo change; it is not silently modified by this Homepage-only PR.
- **Security:** Existing expiry/revocation checks, per-user isolation, sign-out invalidation, atomic persistence/rollback, and original session identity remain enforced.
- **Validation:** Go CI run 42679741455 passed at code/test head ddd01ffc699e14427eac283ed42e6e32be0b1f1e (Go tests, Intel and Apple Silicon builds). Final CI must be checked for the latest PR head, including consolidated Disconnect behavior, role access and UI regression assertions. Local Mac/Safari and Cloudflare Tunnel acceptance remain unverified.
- **Owner acceptance:** After explicitly choosing to deploy this branch, sign in as both regular and admin users; check countdown, renew before expiry, inspect extended expiry after return to Homepage, confirm no renewal after expiration/revocation and no child handoff after sign-out. Verify all four cards, theme, and independent tunnels.
- **Merge boundary:** Draft PR only. Owner approval required before merge, version/release, or restart.

## Mac checkout, validation, and launch

```bash
cd ~/projects/homepage
git fetch origin
git checkout fix/22-homepage-session-renewal
git pull --ff-only origin fix/22-homepage-session-renewal
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

After authorized launch, test local Homepage and https://home.themauris.org/ in Safari (desktop/mobile), login, timer/renewal, other four app links, and cross-app handoff behavior.
