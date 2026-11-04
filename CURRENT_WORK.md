# Homepage current work — 2026-11-04

Authoritative merged baseline: GitHub `main`, including merged [PR #14](https://github.com/richard-mauri/homepage/pull/14) (Personal settings sign-in feedback). Active issue [#11](https://github.com/richard-mauri/homepage/issues/11), branch `fix/11-reusable-app-tabs`, [PR #12](https://github.com/richard-mauri/homepage/pull/12) (same-tab launch and optional browser-only access).

PR #12 is reconciled with the newer main by a merge commit that retains PR #14's Personal settings Sign in hover/pressed/busy feedback, timeout and token/session handling. Homepage card destinations now open in the same tab. Signed-out NewsGo/MauriStocks launches are not blocked or interrupted by native Safari alerts; signed-in launches still request a fresh per-app handoff bearer. A failed handoff allows a browser-only launch. Launcher card links never prepopulate cached handoff tokens.

Changes are limited to browser launcher behavior, the same-tab architecture contract, this current status and the launcher regression test. No server authorization changes, app semantic version change, private settings, credentials, tunnel/DNS or other application code changes.

GitHub Go CI must be verified on the reconciled PR head. Owner Mac/Safari acceptance is still pending, including signed-out launches for all four apps, same-tab Back to Homepage navigation, signed-in handoff and personal settings, expired/rejected sessions, failed handoff fallback, and PR #14 dialog hover/busy states in light/dark themes. No merge, tag, release or restart is authorized before owner acceptance.
