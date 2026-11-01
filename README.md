# The Mauris — homepage

A responsive, dark-first personal launchpad for Sailing Outlook, Stocks, NewsGo and ScraperGo. Standard-library Go server, embedded static assets, zero Node or build tooling. Runs locally and can be published through a Cloudflare Tunnel. Includes light-mode toggle and Pacific-time clock.

## Quick start (macOS)

```bash
mkdir -p ~/projects
unzip homepage.zip -d ~/projects   # creates ~/projects/homepage
cd ~/projects/homepage
go run .
```

Visit http://127.0.0.1:8787. Requires Go 1.22+ and an existing Go installation. Build a standalone executable with `go build -o homepage .` and run `./homepage`. No third-party Go dependencies.

## App launch URLs

The cards launch these absolute HTTPS URLs by default:

| Card | URL |
| --- | --- |
| Sailing Outlook | https://weather.themauris.org/ |
| Stocks | https://stocks.themauris.org/ |
| NewsGo | https://newsgo.themauris.org/ |
| ScraperGo | https://scrapergo.themauris.org/ |

These links open in a new tab. To override an individual destination, set `SAILING_URL`, `STOCKS_URL`, `NEWSGO_URL`, or `SCRAPERGO_URL` in the server environment before starting it. No override is required for the above addresses. Environment variables in an existing LaunchAgent take precedence over the compiled defaults; remove or update any old placeholder values, then restart the service.

## Cloudflare Tunnel

1. In Cloudflare Zero Trust, reuse your existing **named** tunnel running on the machine hosting this server. Install `cloudflared` using a suitable prebuilt installer/binary and follow Cloudflare's current connector instructions.
2. Configure the existing tunnel's public hostname `home.themauris.org` to point to `http://127.0.0.1:8787`. If using a locally managed tunnel, adapt `deploy/cloudflared-config.example.yml` with the tunnel UUID and credentials file path, then set up the Cloudflare DNS route. If the tunnel is managed in the Cloudflare dashboard, configure the public hostname there instead; do not configure competing ingress rules.
3. Start the Go server and the tunnel. The homepage binds to loopback (`127.0.0.1`) by default and needs no inbound router port forwarding.
4. Optionally set up a Cloudflare redirect from `themauris.org` or `www.themauris.org` to `https://home.themauris.org`.

**Security:** Publishing the homepage does not automatically publish, proxy, or protect the linked apps. Use Cloudflare Access for private application hostnames and avoid exposing admin or scraper interfaces publicly without authentication. Cloudflare Tunnel on the host does not provide access control by itself.

## Keep it running on your Mac

You can use a macOS LaunchAgent for the built `homepage` binary. Provide a valid `ProgramArguments` absolute path plus `WorkingDirectory` and all desired URL environment variables. A sample template is in `deploy/org.themauris.homepage.plist.example`; replace `YOUR-USERNAME`, then save it as `~/Library/LaunchAgents/org.themauris.homepage.plist` and load with `launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/org.themauris.homepage.plist`. It runs when you are logged in. On a machine that must host the site while logged out, use a dedicated always-on host and appropriate service management instead.

## Git

```bash
cd ~/projects/homepage
git init
git add .
git commit -m "Initial The Mauris portal"
# To connect a remote later:
# git remote add origin git@github.com:YOUR-USER/homepage.git
# git branch -M main
# git push -u origin main
```

## Customization

Edit `static/styles.css` for colors/layout, `static/index.html` for text/sections, and `main.go` (`config()` function) for cards, labels and descriptions. A `go run .` restart picks up edits. The frontend retrieves card configuration from `/api/config`. Health endpoint: `/healthz`.

No external services are used by the Go backend. The frontend optionally fetches Google Fonts; it uses local system-font fallbacks if blocked.
