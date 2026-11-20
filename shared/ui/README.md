# Homepage-owned suite UI module

Module: `github.com/richard-mauri/homepage/shared/ui`. This remains an ordinary
versioned Go build dependency; it does not connect consumers to Homepage at
runtime. The original `ui.ButtonCSS` remains unchanged.

## Shared tokens and typography

`ui.DesignCSS` contains namespaced, opt-in `.suite-ui` design tokens and
typography helpers. It uses Inter **when locally available**, with a platform
system-font fallback. No remote font import is made. Bundling a licensed WOFF2
font file remains a separate review item before claiming pixel-identical
typography across platforms.

## Banner

For Homepage's **original** blurred teal ambient background, load `ui.DesignCSS`
(directly or through `banner.CSS(ui.DesignCSS)`) and place these two
noninteractive elements immediately inside the page's `<body>`:

```html
<div class="suite-ambient suite-ambient-one" aria-hidden="true"></div>
<div class="suite-ambient suite-ambient-two" aria-hidden="true"></div>
```

The fixed shapes retain Homepage's exact original dimensions, positions, blur,
and `--glow` values (where defined); other apps use corresponding light/dark
fallbacks. The shapes are hidden on print. No application is forced to use them.

The independent `banner` package renders escaped HTML and delivers its embedded
CSS. The host must provide an already-verified account state and display name.
The renderer does not perform login, logout, renewal, session recovery, data
storage, or administrative checks.

```go
import (
  "github.com/richard-mauri/homepage/shared/ui"
  "github.com/richard-mauri/homepage/shared/ui/banner"
)

// Register once; the stylesheet is served from THIS application's origin.
http.Handle("/suite-ui/banner.css", banner.Stylesheet(ui.DesignCSS))
http.Handle("/suite-ui/status.js", banner.StatusScript())

// Render inside the existing HTML page/body template:
err := banner.Render(w, banner.Config{
  AppName: "NewsGo", Version: "0.2.58", Build: "000160",
  HomeURL: "https://home.themauris.org/",
  AccountState: banner.SignedIn, AccountName: verifiedName,
})
_ = err
```

`HomeURL` accepts a same-origin absolute path or the fixed
`https://home.themauris.org/` destination. `TannerURL` is restricted to an
origin-local absolute path; consumers can continue to host their existing
Tanner imagery themselves. These restrictions avoid arbitrary external links.

Consumers must link `/suite-ui/banner.css` on their own origin, serve it with
`style-src 'self'` allowed by their existing CSP, and supply optional Tanner
image paths from their own locally served assets. If no image is supplied, a
simple `M` mark is displayed.

For dynamic identity updates, load `/suite-ui/status.js` before the app script
and call `SuiteBannerStatus.signedIn(verifiedName)` after verifying the session.
For apps that already embed their JavaScript inline, prepend the exact shared
`banner.ScriptJS()` bytes to the existing script instead. The shared formatter
enforces the display text `Signed in as <username>` without role/source suffixes.
Only presentation is shared: each app continues to own authentication, session
expiration, guest behavior, and admin authorization. Do not duplicate the
display string in app-specific JavaScript.

## Rollout

The initial module change is additive. No existing consumer page is migrated.
Consumers should adopt a reviewed immutable nested-module version in separate
PRs, test light/dark and narrow layouts, and preserve all auth behavior. See
Homepage's `SUITE_UI_DESIGN_SYSTEM.md` and draft planning PR #53.
