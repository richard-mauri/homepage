# Current work — Homepage

**Active:** [Issue #54](https://github.com/richard-mauri/homepage/issues/54), branch `ui/54-shared-banner` based on merged Homepage `main` `3f2fe48f7a7a889f5e6af5c764687ff9cf8e45b3`. The related documentation proposal is Issue #52 / draft PR #53 and remains separate. Verify the implementation PR and checks on GitHub rather than trusting this snapshot.

**Implementation candidate:** Add `ui.DesignCSS` opt-in typography and color tokens, a standalone `banner` package with embedded template/CSS, typed account display states, strict URL rules, and a locally mounted CSS handler. Keep `ButtonCSS` unchanged. Add focused tests and module documentation, with a pointer from `ARCHITECTURE.md`. None of the five application banners is switched over.

**Safety:** No authentication/session logic, accounts, activity, storage, subscriptions, watchlists, private credentials, launcher targets, live tunnels, build counters, deployment or release changed. The scratch checkout began clean; owner Mac-local uncommitted/unpushed work remains inaccessible. No service was restarted.

**Verification:** Check static diffs/format, nested-module and Homepage root tests where Go is available; exact-head GitHub CI must pass before owner handoff. The scratch environment's `/usr/bin/go` is not the Go compiler, so local test/build is unverified here. Owner Mac/Safari/mobile/light-dark/live acceptance is unverified. No merge, Issue closure, branch deletion, tag, release or deployment is authorized.
