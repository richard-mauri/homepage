# Current work — Homepage

**Active documentation review:** [Issue #43](https://github.com/richard-mauri/homepage/issues/43), branch `docs/43-session-markdown-refresh` targeting `main`, [draft PR #44](https://github.com/richard-mauri/homepage/pull/44). This branch changes only Markdown; exact-head CI and owner merge approval are pending.

GitHub `main` at `2b53029` contains the merged session suite and themed `/launch/*` recovery bridge (Issue #41, PR #42). There are no open PRs or active implementation branches at this baseline. The owner verified guest handoffs, revocation and that Safari no longer flashes a white recovery page. **[Issue #38](https://github.com/richard-mauri/homepage/issues/38) remains intentionally open** for the suite-wide protected-API authorization audit; session-aware UI and activity reporting do not resolve that separate work.

Homepage owns 12-hour named sessions, one-hour guest sessions, child handoffs, renewal and disconnect. Same-tab launch links and CSP-constrained, themed `/launch/newsgo`, `/launch/stocks`, `/launch/sailing` recover valid Homepage sessions with POST and fragment-only handoffs, retaining manual browser-only fallback. The optional activity collector is diagnostic rather than authorization.

**Next:** reconcile Markdown with merged code via a documentation-only draft PR from `main`; review links/paths and exact-head CI. For runtime changes, use a scoped Issue/branch/PR and owner Mac/Safari/tunnel acceptance. No merge, release, version change, tunnel edit or restart without owner approval. Recheck GitHub and owner-local work at the next session.
