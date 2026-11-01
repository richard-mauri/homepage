# Homepage current work

**Date:** 2026-11-01. **Status:** Feature branch candidate; not merged or deployed.

- Repository: https://github.com/richard-mauri/homepage
- Authoritative merged baseline: `main`; working branch: `feat/versioned-build-workflow`.
- Goal: introduce ScraperGo-style development documentation and a live footer version indicator, `-version` reporting, six-digit local build numbering, dual-architecture Mac builds/checksums and safe Homepage-only launcher.
- Five document files: `README.md`, `AGENTS.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, `CURRENT_WORK.md`.
- Runtime files: `main.go`, `devbuild.sh`, `homego.sh`, `.gitignore`.
- Intended behavior: keep the existing homepage and four destination URLs, default loopback listener, dedicated Cloudflare tunnel, and unrelated applications unchanged.
- Validation: local Linux `go test ./...`, Go cross-build to both macOS architectures, stamped CLI output, and shell syntax checks have been performed on the prepared candidate. GitHub PR CI, real Intel Mac launcher/Safari, live tunnel and four links are **not verified**.
- Next: owner checks out the branch using [DEVELOPMENT.md](DEVELOPMENT.md), validates on Mac/Safari, reviews draft PR, and explicitly authorizes any merge or release.
- Build number is local to each checkout; semantic `appVersion` remains `0.1.0`. No automated release command is included.
