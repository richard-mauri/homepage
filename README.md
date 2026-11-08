# The Mauris — Homepage

Personal Go-powered launchpad published at https://home.themauris.org/ through its dedicated Cloudflare Tunnel. The site links to four **independent** applications:

| Card | Destination |
| --- | --- |
| Sailing Outlook | https://weather.themauris.org/ |
| Stocks | https://stocks.themauris.org/ |
| NewsGo | https://newsgo.themauris.org/ |
| ScraperGo | https://scrapergo.themauris.org/ |

The cards navigate in the current tab. When signed in, launching NewsGo or Stocks passes a session-bound Homepage handoff token; when signed out, those applications open in browser-only mode. Homepage manages the single 12-hour login session. While that login is still valid, NewsGo's **Renew** button can extend it to twelve hours from the renewal time without asking for the personal approved credential again. The operation rotates the NewsGo handoff bearer securely and remains subject to credential revocation; expired sessions must sign in again. Homepage validates its stored session ID when returning from NewsGo so renewed expiration is recovered. Override their targets with `SAILING_URL`, `STOCKS_URL`, `NEWSGO_URL`, and `SCRAPERGO_URL` if necessary.

## Development and local use

Source: https://github.com/richard-mauri/homepage. Requires Go 1.22+ and Git. The Go server uses the standard library and embedded assets; no Node toolchain is needed.

```bash
cd ~/projects/homepage
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
```

`devbuild.sh` tests and builds both `darwin/amd64` and `darwin/arm64` executables under `dist/`, generating `SHA256SUMS`. It assigns a six-digit build number stored at the private Git path `homepage-build-counter`; numbers increase for each build attempt after tests pass, are independent of version and branch, and are **local to the checkout**, not globally synchronized. Do not commit or reset that counter. Normal development does not change `appVersion` (currently `0.1.0`); a release/version decision requires owner approval. The `-version` flag prints `Homepage v0.1.0 · Build 000001` for an example stamped build. A plain `go run . -version` uses the unstamped `Build 000000`.

`homego.sh` is machine-specific: it checks the version stamp before restarting only the Homepage executable and its own `cloudflared` tunnel using the private `.cloudflared/token`. It uses `/Users/richardmauri/projects/homepage` and `~/Applications/cloudflared`; never commit the tunnel token. Use `http://127.0.0.1:8787/healthz` for local health checks. `LISTEN_ADDR` defaults to `127.0.0.1:8787`.

Cloudflare Tunnel's published hostname `home.themauris.org` points to `http://127.0.0.1:8787`. The `www.themauris.org` DNS record is not part of this application's deployment and should not be changed implicitly. No router port forwarding is required. Restrict sensitive linked applications with their own appropriate access controls.

## Repository reference documents

Read [AGENTS.md](AGENTS.md) and [CURRENT_WORK.md](CURRENT_WORK.md) before making changes. [ARCHITECTURE.md](ARCHITECTURE.md) specifies ownership and security boundaries; [DEVELOPMENT.md](DEVELOPMENT.md) governs Issues, branches, PRs, build, Mac acceptance, and releases.

The homepage footer is rendered by the Go server with the running executable's semantic version and exact build number, matching `-version`. `/api/config` also returns those values. The version is visible even if the browser JavaScript fails or is cached.

With `SUITE_DATA_DIR` configured, Homepage stores its 12-hour session state at `SUITE_DATA_DIR/auth/homepage-sessions.json`. It persists hashed session IDs and app handoff references, without writing approved personal credentials or raw session/bearer tokens. Keep that directory on persistent local storage with restrictive access. A Homepage process restart must not shorten an unexpired session, but sign-out and credential revocation still invalidate it. The service fails to start rather than silently discarding an unreadable state file; write errors are surfaced instead of reporting a successful login or logout. Existing sessions created by earlier in-memory-only versions cannot be recovered on first upgrade.\n\nThe browser assets live under `static/`. `main.go` owns HTTP setup, embedded assets, and the JSON card configuration at `/api/config`. The health check is `/healthz`. The homepage intentionally does not proxy or manage the four linked services.

## Opt-in authentication trace

For troubleshooting a premature **Sign in** display, start Homepage with `./homego.sh -trace` **only after authorizing the application restart**. The launcher passes the flag to the Homepage executable and redirects its output to `~/projects/homepage/homepage.log`. This is a temporary diagnostic option, not permanent debug logging. Normal `./homego.sh` starts without trace. The launcher **truncates** `homepage.log` on each restart: copy an important trace before restarting.

With `-trace`, Homepage logs `[AUTH-TRACE]` records for session validation requests, status codes, authentication transport (session header, app bearer or none), presence of the Homepage cookie, selected session rejection reason, and navigation/UI events reported by the Homepage browser. Events include page load, focus, page restoration, network reconnection, sign-in result, validated state, app launch and handoff result. Correlation uses an ephemeral random browser page ID, **not** a credential. Trace messages contain no raw bearer, session ID, approved token, cookie value, personal preferences or identity. The tracing endpoint `POST /api/trace` accepts only strict enumerated browser diagnostics from Homepage's origin and is unavailable without `-trace`.

After reproducing, inspect with `grep '\[AUTH-TRACE\]' ~/projects/homepage/homepage.log | tail -n 120`. Check whether the browser recorded `sessionStored=false`, whether the server returned `status=401`, and whether `reason=session_not_found`, `session_expired` or `credential_unapproved` appears. `status=200` paired with `ui_signed_out` would suggest browser UI state/race rather than server revocation. Copy only redacted trace lines when sharing diagnostics; do not share the complete private log without reviewing it first.
