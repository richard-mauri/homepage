# Suite development policy

**Policy version:** 1.2.0
**Editable authority:** [richard-mauri/homepage](https://github.com/richard-mauri/homepage), `SUITE_DEVELOPMENT_POLICY.md` on merged `main`  
**Coordination:** [Homepage Issue #45](https://github.com/richard-mauri/homepage/issues/45)

This is the single source of *shared development and agent workflow rules* for the five independent applications: [Homepage](https://github.com/richard-mauri/homepage), [NewsGo](https://github.com/richard-mauri/newsgo), [MauriStocks](https://github.com/richard-mauri/mauristocks), [ScraperGo](https://github.com/richard-mauri/scrapergo), and [SailingGo](https://github.com/richard-mauri/pittsburg-saildata). No sixth standards repository is needed. The applications, their GitHub repositories, processes, persistent data, listeners and Cloudflare tunnels remain independent. This policy changes no runtime behavior.

## Authority, precedence, and agent entry

For any development session, read the **actual text** of this policy before editing, along with the target repository's `AGENTS.md`, `DEVELOPMENT.md`, `ARCHITECTURE.md`, `CURRENT_WORK.md`, and `README.md` from the working branch. Verify GitHub's current `main`, Issues, branch, PR, head SHA and checks; `CURRENT_WORK.md` is a current-state handoff, not evidence that GitHub still has that state. Review owner-local uncommitted and unpushed work when accessible. Do not overwrite it.

This file controls shared rules: branches, Issues, PRs, data/security safeguards, verification, handoff, approval, FINISH and branch lifecycle. Each repository's local documentation controls its own architecture, implementation, runtime startup, builds, platform testing, data migration and operations **provided it does not contradict shared policy**. `README.md` serves users; `ARCHITECTURE.md` records implemented contracts; `DEVELOPMENT.md` supplies app-specific procedures/commands; `AGENTS.md` is the mandatory entry point; `CURRENT_WORK.md` is one factual current snapshot, not a running historical diary. History belongs in Git. Read all five; edit only those affected and disclose review exceptions.

An exception to shared policy requires explicit owner approval recorded in the relevant Issue/PR and, if durable, a policy change in Homepage. A conflicting local instruction is not an implied exception. Stop and ask if instructions conflict or policy access/provenance is uncertain.

### Single-source access; offline behavior

**Only Homepage `main` holds this policy.** Consumer repositories must not check in generated, manually copied or cached policy files. Each consumer's `AGENTS.md` must identify the exact Homepage policy path and require agents to retrieve and read its **current full text** from `main` before changing code, documentation or workflow. A URL alone is an entry point, not evidence that the policy was actually read. Agents must also read the consumer's five local app documents.

If the Homepage policy or its current merged-main status cannot be accessed or verified (including offline development), **stop development changes and promotion** until access is restored or the owner explicitly approves a documented exception. Do not use a previously remembered, locally cached or copied policy as an authoritative fallback. Local application inspection and non-mutating diagnostics may continue. This limitation is intentional and accepted by the owner in Homepage Issue #45.

No automated synchronization, generated consumer copies, copy-version metadata or cross-repo write permissions are needed. Policy updates use a Homepage-scoped branch and PR; consumer entry points ordinarily remain unchanged. The active merged Homepage policy, not a development-branch proposal, governs until a new policy PR receives owner-authorized FINISH.

## Language, dependencies, and development environments

Use Go for new backend services, standalone utilities, data acquisition, file conversion, and command-line tools. Use JavaScript, HTML, and CSS for browser functionality, preserving the application-specific architecture. This rule also applies to ad hoc utilities and development helpers prepared for suite projects, whether or not committed to a repository.

**Do not introduce or use Python for new development, scripts, utilities, or dependencies without explicit owner approval.** Prefer the Go standard library and existing installed tools; minimize unnecessary third-party dependencies. Do not install, replace, or reconfigure a functioning development toolchain without a demonstrated technical requirement and explicit owner approval.

Docker, containerized development and deployment, and virtual environments are permitted where appropriate. The Python restriction is not a restriction on virtualization, containers, or isolated development environments.

## Development workflow

1. Acknowledge verified merged `main`, actual active Issue(s)/PR(s), branch, scope, architecture boundaries, local divergence, expected checks, owner Mac/Safari/live acceptance and approval restrictions *before* edits. Reconcile existing open work first.
2. For a new scoped task use a short-lived Issue-linked branch based on current `main` and one **draft PR targeting `main`**. Never reuse a legacy `dev` or another work item's branch/PR. Track reproducible bugs in the affected application's Issue with observed/expected behavior, reproduction, build, diagnostics, redacted evidence and regression risks. Use `Fixes #N`/`Closes #N` only when completion is real and appropriate. If the owner directs related changes to be tested together in the **same active PR**, keep them integrated even when separately tracked; do not create a replacement PR. Seek approval before expanding scope or starting separate cross-repository consumer PRs.
3. Inspect the current tree and dependencies before edits. Preserve tests, app boundaries and security controls. Before commits review `git status` (including untracked files and ahead/behind), unstaged and staged diffs, `git diff --check`, the exact staged file inventory, and any commits ahead of origin. Never derive `git add` paths from an old manifest. No force pushes, destructive resets, checkout over dirty work or silent data migration.
4. Verify relevant tests/regressions, formatting and builds for runtime changes; for docs-only changes verify contents, relative paths, commands, cross-links, generated policy provenance and applicable CI. After push verify required GitHub Actions/checks against the **exact latest PR head SHA**. A stale green run does not count. If no CI exists or permission prevents inspection, say **unverified**, never passed. CI does not prove owner Mac, Safari, live publisher/NOAA behavior or tunnels.
5. Refresh the draft PR title/body and `CURRENT_WORK.md` before handoff. Record exact base/head, scope, changed files, Issue links, regression risk, tests, CI run/status, remaining Mac/Safari/live checks and approval boundary. Verify the saved PR description. Every status/handoff includes the six-part checklist below and exact local Mac commands appropriate to the change. Do not represent drafted, pushed, CI-tested, locally tested, merged, deployed or released as interchangeable statuses.

For owner Mac handoff commands, inspect the target application's `devbuild.sh` before composing the command sequence. If it already runs Go tests and builds, invoke that script once; **do not add standalone `go test` or `go build` commands** unless an additional diagnostic run is specifically justified. Independent GitHub CI tests remain mandatory.

### Data and operations safety

Preserve owner's local/uncommitted/unpushed files, subscriptions, exclusions, feed catalogs, watchlists, preferences, account records, per-user storage, browser history/session state, private credentials and tokens, private activity configuration, persistent data, Cloudflare/DNS/tunnel configuration and Git-private build counters. Do not print or commit secrets, cookies, identity records, logs, local binaries or private settings. Keep Homepage server authorization and per-user isolation; never infer permissions solely from UI. Application processes, tunnels and listeners are separate: do not restart or reconfigure them while changing policy. Avoid release-mode build scripts. Changes to live data, authentication, deployment, tags, release or restarts require separately explicit owner authorization. Source-controlled semantic version updates required by FINISH are authorized by that command as defined below; other version changes and changes to Git-private build counters require separately explicit owner authorization.

## Mandatory visible six-item checklist

For **every development status and handoff**, give a short factual account of:

1. **Baseline and scope:** verified GitHub `main`, Issue(s), actual branch/PR/base/head; whether related work belongs together in the same PR.
2. **Data safety:** what local changes, subscriptions, exclusions, preferences, accounts, settings, tokens, persistent data, browser history, tunnels and counters were checked/preserved; disclose inaccessible owner-local state.
3. **Documentation current:** review `README.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, `AGENTS.md`, `CURRENT_WORK.md` for accuracy; list edits and exceptions; verify access to the current Homepage policy and identify any access exception.
4. **Verification:** relevant regression/Markdown/link/path checks and exact-head CI run/result; distinguish passed, failed and unverified owner Mac/Safari/live/tunnel testing. For FINISH, report integration-branch version, candidate version, semantic increment classification, consistency of version declarations and relevant cache revisions; explicitly justify any documentation-only exemption.
5. **Local handoff:** complete copy-pasteable **Mac commands** with actual checkout directory, branch, clean-tree check, fetch/checkout/pull, and (for runtime changes) test/build/launch from that app's `DEVELOPMENT.md`; after FINISH, exact safe restoration commands. Use `git checkout` for the owner's Git, not `git switch`. Never force checkout over dirty work.
6. **Promotion authority and outcome:** specify approval and actual completion status separately for merge, associated Issue closure, remote branch deletion, deployment, tag and release. Never claim a step completed without observing it.

## Owner development-session commands: `BEGIN <project>` and `NEXT`

These commands apply to the five suite applications and supplement, rather than replace, the development workflow and six-item checklist above. Only the owner issues these commands. They do not waive safety, testing, CI, branch, PR or approval requirements.

### `BEGIN <project>` — establish the project context

The owner's `BEGIN <project>` instructs the agent to retrieve and read the **current merged-main** Homepage policy and the named project's `AGENTS.md`, `README.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md` and `CURRENT_WORK.md` in full. The agent must verify current GitHub `main`, relevant Issues, branches, PRs and checks; inspect available owner-local repository state, including uncommitted and unpushed changes; identify applicable tests, CI, operational constraints and approvals; and report a factual readiness assessment and any blockers **before editing**. A fresh clone is not evidence of the owner's other local working-tree state. If required policy access or a necessary repository inspection cannot be completed, stop and request direction rather than silently substituting a different environment or claiming verification. `BEGIN` alone does not authorize implementing an unspecified development phase, merging, deploying or restarting services.

### `NEXT` — authorize and execute the next development phase

The owner's `NEXT` **authorizes and instructs the agent to begin the next agreed, scoped development phase** for the active project. It is not a request for an already-completed phase's local handoff. If the next phase or active PR is ambiguous, ask before changing anything. Within the approved scope, implement the work, review and update affected regression tests and documentation, run applicable local checks, and push to the appropriate scoped draft PR. Continue correcting failed tests or CI within that scope; do not expand scope, weaken checks or erase tests merely to obtain a green result.

After pushing, verify the **required passing GitHub CI/checks against the exact latest PR head SHA** (identify the run and result; a stale, missing, pending or inaccessible run is not a pass). **Only after exact-head CI passes**, present the mandatory six-item checklist **as a table**, with statuses and concrete evidence, immediately followed by complete, copy-pasteable owner Mac local handoff instructions for clean-tree inspection, fetch/checkout/pull, build, test and run using the named app's `DEVELOPMENT.md`. Mark owner Mac/Safari/live acceptance as unverified until actually performed. If CI cannot be made to pass, report the blocker and do not present a successful local-testing handoff.

`NEXT` does not authorize the agent to execute a production launcher, restart a service or tunnel, deploy, merge, close Issues, delete branches, change versions or release. Identify the expected semantic version increment during NEXT and record it in the PR; FINISH authorizes and requires the actual update unless already separately authorized and applied. Supplying documented owner-run commands is not permission to execute them. The owner reviews local results and separately issues `FINISH` for an identified PR when ready.

## Owner command `FINISH`

Only the owner's explicit **`FINISH`** for a clearly identified active scoped PR authorizes its *complete closeout*, in this order:

1. Confirm exactly which active PR FINISH applies to, its `main` base (or the documented integration branch), current head SHA, required **passing exact-head CI**, merge safety/protection/reviews and absence of blocking issues. Stop and report if ambiguous, unavailable, failed or unsafe; never bypass protection.
2. **Mandatory pre-merge version update.** FINISH explicitly authorizes and requires updating the source-controlled application semantic version for every runtime-changing PR, before merging that PR. Compare the candidate to the actual current integration-branch version. Apply the appropriate Semantic Versioning increment: PATCH for backward-compatible fixes, MINOR for backward-compatible new functionality, MAJOR for incompatible changes; use the highest applicable classification. **Documentation-only FINISH: do not increment any application version or build counter.** Determine this exemption from the complete PR diff and its runtime effects, not the repository that hosts the documents. In particular, editing SUITE_DEVELOPMENT_POLICY.md in Homepage does not constitute a Homepage runtime change and must not increment the Homepage application version. The policy document's own revision identifier is independent of all application versions. For documentation-only PRs, verify that application version declarations, build counters and runtime cache revisions remain unchanged; record the exemption in the PR and FINISH report. Mixed documentation/runtime PRs follow the runtime version rule. If the candidate already contains the correct increment, verify it rather than incrementing twice. Update all authoritative version declarations and affected cache revisions consistently according to the project's documented version mechanism; do not modify Git-private counters. Record old/new version, classification, and evidence in the PR and CURRENT_WORK.md. Commit and push any necessary version/documentation updates to the same scoped PR branch, run relevant tests, and verify required passing GitHub CI against the resulting **exact latest head SHA**. Reconfirm merge safety and obtain renewed owner acceptance if anything beyond version metadata or its necessary cache/documentation synchronization changes, or if existing protection or project policy requires it. Do not merge with an unchanged, invalid, inconsistent or unverified required version. If the version mechanism or classification is ambiguous, stop and ask. No separate authorization is required for the narrowly scoped FINISH version update.
3. Merge **that PR only** using an allowed normal method; verify it actually merged into `main`.
4. Close **only the completed Issue(s) associated with that PR**; verify each result. Do not automatically close the central suite coordination Issue while related consumer work remains.
5. Check that the exact short-lived **remote head branch** is merged and has no open PR or unique unmerged work depending on it. Delete that branch only, without force, or verify GitHub already removed it. Do not delete `main`, retained release/archive branches, unrelated/historical branches, or anything ambiguous.
6. Provide these exact safe owner Mac commands with the **actual repository directory and merged branch** substituted. Stop if `git status` shows uncommitted work; never discard it:

```bash
cd ~/projects/ACTUAL-REPO
git status --short --branch
# Stop here if any uncommitted changes are shown.
git fetch --prune origin
git checkout main
git pull --ff-only origin main
git branch -d ACTUAL-MERGED-BRANCH
git status --short --branch
```

Report the observed merge commit, previous and merged semantic versions (or justified exemption), associated Issue states, remote cleanup, CI and owner acceptance status, and any blocked/failed operation. FINISH authorizes the required source-controlled semantic version update described above, but is **not** consent for deployment, tag, release, Git-private build-counter changes, DNS/tunnel changes, other PRs or branches, force operations, or local data loss. Separate owner approval is needed for historical/bulk branch cleanup. No automatic promotion.

## Consumer adoption and ongoing maintenance

Consumer repositories keep minimal `AGENTS.md` entry points linking to this policy and application-specific `DEVELOPMENT.md`, `ARCHITECTURE.md`, `CURRENT_WORK.md` and `README.md` documents. Never duplicate shared workflow paragraphs as an alternate authority. The owner reviews each consumer's adoption PR and the Homepage policy amendment as one coordinated documentation set, while each repository remains independently mergeable only after required CI and explicit scoped FINISH authorization.

A policy change in Homepage takes effect when merged by the owner; on their next session, agents in every consumer retrieve its current `main` text. No scheduled copy drift is possible because no consumer copy exists. If policy retrieval fails, the work stops as described above. These documentation changes do not authorize runtime, tunnel, deployment, data migration, release or tag operations.
