# Homepage current work

**Date:** 2026-11-03. **Status:** Mobile header layout fix in draft; not merged or deployed.

- GitHub `main` is the merged baseline. Scope: [issue #4](https://github.com/richard-mauri/homepage/issues/4), branch `fix/4-mobile-header-layout`, [draft PR #5](https://github.com/richard-mauri/homepage/pull/5).
- Problem observed on iPhone Safari: a signed-in admin username was clipped at the top, and the header separator overlapped the light/dark theme button.
- Root cause: at viewports 600px wide or smaller, the header was fixed at 86px tall while its injected authentication controls wrapped onto multiple lines.
- Fix: use content-driven mobile header height and wrapping; put the signed-in account controls on a full-width row, keep the username in view, and position clock/theme controls above the header's bottom divider. Desktop layout, sessions, storage, portrait, and other applications untouched.
- Validation: GitHub CI pending; final iPhone Safari signed-in/signed-out, admin/ordinary user and theme checks pending. No claim of deployment.
- Owner's Mac verification:

```bash
cd ~/projects/homepage
git fetch origin
git checkout fix/4-mobile-header-layout
git pull --ff-only origin fix/4-mobile-header-layout
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

Do not merge, release, or restart any service without owner approval.
