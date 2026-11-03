# Homepage current work

Date: 2026-11-03. Baseline: GitHub `main`. Active issue [#6](https://github.com/richard-mauri/homepage/issues/6), feature branch `feat/6-hourly-suite-token` (draft PR in progress).

Implemented candidate: one-hour random handoff bearer issued by Homepage with `POST /api/handoff/token`, from its existing session or immediately after approved-token sign-in. Handed-off bearer uses fragment on NewsGo and MauriStocks links and Homepage browser tab sessionStorage; the server stores token digests only in memory, checks expiry and approved-token digest/role on use. Homepage account profile and storage endpoints accept that bearer under their origin policy. Existing cookie flow retained. Regression tests cover expiry, revocation and storage. No changes to appVersion or private data.

Follow-up: NewsGo and MauriStocks client draft branches must also be deployed for end-to-end verification. iOS Safari acceptance required for same-tab launch, reload, expiration, loss of cookie and account isolation. Server restart clears one-hour bearer records; subsequent login/handoff required. CI and owner acceptance pending. No merge/restart authorized.

Mac test:
```
cd ~/projects/homepage
git fetch origin
git checkout feat/6-hourly-suite-token
git pull --ff-only origin feat/6-hourly-suite-token
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```
