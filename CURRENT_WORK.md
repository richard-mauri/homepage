# Current work — Homepage

## Issue #41 — theme the Homepage session recovery bridge

- **Merged baseline:** GitHub `main` commit `29e12ce` includes PR #37 (guest revocation and verified launch activity). No existing open PRs when this task began. New short-lived branch `fix/41-theme-session-recovery` from current `main`, Issue https://github.com/richard-mauri/homepage/issues/41; draft PR to be opened.
- **Observed Safari defect:** Owner's screenshot shows Homepage's unstyled white `/launch/sailing` bridge during SailingGo → full details navigation. SailingGo's details link is same-origin `/report?...` and normally should retain its tab-scoped handoff. Companion SailingGo Issue #23 / draft PR #24 addresses a guest identity/local preferences failure that can trigger unnecessary recovery. This Homepage task fixes the white bridge presentation even when a legitimate direct-visit recovery is necessary.
- **Change:** Add an external, same-origin `/static/launch-suite.css` with Homepage dark default and saved light-theme support, accessible contrast and keyboard focus; load it from the existing recovery HTML. Keep the bridge script's fixed destination map, POST-only handoff, session ID/cookie fallback, strict CSP, token fragments and navigation entirely unchanged.
- **Checks:** JavaScript syntax and Git whitespace checks passed locally. Go tests/build and exact-head GitHub CI must be checked on the draft PR. Mac Safari live-bridge and light/dark confirmation are owner-side tests, not yet performed. No semantic version, credential, tunnel, release, merge or process restart authorized.

## Owner Mac acceptance

```sh
cd ~/projects/homepage
git status -sb
git fetch origin
git checkout fix/41-theme-session-recovery
git pull --ff-only origin fix/41-theme-session-recovery
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-arm64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

Do not run `homego.sh` without a separate restart approval. Check the Homepage recovery bridge in Safari for named and one-hour guest session, both themes and the manual local-mode link. Confirm NewsGo, ScraperGo, Stocks and Sailing fixed launch destinations and no automatic renewal.
