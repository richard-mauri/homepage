# Current work

**Authoritative source:** GitHub `main` contains merged Homepage PR #18 for hashed durable 12-hour sessions and browser recovery. The owner reports that the same-tab round trip Homepage → MauriStocks → Back to Homepage **still** displays **Sign in**. No exact failing HTTP status or browser state has been captured, so root cause remains open under [Issue #17](https://github.com/richard-mauri/homepage/issues/17).

**Current scoped work:** `fix/auth-trace-homepage-20261105`, draft PR pending. Add an opt-in `-trace` command-line flag to log redacted auth request outcomes and Homepage browser navigation/UI state transitions to the existing `homepage.log`. This captures whether Safari has a stored Homepage session, whether the server rejects it, whether the cookie fallback succeeds, and what UI decision follows. The browser telemetry receiver is disabled by default, same-origin, bounded, and strictly allowlisted.

**Validation:** Go regression tests check trace gating, no raw-token leakage, strict same-origin input validation and critical browser instrumentation. Exact PR head CI and owner-side Safari/tunnel reproduction are separate acceptance gates.

**Handoff:** Review scoped branch and draft PR. Owner chooses when to build and restart Homepage in trace mode; preserve the old log first, since `homego.sh` redirects via `>` and truncates it. After testing, inspect `grep '\\[AUTH-TRACE\\]' ~/projects/homepage/homepage.log | tail -n 120` and identify the actual failed layer. Do not merge, release, tag, restart, change tunnel/DNS, or touch sister services without explicit authorization. Follow [DEVELOPMENT.md](DEVELOPMENT.md) for complete owner-side Mac commands.
