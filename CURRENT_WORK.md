# Current work — Homepage

**Active:** [Issue #56](https://github.com/richard-mauri/homepage/issues/56), branch `ui/56-homepage-banner`, dependent on still-open foundation Issue #54 / draft PR #55. The proposal Issue #52 / draft PR #53 also remains open. The branch starts from PR #55 head `dc72f4af7698664dd8f146304564aa013302432d`, itself derived from merged main `3f2fe48f7a7a889f5e6af5c764687ff9cf8e45b3`. Verify the actual new draft PR and checks on GitHub.

**Implementation candidate:** Homepage consumes the refined shared borderless banner for its launchpad topbar. Operational controls remain in a secondary toolbar; the embedded Tanner JPEG is locally served without a second source asset. Existing JS updates only the verified status label in the banner while preserving Manage Suite, guest, clock and theme behavior. Hero/cards and server session APIs are unchanged. ScraperGo PR #144 remains open awaiting dependency re-pin.

**Safety:** No authentication/session logic, accounts, activity, storage, subscriptions, watchlists, private credentials, launcher targets, live tunnels, build counters, deployment or release changed. The scratch checkout began clean; owner Mac-local uncommitted/unpushed work remains inaccessible. No service was restarted.

**Verification:** Check static diffs, Homepage root and nested-module tests and exact-head GitHub CI before owner handoff. The scratch environment's `/usr/bin/go` is not the Go compiler, so local test/build remains unverified here. Owner Mac/Safari/mobile/light-dark/live acceptance remains unverified. No merge, Issue closure, branch deletion, tag, release or deployment is authorized.
