# Homepage development and validation

GitHub [main](https://github.com/richard-mauri/homepage/tree/main) is authoritative for merged source. Read [AGENTS.md](AGENTS.md), [CURRENT_WORK.md](CURRENT_WORK.md) and [ARCHITECTURE.md](ARCHITECTURE.md) before changes. Start each feature/fix/docs branch from the newest `main`, keep it focused and open a **draft PR** against `main`. Record real issues for bugs and include reproduction, expected/observed behavior and regression risks. Do not broaden scope silently.

No merge, release, tag, `appVersion` change or tunnel/DNS modification without owner authorization. After a confirmed merged PR, check branch safety and only then delete that scoped branch; never delete `main` or a branch with unmerged work.

## Runtime-change Mac acceptance

For a scoped runtime-change branch (substitute its exact name), the owner runs:

```bash
cd ~/projects/homepage
git fetch origin
git checkout ACTUAL-PR-BRANCH
git pull --ff-only origin ACTUAL-PR-BRANCH
go test -count=1 ./...
./devbuild.sh
./dist/homepage-darwin-amd64 -version
./homego.sh
curl -fsS http://127.0.0.1:8787/healthz
```

The dedicated `devbuild.sh` already tests and compiles both macOS architectures; do not add a separate manual build to this sequence. Verify the stamped `-version` output, then open both the local site and https://home.themauris.org/ in Safari. Check desktop/mobile layout, theme, clock, four external links, and that other tunnels/services remain running. This is owner-side acceptance. Linux builds and tests cannot establish it.

Normal development uses `./devbuild.sh` only, with no release flag. It stores an increasing six-digit local build counter under Git's private directory, independent of semantic version; never reset or commit the counter. `homego.sh` uses the Intel executable and a private tunnel token. Build output and token are untracked. No automated GitHub release process is currently implemented.

Documentation-only PRs require Markdown, links, command/path review and relevant CI; they do not require a local restart. Inspect CI on the exact PR head; if no CI is configured, state that rather than claiming it passed. Report checks individually (passed/failed/not run). Update PR description and [CURRENT_WORK.md](CURRENT_WORK.md) before handoff.
