# Homepage — agent instructions

Read this file, [CURRENT_WORK.md](CURRENT_WORK.md), [ARCHITECTURE.md](ARCHITECTURE.md) and [DEVELOPMENT.md](DEVELOPMENT.md) from the **actual working branch** before development. Confirm GitHub branch and PR state rather than trusting stale handoff notes.

Before action, tell the owner: authoritative `main` baseline, selected scoped branch and PR state, Issue/branch/draft-PR procedure, architecture constraints, test/CI expectations, complete Mac checkout/test/build/launch sequence for runtime changes, and Safari/live-tunnel verification requirements.

Maintain the existing launch destinations, loopback-only listener, embedded static assets, and isolation of each application's tunnel. Do not reveal credentials or commit `.cloudflared/`, build outputs or log files. Do not touch ScraperGo, Sailing, Stocks or NewsGo processes/configuration.

Use a short-lived branch from `main` for each change and a draft PR; bugs with wider impact get repository Issues. Keep scope narrow; do not merge unrelated tasks. Preserve tests and update docs with implementation changes. Report actual outcomes as passed, failed or unverified. CI is not Mac/Safari or live Cloudflare acceptance.

Never merge, tag, release, change `appVersion`, change DNS/tunnel credentials, or restart the owner's Mac application without explicit authorization. The owner decides acceptance.

Update [CURRENT_WORK.md](CURRENT_WORK.md) as a **single current portable snapshot** at meaningful milestones, retaining unresolved TODOs, but replacing stale active status. Before PR handoff refresh the PR title/body with the exact scope, commit, tests, CI status and unverified checks. Follow [DEVELOPMENT.md](DEVELOPMENT.md) for the complete Mac handoff block.
