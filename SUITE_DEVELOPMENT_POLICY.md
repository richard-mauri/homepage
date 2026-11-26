# Suite development policy

**Policy version:** 1.3.0  
**Authority:** [Homepage main](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md)

## 1. Authority and membership

This is the sole shared development policy. Six independent applications, repositories, processes, data stores and deployment boundaries:

| Application | Repository | Integration branch |
|---|---|---|
| Homepage | [homepage](https://github.com/richard-mauri/homepage) | main |
| NewsGo | [newsgo](https://github.com/richard-mauri/newsgo) | main |
| MauriStocks | [mauristocks](https://github.com/richard-mauri/mauristocks) | main |
| ScraperGo | [scrapergo](https://github.com/richard-mauri/scrapergo) | main |
| SailingGo | [pittsburg-saildata](https://github.com/richard-mauri/pittsburg-saildata) | main |
| Sailing GPS Map | [sailing-gps-map](https://github.com/richard-mauri/sailing-gps-map) | gh-pages |

Suite membership alone changes no runtime, login, launcher, hosting, storage or deployment behavior. Preserve each app's architecture, including Sailing GPS Map's static PWA, browser/offline storage and Cloudflare services.

Read the **full current merged Homepage main policy** before development. A proposal, cached copy or remembered version is not authoritative. Consumer AGENTS.md files link to it; never duplicate it. If access/provenance cannot be verified after bounded retries, stop changes/promotion; non-mutating diagnostics may continue. An explicit owner exception must be recorded in the Issue/PR; durable exceptions require a policy amendment.

This policy governs shared workflow; local docs govern app-specific architecture, builds, tests and operations. If a local rule contradicts shared policy, shared policy prevails. Identify the exact conflicting instructions and resolution; if a genuine safety, scope or authority conflict remains, report the blocker and request a specific decision. Never silently ignore NEXT or invent ambiguity.

## 2. Repository documents and agent bootstrap

Every project maintains AGENTS.md (entry point), README.md (user behavior), ARCHITECTURE.md (implemented contracts), DEVELOPMENT.md (app-specific procedures), CURRENT_WORK.md (current checkpoint). Review all five; edit affected documents only. History belongs in Git.

Every development Issue MUST include the following bootstrap, with actual PR/branch links as soon as they exist:

> Read and follow the [current Homepage suite policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md). Read this Issue, its linked active PR and discussion, and the five project documents from the active branch, especially CURRENT_WORK.md. Reconcile source, commits, CI, owner decisions and next actions. Resume unfinished authorized work; disclose blockers.
>
> Active PR: (link or not yet created). Development branch: (link or not yet created).

The owner's instruction **Read and follow <Issue URL>** is sufficient to initiate BEGIN or recover an interrupted session. It does not authorize new scope or promotion.

CURRENT_WORK.md is persistent working memory, not an archive. Keep one factual snapshot: objective, approved scope, Issue/PR/branch/base, implementation and affected files, decisions/owner feedback, failed approaches to avoid, verification/CI, unresolved problems, exact next action, safety constraints and approval boundary. Record material progress before moving on; commit/push checkpoint updates with the associated work, including interrupted or failed work. Do not embed secrets. Record CI evidence by actual tested SHA; a checkpoint commit does not inherit a green run from its parent. Keep terminal CI result in PR metadata if recording it in the checkpoint would trigger an endless new-head cycle; the checkpoint must direct the next agent to that exact-head evidence.

## 3. Shared execution and safety

The remote agent uses approved GitHub connector operations; the owner operates the actual Mac checkout, browser, data and services. Lack of access to that Mac, a clone or staging area alone is not a blocker. Disclose unavailable checks. Never claim a local test ran unless observed. The owner checks dirty/untracked/unpushed work before checkout and performs applicable acceptance.

Use Go for new backend services, utilities, data acquisition, conversion and CLI tools; JavaScript/HTML/CSS for browser work. No Python development or helpers without explicit owner approval. Prefer standard libraries/existing tools; new dependencies or toolchain installation/reconfiguration require demonstrated need and owner approval where toolchain changes are involved. Containers and isolated environments are permitted.

Use a new Issue-linked short-lived branch from the verified integration branch and one draft PR targeting that integration branch. Reuse the active branch/PR for related work the owner wants tested together; never repurpose unrelated or legacy branches. Reproducible bugs belong in Issues with observed/expected behavior, reproduction, build, redacted diagnostics and regression risks. Link Issues accurately; close only completed work. Separate cross-repository work or scope expansion requires approval.

Before writes review actual tree, dependencies, exact base/head and intended diff. Local Git: review status (including untracked/ahead/behind), staged/unstaged diff, diff --check, exact staged paths and unpushed commits. Connector: inspect current source/head and proposed paths; verify actual committed diff/head afterward. Preserve tests and security controls; never weaken tests just to pass.

Preserve local work, subscriptions/exclusions, feeds/watchlists, preferences, accounts and per-user records, browser history/session state, offline archives, credentials/tokens, persistent data, private activity configuration, binaries, Git-private counters and DNS/Cloudflare/tunnel settings. Never expose or commit secrets/private logs or settings. Server-side authorization and per-user isolation remain mandatory where implemented; UI is not authorization. No force push, destructive reset, unsafe checkout, silent migration or unapproved deletion. Live data/auth changes, deployment, release/tag, restart and operational reconfiguration require separate explicit approval. No release-mode build scripts in ordinary development.

### Verification and recovery

Check relevant regressions, formatting/builds, or documentation content/links/paths/commands; use required GitHub CI for executable checks unavailable remotely. CI success must match the **exact latest PR head**. Missing, pending, stale or inaccessible CI is not a pass; CI does not establish owner Mac, Safari, upstream/live or tunnel acceptance.

During NEXT and FINISH the agent owns CI monitoring and correction. Use available synchronous polling/waiting until terminal results; pending CI alone is not a completion condition. For transient infrastructure failures retry at least three times where supported, with bounded increasing delays (e.g. 5/15/30 seconds), respecting rate limits. Verify remote state before retrying writes to avoid duplicate Issues/PRs/commits. If failure persists, capability/time limits prevent continuation, or a genuine permission/safety blocker exists, persist the checkpoint where possible and report the specific obstacle and required action. No background-work promises or repeated owner authorization for routine retries.

## 4. BEGIN — establish or recover

Only the owner authorizes scope. BEGIN <project> or an Issue bootstrap requires reading the merged policy and all five project docs in full (active branch if resuming), active Issue(s), PR description/discussion/reviews, actual source/commits, integration branch, exact head and checks. Inspect accessible owner-local divergence; otherwise assign inspection to the owner. Identify scope, tests, operational constraints and approvals before edits.

For an interrupted authorized phase, resume the existing branch/PR from the checkpoint without another NEXT. For new work, report readiness/blockers and propose the next scoped phase. BEGIN alone does not authorize unspecified implementation or promotion.

## 5. NEXT — execute and verify

NEXT accepts the immediately preceding clearly identified proposed phase; implement it without asking again. If scope/PR truly ambiguous, ask one targeted question. Implement, test, update affected docs, **maintain/commit/push CURRENT_WORK.md throughout**, and push to the scoped draft PR. Keep accepted related work together.

Monitor required CI, diagnose failures, correct, checkpoint, commit/push and recheck within scope until exact-head CI passes or a concrete blocker prevents continuation. Routine pending CI, failure correction or retry requires no new NEXT. Refresh and verify PR metadata with scope, changed files, risks, tests, exact base/head and CI evidence, remaining acceptance and expected semantic increment. Present the six-item checklist and complete owner handoff only as successful when required CI passes; otherwise report the blocker and actual status.

NEXT accepts development scope, not delivered quality or local acceptance. No merge, Issue closure, branch deletion, semantic version increment, deployment/restart/tag/release. Normal app-specific development build lineage is allowed; Git-private counters are not remotely modified. Owner-run launcher commands are not authorization to execute them.

## 6. FINISH — consolidate, verify, merge and clean

Only explicit owner FINISH for an identified PR authorizes this sequence:

1. Verify scope, integration base/head, required passing exact-head CI, owner acceptance, reviews/protection and absence of blockers. Do not bypass protection.
2. **Before merge, houseclean CURRENT_WORK.md** into the intended new baseline: remove completed development steps, superseded decisions, failed experiments, obsolete commands and resolved issues. Retain current application state, unresolved limits, operational constraints and outstanding work. Do not claim a merge/closure/deployment not yet observed; use PR/GitHub for those results.
3. Runtime PRs MUST update source-controlled semantic version relative to the actual integration version: PATCH backward-compatible fix, MINOR backward-compatible feature, MAJOR breaking change (highest applicable). Update all authoritative declarations and affected cache revisions consistently; verify an already-correct increment rather than double-bump. Record classification and old/new versions in PR and checkpoint as appropriate. Docs-only changes (including policy amendments) change no app version, build counter or runtime cache revision; verify and record exemption. Policy revision is independent. FINISH authorizes these scoped metadata/checkpoint edits; other substantive changes require renewed owner acceptance.
4. Commit/push FINISH changes to the same PR. Run applicable checks and monitor/retry through passing exact latest-head CI. Reconfirm merge safety and acceptance; stop if version/classification or safety cannot be resolved.
5. Merge only the authorized PR using an allowed normal method with verified head; verify actual merge and integration result. Close only completed associated Issues and verify closure. Do not close ongoing suite coordination while consumer work remains.
6. Verify the exact head branch is integrated (ancestry, or exact merged PR head and verified squash/rebase equivalence), unchanged and has no open PR dependency or unique unmerged work. Delete only that short-lived branch without force, or verify automatic deletion. Never delete integration, retained archive/release or unrelated branches. Historical/bulk cleanup requires separate approval. If connector deletion unavailable, provide exact owner remote-delete command only for verified-safe branch; report pending. If safety cannot be verified, stop.
7. Report observed merge SHA, versions/exemption, CI, acceptance, Issue and branch results and outstanding operations. Provide safe Mac restoration below. FINISH does not authorize deployment, restart, tags/releases, other PRs, data loss or private counter changes.

### Owner Mac commands

Use actual checkout directory, branch names and commands from DEVELOPMENT.md. Use git checkout (not switch). Stop on dirty state, failure or unmerged local work. For runtime handoff inspect devbuild.sh; if it tests/builds, invoke once, no redundant standalone commands unless diagnostically necessary. Separate module checks as applicable. Docs-only handoff requires review, not build/restart. Launch/restart only with separate approval.

FINISH cleanup template (substitute actual values; omit remote checks/deletion if verified already deleted):

```bash
(
  set -e
  cd ~/projects/ACTUAL-REPO
  git status --short --branch
  if [ -n "$(git status --porcelain)" ]; then
    echo "STOP: working tree is not clean" >&2
    exit 1
  fi
  git fetch --prune origin
  git checkout ACTUAL-INTEGRATION
  git pull --ff-only origin ACTUAL-INTEGRATION
  if git show-ref --verify --quiet refs/heads/ACTUAL-BRANCH; then
    git merge-base --is-ancestor ACTUAL-BRANCH ACTUAL-INTEGRATION
    git branch -d ACTUAL-BRANCH
  fi
  test "$(git rev-parse origin/ACTUAL-BRANCH)" = "VERIFIED-HEAD-SHA"
  git merge-base --is-ancestor origin/ACTUAL-BRANCH origin/ACTUAL-INTEGRATION
  git push origin --delete ACTUAL-BRANCH
  git fetch --prune origin
  git status --short --branch
)
```

Before remote deletion reverify no open PR dependency. For squash/rebase, replace ancestry guard only with verified exact-head integration equivalence checks; never force-delete local work. If local safety cannot be established, retain it and report. Remote deletion command is mandatory when safe but connector unavailable.

## 7. Six-item reporting checklist

For development status/handoff/FINISH report a concise factual table:

| Item | Required evidence |
|---|---|
| Baseline and scope | Integration SHA, Issue, actual branch/PR/base/head and accepted scope. |
| Data safety | Preserved data/work/config; inaccessible local state and owner responsibility. |
| Documentation | Five docs reviewed, changes/exceptions, authoritative policy access and checkpoint. |
| Verification | Tests/docs checks, exact-head CI link/result, unverified local/live acceptance; FINISH versions/classification/exemption. |
| Local handoff | Actual Mac commands appropriate to phase; clean-tree guard, fetch/checkout/ff-only pull, app-specific tests/build if runtime, separate launch authority; FINISH safe cleanup. |
| Promotion | Approval and observed merge, Issue, branch, deployment/tag/release status separately. |

Keep interim progress brief; do not present pending/failed verification as completed. Follow the final table with the appropriate complete copy-pasteable Mac command block.

Consumer adoption is separate scoped work; each repo links to the current merged policy and independently verifies CI and FINISH. No consumer policy copies or automated sync. Shared policy changes take effect only on merged Homepage main.
