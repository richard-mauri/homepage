# Current work — Homepage

## Baseline

Homepage is the independent Go launchpad and shared policy authority. Read the [current merged suite policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md) first. This snapshot describes the intended policy v1.3.0 baseline; confirm actual integration, Issues, PRs and CI in GitHub.

The suite comprises six independent projects, including Sailing GPS Map (integration branch gh-pages). Suite membership does not change runtime access, launcher cards, authentication, deployment or storage. Policy changes are documentation-only; application versions, build counters, runtime cache revisions and operational services remain unchanged.

## Outstanding work

- Sailing GPS Map: verify its standard documents and adoption of the shared policy in a separately authorized consumer task.
- SailingGo: Issue #45 (public street tile proxy), followed by dependent Sailing GPS Map Issue #46; neither is implemented by Homepage policy work.
- Homepage: historical runtime-documentation reconciliation and residual branch cleanup require separate scope/approval.

## Operational constraints

Homepage runs on the owner's Mac with its own loopback listener, persistent per-user data and Cloudflare Tunnel. Preserve private credentials, local changes, browser/session state, data and Git-private build counters. The owner checks local divergence and performs applicable Mac/Safari acceptance; remote agents must not infer it from CI. Follow DEVELOPMENT.md for app procedures. No deployment or restart without separate authorization.
