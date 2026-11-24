# Current work — Homepage

**Policy amendment candidate:** [Issue #68](https://github.com/richard-mauri/homepage/issues/68), branch `docs/68-finish-remote-cleanup`, based on merged `main` at `ff54c83b056b0fe744c5a8befaccca82a5ecf264`. Owner BEGIN/NEXT authorizes documentation-only clarification of FINISH remote branch deletion fallback. Merged policy v1.2.0 remains authoritative until owner-authorized FINISH; candidate policy revision v1.2.1.

**Scope:** SUITE_DEVELOPMENT_POLICY.md and this current handoff only. Require actual safe remote deletion command in owner cleanup block whenever direct deletion is unavailable and branch safety verified; conditional local deletion, clean-tree guard, actual integration branch, ancestry checks, prune and final status. No Homepage application version or private build counter increment, runtime/cache changes, restart, deployment, tag or release. No consumer copies or changes.

**Current GitHub state:** unrelated open PRs #63 (NEXT clarification), #59 (CORS), #53 (UI architecture) remain outside scope. Owner-local working tree and unpushed commits cannot be inspected; do not assume clean state. No owner data, tokens, account records, private activity configuration, tunnels or application processes touched.

**Verification:** review policy wording, Markdown, shell syntax and actual changed file inventory; required Go CI on exact final PR head. PR metadata is authoritative for final SHA/run and merge state. No runtime Mac/Safari re-acceptance necessary for docs-only amendment.

**Owner Mac review:** `cd ~/projects/homepage`; `git status --short --branch` (stop if dirty); `git fetch --prune origin`; checkout/pull `docs/68-finish-remote-cleanup` only if local review desired. No build or restart required. FINISH cleanup must substitute actual branch names and include remote deletion fallback when safe and pending.

**Promotion:** no merge, Issue closure or branch deletion authorized by NEXT. Current policy retains separate explicit FINISH requirement.
