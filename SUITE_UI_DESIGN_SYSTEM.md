# Suite UI design system — initial button proof of concept

**Owner:** Homepage, [Issue #47](https://github.com/richard-mauri/homepage/issues/47). This document governs appearance, not agent workflow. The separate [suite development policy](SUITE_DEVELOPMENT_POLICY.md) remains authoritative for branches, tests, approvals and release.

## Canonical button contract

NewsGo's merged, owner-corrected button behavior is the reference. All standard action buttons have a neutral dark resting fill, neutral one-pixel border, 9px radius, 9px by 13px padding, inherited font with weight 650, and readable foreground text. Large buttons increase padding to 12px by 17px without introducing a new appearance. Hover changes **text** and border to the accent, adds a temporary highlight and brightness; keyboard `:focus-visible` has an explicit two-pixel outline and offset. Neither an active panel nor a selected data filter receives a permanent accent border merely for being selected. Disabled controls are subdued and do not show a pointer cursor; reduced-motion users get no transition.

Homepage owns the implementation at [`shared/ui/buttons.css`](shared/ui/buttons.css), embedded by [`shared/ui/buttons.go`](shared/ui/buttons.go), in the nested Go module `github.com/richard-mauri/homepage/shared/ui`. The consumer defines four semantic variables (`--soft`, `--text`, `--line`, `--accent`) and applies `ui-button` to its action buttons; `ui-button-large` adds the large size. The Homepage Suite Management page temporarily scopes the same rule to `body.suite-admin button` so existing dynamically generated controls receive it too. That adapter is deliberate, not a second appearance definition.

Homepage serves the embedded CSS from `/suite-ui/buttons.css` under its same-origin CSP. Other consumers must serve the same embedded `ButtonCSS` bytes from their **own** origin, or inject those bytes into their own embedded HTML. Applications remain independent at runtime. Never reference the live Homepage site as a CSS CDN.

This initial contract does **not** standardize accordion `summary` headings, links, cards, input fields, layout or application-specific logic. Destructive actions need an owner-approved named variant before the generic button component is expanded; do not silently add permanent red or cyan borders to special buttons.

## Distribution and rollout

The nested module is independently addressable by Go, with module tags of the form `shared/ui/vX.Y.Z` after explicit owner approval. Consumers will pin a reviewed immutable tag or pseudo-version in their `go.mod`; updates require separate reviewable dependency PRs. No version tag is created as part of the initial draft. Homepage uses a local `replace` in its root `go.mod` so its own build consumes the adjacent module directly. Consumer `replace` directives pointing at another checkout must **not** be committed: they would break standalone builds.

Rollback means restoring the previous consumer `go.mod` version and UI integration in a new scoped PR. Avoid runtime network dependency, copy-and-paste CSS forks, or an unversioned shared CDN. Homepage's first UI integration is the Suite Management page; the public launch cards are navigation surfaces and have a different component contract.

## Five-app audit (initial inventory)

| Application | Current UI location and integration | Initial risk |
| --- | --- | --- |
| Homepage | `static/suite-admin.html` embeds admin CSS; `static/styles.css` styles launcher cards and theme control. | Admin action buttons used a separate teal rule. Homepage now consumes the shared button module for Suite Management. |
| NewsGo | `internal/dashboard/dashboard.go` embeds large inline CSS; `personal_script.go` adds `ui-button` to generated controls. | Migrate its already-approved button component without overriding placement-specific selectors. It is the first planned external consumer. |
| ScraperGo | `internal/ui/static/reader.css` is embedded with the reader template by Go. | One prominent reader action plus toolbar/account controls; test independent standalone asset delivery. |
| MauriStocks | Go-rendered frontend; detailed CSS selector audit outstanding. | Preserve chart and market-specific interaction controls during eventual migration. |
| SailingGo | `web/static/theme.css` and HTML templates. | Avoid changing map layers, overlays and mobile/weather controls when introducing shared buttons. |

This is a first-pass source inventory, **not** completed Safari visual inspection or approval for the other three app migrations. Before expanding beyond Homepage/NewsGo, record desktop/narrow/mobile dark/light hover, focus, active, disabled and touch results; include a reproducible visual comparison and owner sign-off. The full suite audit and showcase remain open under Issue #47.
