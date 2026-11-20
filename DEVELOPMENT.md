# Homepage development and validation

The shared Issue/branch/draft-PR, verification, six-item handoff, owner `FINISH` and branch-cleanup rules live in [SUITE_DEVELOPMENT_POLICY.md](SUITE_DEVELOPMENT_POLICY.md); read the **actual policy text** before development. Homepage's merged GitHub [main](https://github.com/richard-mauri/homepage/tree/main) remains authoritative. Verify current Issues, branches, PRs, exact head and local work before editing. The merged Homepage policy is effective for shared workflow; a policy-amendment draft does not supersede `main` before owner-authorized merge. Homepage-specific architectural boundaries are in [ARCHITECTURE.md](ARCHITECTURE.md).

## Runtime-change Mac acceptance

`./devbuild.sh` already runs `go test -count=1 ./...` and builds both macOS architectures; no separate test or build command is needed unless diagnosing a specific failure.

First check `git status --short --branch` and stop rather than overwrite uncommitted changes. For a scoped runtime-change PR the owner runs the following with its real head branch substituted:

```bash
cd ~/projects/homepage
git status --short --branch
# Stop if uncommitted changes are shown.
git fetch origin
git checkout ACTUAL-PR-BRANCH
git pull --ff-only origin ACTUAL-PR-BRANCH
./devbuild.sh
./dist/homepage-darwin-arm64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

`homego.sh` restarts Homepage and its dedicated tunnel: do **not** run it without separate owner authorization. `devbuild.sh` tests and compiles both macOS architectures; do not add redundant manual builds. The Apple Silicon owner Mac runs `dist/homepage-darwin-arm64`, not the Intel compatibility binary. Inspect the stamped `-version` output, then check local and https://home.themauris.org/ in Safari for desktop/mobile layout, themes, clock, four links, sessions, account behavior and independent sister services. CI cannot establish Mac/Safari, Cloudflare or live integration acceptance.

Normal development uses `./devbuild.sh` **without** a release flag. Its six-digit build counter is Git-private, monotonically increasing and independent of semantic version; never reset or commit it. `homego.sh` loads a private tunnel token and `private/activity.env`, which must stay untracked. The listener is loopback-only by default. No automated GitHub release process is implemented.

Documentation-only PRs need content, Markdown link/path, command and policy provenance review and applicable exact-head CI; they require no restart, build or Safari re-acceptance. Refresh PR title/body and [CURRENT_WORK.md](CURRENT_WORK.md) before handoff. After an authorized `FINISH`, use the shared policy's clean-tree-guarded restoration sequence with the exact Homepage branch; never force checkout or delete unmerged work.
