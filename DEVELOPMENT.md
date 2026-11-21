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

## Mandatory suite authentication recovery candidate (Issue #60)

Before any owner launch/restart, run `./devbuild.sh` from a clean `auth/60-fail-closed-suite-recovery` checkout; it runs the existing Go suite, nested `shared/auth` tests, Node direct-recovery behavior tests, and both macOS builds. CI executes the same added nested and Node tests. In a fresh Private Safari window, open Homepage's Stocks card without signing in: the existing sign-in/guest choice must block launch. Direct `https://stocks.themauris.org/` entry with no verified session requires a consumer enforcement PR as well; test the Homepage bridge route `https://home.themauris.org/launch/stocks` to confirm it returns to Homepage's personal/guest choice instead of launching Stocks in local mode. Repeat NewsGo, SailingGo and ScraperGo bridge routes. Verified personal and guest handoffs must still launch with valid child tokens; expired/revoked personal sessions must not. Simulate a Homepage outage and verify that no app opens anonymously and the bridge provides a return/retry path. Confirm no raw tokens or private settings in console/log output. **Do not claim suite-wide enforcement until each consumer has a separately approved and tested integration.**

`shared/auth` uses an injected trusted `Verifier`; a consumer must validate the parent-bound Homepage child bearer server-side before calling `Guard`, and guard sensitive APIs against unsigned requests. Guests are allowed app use, denied personal settings. Review the actual API surface before migration to avoid breaking the initial no-token HTML bootstrap. `./homego.sh` remains a separate restart-and-tunnel operation requiring explicit authorization.

### Account chooser Safari acceptance — Issue #60

Open the **Sign in** dialog in a fresh Safari Private window. Confirm **Sign in as another user** opens the personal token form with a back-to-choices control; **Sign in as Guest** creates a verified temporary guest. With a guest session, the same Sign in control must remain accessible, the dialog must report current Guest status, and **Sign in as another user** must let the owner authenticate as a named account. A wrong token must not revoke the Guest session. Once a personal token is accepted, the account identity must change without moving guest preferences into personal account storage. When signed in as a named user, verify the same chooser remains available, and choosing Guest explains that sign-out is required first. Test **Sign out** from both Guest and named-user sessions; it must call server-side DELETE, invalidate the current session and its child handoffs, and restore the signed-out UI without erasing local preferences. Verify Escape, close, click-outside, keyboard focus order, hidden form, mobile sizing and light/dark themes; ensure dialogs do not flash or remain unexpectedly visible. `./devbuild.sh` now checks the account script's JavaScript syntax as well as the existing Go tests and builds. Do not expose test tokens or restart the live service without separate authorization.

**Single-button toolbar regression:** On a signed-out Homepage, Safari must show precisely one **Sign in** button (not **Sign in** plus **Sign in as Guest**). Click it and verify the shared chooser offers personal-user sign-in and Guest, plus sign-out only when already authenticated. Repeat signed in as Guest and as a named user. The current production Build 000088 predates this unmerged candidate; screenshots from that build cannot establish the candidate's UI. The existing Go chooser regression fails if a second guest toolbar button is reintroduced.

### Node.js prerequisite and stale-binary protection (Issue #60)

The owner observed on Mac that `./devbuild.sh` passed Go tests but then stopped at `node: command not found`. Because the script had not reached its compile or local build-counter stage, `./dist/homepage-darwin-arm64 -version` still reported **Build 000088**, and running `./homego.sh` launched that old binary. The candidate script now checks for `node` **before starting any tests**, with a clear error and instruction not to launch an old executable. Install a prebuilt Node.js binary/package appropriate to the Mac (avoid compiling Node from source), verify `node --version`, then run `./devbuild.sh` once. Confirm the output executable has a **new** build ID before considering the separately authorized service/tunnel restart. Do not infer new code is running merely because `git pull` succeeded.
