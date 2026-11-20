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
simple `M` mark is displayed. No runtime JavaScript is needed.

## Rollout

The initial module change is additive. No existing consumer page is migrated.
Consumers should adopt a reviewed immutable nested-module version in separate
PRs, test light/dark and narrow layouts, and preserve all auth behavior. See
Homepage's `SUITE_UI_DESIGN_SYSTEM.md` and draft planning PR #53.
