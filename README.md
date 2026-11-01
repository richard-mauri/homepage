# The Mauris — Homepage

Personal Go-powered launchpad published at https://home.themauris.org/ through its dedicated Cloudflare Tunnel. The site links to four **independent** applications:

| Card | Destination |
| --- | --- |
| Sailing Outlook | https://weather.themauris.org/ |
| Stocks | https://stocks.themauris.org/ |
| NewsGo | https://newsgo.themauris.org/ |
| ScraperGo | https://scrapergo.themauris.org/ |

The cards open new tabs. Override their targets with `SAILING_URL`, `STOCKS_URL`, `NEWSGO_URL`, and `SCRAPERGO_URL` if necessary.

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

The browser assets live under `static/`. `main.go` owns HTTP setup, embedded assets, and the JSON card configuration at `/api/config`. The health check is `/healthz`. The homepage intentionally does not proxy or manage the four linked services.
