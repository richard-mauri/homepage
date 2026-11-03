# Homepage current work

Date: 2026-11-03. Baseline: GitHub `main`. Active issue [#6](https://github.com/richard-mauri/homepage/issues/6), feature branch `feat/6-hourly-suite-token` (draft PR in progress).

Implemented candidate: one-hour random handoff bearer issued by Homepage with `POST /api/handoff/token`, from its existing session or immediately after approved-token sign-in. Handed-off bearer uses fragment on NewsGo and MauriStocks links and Homepage browser tab sessionStorage; the server stores token digests only in memory, checks expiry and approved-token digest/role on use. Homepage account profile and storage endpoints accept that bearer under their origin policy. Homepage retains a host-only login cookie solely for its own sign-in UI and token issuance. Cross-app cookie authentication is removed; suite profile and storage requests from NewsGo/MauriStocks require the one-hour bearer. Regression tests cover expiry, revocation and storage. The server CLI flag `-handoff-minutes` sets bearer lifetime (1–1440 minutes, default 60); `./homego.sh -handoff-minutes 5` forwards this flag for a short browser test. It only affects newly issued handoff tokens, not Homepage's own login-session duration. No changes to appVersion or private data.

Owner's first Safari trial showed both applications failing account identity despite Homepage sign-in. The root issue was Homepage's read-only profile endpoint replying 405 to the CORS OPTIONS preflight triggered by the Authorization header. This candidate now handles trusted GET+Authorization OPTIONS and includes success/rejection tests. Owner must rebuild Homepage and relaunch the apps from its cards to verify.

Follow-up: NewsGo and MauriStocks client draft branches must also be deployed for end-to-end verification. iOS Safari acceptance required for same-tab launch, reload, expiration, token expiry and per-user account isolation. Server restart clears one-hour bearer records; subsequent login/handoff required. CI and owner acceptance pending. No merge/restart authorized.

Mac test:
```
cd ~/projects/homepage
git fetch origin
git checkout feat/6-hourly-suite-token
git pull --ff-only origin feat/6-hourly-suite-token
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh -handoff-minutes 5
curl -fsS http://127.0.0.1:8787/healthz
```

For the normal 60-minute behavior, restart the same launcher without the argument (`./homego.sh`). Existing browser account tabs will need a new token after the Homepage restart.

Latest candidate: Clicking a NewsGo or Stocks launcher card now issues a fresh handoff before same-tab navigation. Homepage `POST /api/handoff/renew` rotates an unexpired bearer from the app origin; profile returns `expiresAt` so browser UI can show actual time remaining. Rotation removes the old bearer. Five-minute and 60-minute CLI settings still apply. Direct app renewal does not use cookies or a Homepage redirect while valid. CI and iOS Safari acceptance must be rechecked on the latest head.

## Homepage return-after-expiration correction
Owner observed that NewsGo's expired-token Open Homepage link landed on a signed-out Homepage. Investigation found Homepage's browser UI could treat a still-valid short-lived app bearer as proof of Homepage login if GET `/api/session` failed; the condition went undetected until the five-minute bearer expired. The candidate now requires the independent Homepage session to validate before displaying a signed-in identity or confirming login. The host-only Homepage cookie is uniquely named `mauri_home_session` and expired old `mauri_suite_session` cookie variants are removed to avoid duplicate-name handling. The 12-hour Homepage session is still process-memory-backed; a Homepage server restart invalidates it, unlike a continuously running service. Owner should test returning from NewsGo after a five-minute expiration and inspect whether Homepage still shows signed in. No NewsGo changes required for the return link. Safari verification pending.
