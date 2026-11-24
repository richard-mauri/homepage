# Current work — Homepage

**Policy amendment candidate:** [Issue #70](https://github.com/richard-mauri/homepage/issues/70), branch `docs/70-remote-agent-owner-local-roles`, based on merged `main` at `f18d206e60d075402b43454a7fa0eff8802a2e50`. Owner explicitly requested Issue and draft PR. Merged policy v1.2.1 remains authoritative until owner-authorized FINISH; candidate policy revision v1.2.2.

**Scope:** SUITE_DEVELOPMENT_POLICY.md and this handoff. Distinguish remote AI GitHub connector agent from owner-local Mac Git operator; define equivalent connector inspection, available execution checks and exact-head CI, with owner-local inspection and acceptance. No routine remote development stop or repeated NEXT for absence of a local clone. Preserve genuine safety and policy blockers. No runtime, app version, cache, private build counter or consumer changes.

**Existing work:** open PR #63 (Issue #62 NEXT semantics) overlaps policy wording but is separately scoped and must be reconciled against current main before either affected amendment is merged. PRs #59 and #53 are unrelated. Owner-local tree and unpushed commits are inaccessible to the remote agent; owner checks them before checkout. No data, credentials, services, tunnels or deployment configuration changed.

**Verification:** review wording, Markdown, paths, commit diff and exact-head GitHub Go CI. No runtime Mac/Safari acceptance or local build required for this documentation-only amendment. PR metadata is authoritative for final head and CI run.

**Owner Mac review:** `cd ~/projects/homepage`; inspect `git status --short --branch` and local-only commits, stop on dirty or conflicting work, fetch origin and checkout/pull `docs/70-remote-agent-owner-local-roles` only if local review desired. No build or restart required.

**Promotion:** explicit FINISH required; no merge, issue closure, branch deletion, deployment, tag or release authorized by this request.
