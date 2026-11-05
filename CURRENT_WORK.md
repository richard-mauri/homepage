# Current work

**Authoritative source:** GitHub `main` is the merged baseline. The previously outstanding feature/fix issues and PRs were closed as of 2026-11-04; do not interpret old handoff notes as active work.

**Active change now:** documentation-only reconciliation on `docs/current-suite-baseline-20261104`, [draft PR #16](https://github.com/richard-mauri/homepage/pull/16) targeting `main`. The goal is to make the five standard Markdown files accurately describe the current implemented suite. There is no active runtime feature/fix PR in this repository.

**Current implementation:** Homepage owns the fixed 12-hour sign-in session and on-click app handoff. All four cards navigate in the same tab. Signed-out NewsGo/Stocks launches explicitly enter browser-only mode. The 12-hour implementation is merged in Homepage PR #12, NewsGo PR #53 and MauriStocks PR #9.

**Next action:** review the documentation diff, check Markdown links and the PR's exact-head CI status, correct any factual errors, then request explicit owner approval to merge this documentation-only PR. Mac/Safari runtime retesting or app restart is not required for text-only changes. No merge, tag, release, deployment, branch deletion or tunnel change has been approved for this PR.
