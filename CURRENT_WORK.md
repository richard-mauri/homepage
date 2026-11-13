# Current work — Homepage

## November 12, 2026 — collapsible suite administration UI

- **Merged baseline:** Homepage activity collector/dashboard PR #30 has been merged to `main`, Issue #28 closed, and its feature branch deleted. The other four applications' suite activity PRs are also merged, with their issues closed and branches deleted. The owner confirms five applications are running from `main`.
- **Current scoped development:** `feat/suite-admin-collapsible-panes` from current `main`. Homepage-only UI changes: rename the signed-in header's `Manage tokens` link to **Manage Suite**, preserving `/suite-admin`; group the page with native keyboard-accessible `<details>/<summary>` sections. Connect and Current session open by default; token administration is collapsed initially; Session activity opens initially with Hourly activity patterns collapsed. Nested Create user stays under administrator-only token administration. Existing token/session endpoints, session timer, admin visibility rules, activity polling/heartbeat, and browser storage remain unchanged.
- **Tests:** `TestSuiteAdminUsesCollapsiblePanesAndManageSuiteLink` checks panel markup, existing form IDs, admin visibility checks and launcher link target/label. Earlier activity/admin integration tests remain. The six native details elements have matching close tags. Exact-head Go CI and owner Safari verification are required; a code review or CI is not a live-browser test.
- **Next:** open/update draft PR, check exact-head CI, then owner fetches branch on Mac, runs `go test -count=1 ./...`, `bash ./devbuild.sh`, verifies `./dist/homepage-darwin-arm64 -version`, and deploys with `./homego.sh` only if desired. Verify via `https://home.themauris.org/suite-admin` in Safari on desktop/mobile, including keyboard toggles, non-admin visibility, current session renewal, token creation and activity polling. Preserve the locally modified `homego.sh` private activity environment loader and `private/activity.env`.
- **Restrictions:** no merge, tag, release, semantic version change, credential modifications, tunnel changes or production restarts without owner authorization.

## Other work

Keep unrelated issues and PRs independent. No changes to NewsGo, MauriStocks, SailingGo or ScraperGo in this scope.
