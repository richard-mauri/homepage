# Unified suite UI — proposed architecture and migration map

**Status:** Proposal for owner review, not an implemented contract.  
**Scope:** Homepage Issue #52, documentation only.  
**Owner:** `richard-mauri/homepage`, nested module `github.com/richard-mauri/homepage/shared/ui`.

## Current baseline

The nested `shared/ui` Go module already exists, declares Go 1.22, and exports `ui.ButtonCSS []byte` embedded from `buttons.css`. Homepage uses a local `replace` to `./shared/ui`; NewsGo, MauriStocks, and ScraperGo pin the same November 18, 2026 pseudo-version. The first contract in `SUITE_UI_DESIGN_SYSTEM.md` intentionally covers standard buttons only. Its `body.suite-admin button` adapter is a migration bridge, not a pattern for general selectors.

Do not change the existing module path or remove `ButtonCSS`. Introduce the proposed APIs additively, initially with the existing stylesheet unchanged. SailingGo integration and precise selector inventory require an additional source audit before its consumer PR.

## Proposed module organization

```text
shared/ui/
  go.mod
  buttons.go             # Keep legacy public ButtonCSS API
  buttons.css            # Keep bytes stable during compatibility phase
  design/
    tokens.css           # Semantic color, spacing, radius, typography variables
    typography.css       # App chrome and default component text rules
    fonts/               # Locally embedded WOFF2 files, if approved/licensed
  components/
    buttons.css          # Future scoped component evolution
    forms.css
    feedback.css
  banner/
    banner.go            # Config, Render, embedded assets
    banner.html
    banner.css
  README.md              # API, assets, CSP, versioning and examples
```

Only create packages/files as they become necessary. Avoid premature abstractions and any dependency on Homepage's server, session, or storage packages. The package should remain Go-standard-library-only unless justified. `embed.FS` packages assets at build time; each application serves its own CSS/fonts under its own origin and CSP. No runtime Homepage CDN or filesystem coupling.

## Design tokens and typography proposal

Keep the legacy host tokens `--soft`, `--text`, `--line`, `--accent` compatible. Add explicitly namespaced tokens (`--suite-*`) for background, surfaces, border, foreground, muted text, accent, success/warning, spacing, radii, and typography. Provide both light and dark variants without overriding host `body` or application data colors.

Proposed font: Inter, self-hosted WOFF2 after confirming redistribution terms and rendering on owner Safari; fall back to system sans-serif. The canonical **chrome** scale is:

| Element | Desktop | Weight | Notes |
| --- | --- | --- | --- |
| App title | 28px | 700 | Mobile 22–24px |
| Suite eyebrow | 12px | 600 | Uppercase and tracked |
| Version/build | 12px | 400 | Muted, never hidden from diagnostics |
| Account identity / navigation | 14px | 600 | Verified account state only |
| Section heading | 22px | 650 | Mobile 20px |
| Body | 15px | 400 | Default 1.5 line-height |
| Buttons/inputs | 14px | 600 | Shared padding and focus treatment |
| Table heading | 13px | 600 | Tabular numbers where relevant |
| Table value | 14px | 400–600 | `font-variant-numeric: tabular-nums` |

These are review targets, not immediate global overrides. Homepage's hero display, NewsGo's editorial headline, MauriStocks' dense financial table, ScraperGo's article body, and SailingGo's maps remain application-controlled. Do not apply broad `h1`, `button`, `body`, `input`, or `table` selectors to all consuming pages.

## Common banner contract

Canonical structure: small suite identity and existing Tanner image at left, application name, muted semantic version/build, verified account label at right, with an explicit return-to-Homepage link on consumer apps. Mobile layout may stack these elements but must not truncate identity or obscure navigation. Homepage's banner omits the self-return link.

Suggested API (illustrative, not yet final):

```go
type Config struct {
    AppID       string
    AppName     string
    Version     string
    Build       string
    Description string
    HomeURL     string
    AccountLabel string // already verified or intentionally local/guest
    IsHomepage  bool
}
func Render(w io.Writer, cfg Config) error
```

The renderer escapes all supplied text and attributes, uses a fixed allowlisted/validated home URL, and has no authentication, session refresh, storage, telemetry, or JavaScript authority. Define account states carefully (signed in, signed out/browser-only, guest, temporarily unavailable); never infer a sign-out from a network failure. Render neutral/unknown states honestly and preserve each application's existing server-authoritative auth flow. The banner must work without Tanner's image and without client JavaScript.

The module embeds the template/CSS and optionally the licensed font assets. Expose an asset-handler/helper or bytes/`fs.FS` API for per-app routing; use `text/css; charset=utf-8` and correct font content types and caching. Ensure the banner stylesheet remains scoped to `.suite-banner*`. Do not make consumers fetch resources from the live Homepage origin.

## Application migration inventory and guardrails

| App | Existing UI | Scope of later consumer PR | Must preserve |
| --- | --- | --- | --- |
| Homepage | `static/styles.css`, `static/suite-admin.html`; shared button CSS already served locally | Common banner; token adoption for chrome; move Manage Suite/systems status/clock/theme control into an app toolbar only if reviewed | Homepage 12-hour session controls in Manage Tokens, admin-only permissions, same-tab launcher, hero and cards, tunnel/listener boundaries |
| ScraperGo | `internal/ui/static/reader.css` and embedded reader template | Replace masthead/header and load self-served assets | URL submission, reading and printing, article typography, access constraints |
| NewsGo | `internal/dashboard/dashboard.go` with substantial inline CSS; already serves `/suite-ui/buttons.css` | Replace header only, standardize chrome typography | Subscription/exclusion state, filters, selection, headline reader links and auth recovery |
| MauriStocks | Go-rendered page with embedded shared `ButtonCSS` | Replace header and normalize chrome; avoid table-wide overrides | Watchlist, dialogs/focus, refresh, sorting, numerical density, account preferences |
| SailingGo | `web/static/theme.css` and weather/map UI (prior suite design inventory) | Audit source and CSS first; then replace top chrome | Weather overlays/maps, controls, storage/handoff, mobile behavior |

Migration order proposed: shared-module additive PR; ScraperGo; NewsGo; MauriStocks; SailingGo; Homepage last. Seek separate owner approval before starting cross-repository consumer PRs. Do not combine migrations with auth, persistence, operational changes, or release bumps.

## Verification and acceptance gates

For the **shared-module PR**, use Go unit tests for escaping, missing/invalid config, rendered structure, embedded CSS/font integrity, content types and absence of network imports. Test CSS scoping and backward compatibility for `ButtonCSS`. Check `go test ./...` in both the root and nested module as applicable, exact-head CI, and docs/links. Ensure no dependency on a running Homepage server.

For **each consumer PR**, compare wide/narrow/mobile in both themes; hover, keyboard focus, reduced motion and touch; verify font loading and fallback; compare app-specific controls with baseline screenshots. Run relevant app tests, build, exact-head CI, and owner Mac/Safari/live checks before FINISH. In particular verify browser-only, signed-in, guest, transient-network, and revoked session presentation without changing existing authentication handling. Keep version/build displays and the existing runtime/launcher boundaries.

Each consuming app pins an immutable reviewed nested-module tag (e.g. `shared/ui/v0.2.0`) only after the owner explicitly approves the tag/release. A new module commit does not change already built consumers. Roll back an individual consumer via a scoped PR restoring its prior pinned version and prior header integration.

## Decision points for owner

1. Confirm Tanner placement (left as Homepage today), title/version/account alignment, and banner height.
2. Approve Inter licensing/font loading or choose the existing system-font stack.
3. Confirm which Homepage controls belong in its secondary toolbar.
4. Approve the asset-serving API after the first minimal banner implementation and tests.

This document is a plan only; no source, authentication, storage, service, deployment, build counter, or release change is authorized here.
