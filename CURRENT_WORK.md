# Current work — Homepage

**Active:** [Issue #49](https://github.com/richard-mauri/homepage/issues/49), branch `docs/49-begin-next-commands` from GitHub `main` commit `c5b7694` (merged PR #48). A scoped documentation-only draft PR is being prepared against `main`. It defines owner commands `BEGIN <project>` and `NEXT` while preserving existing `FINISH`, verification, data safety and approval boundaries. It is a proposal, not effective shared policy until owner-authorized merge.

**Merged baseline:** Homepage PR #46 established the suite policy, and [PR #48](https://github.com/richard-mauri/homepage/pull/48) subsequently merged its single-source access amendment. No Homepage runtime, session, activity, listener or tunnel change is in this amendment. [Issue #38](https://github.com/richard-mauri/homepage/issues/38) remains a separate protected-API audit.

**Scope and safety:** Change only `SUITE_DEVELOPMENT_POLICY.md` and this current handoff. Consumer repositories continue to read Homepage merged `main`; no copied policy, cross-repository writes, app behavior changes, credentials, storage, counters, launcher or tunnel modifications. The cloud checkout started clean; owner Mac-local uncommitted/unpushed work could not be inspected and must be checked before local acceptance.

**Verification and next:** Review Markdown wording, cross-links, status and diff; verify exact draft PR head and applicable CI after push. As a docs-only change no local Mac build, restart or Safari retest is required. Await owner review and explicit `FINISH` for this scoped PR; no merge, deployment, restart, version bump, tag, release or branch deletion is authorized.
