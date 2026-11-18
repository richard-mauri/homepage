# Current work — Homepage

**Active:** [Issue #47](https://github.com/richard-mauri/homepage/issues/47), branch `ui/47-suite-buttons` from merged Homepage `main` `828d6c14f858fddccd4efb349cecb402dc5abab8`. This is the first Homepage-only button-module and Suite Management integration candidate. Verify the current draft PR number/head and checks on GitHub rather than trusting this snapshot.

**Implementation:** Added nested Go module `shared/ui` with embedded canonical button CSS derived from NewsGo's owner-approved style; Homepage serves it from `/suite-ui/buttons.css` on the Suite Management page. Replaced legacy teal button rules while leaving accordion summaries and app logic intact. Added a contract test, `SUITE_UI_DESIGN_SYSTEM.md`, and architecture/user documentation pointers. NewsGo's dependency integration is a subsequent coordinated consumer PR once this module has an immutable version; the other apps remain unchanged.

**Scope and safety:** No session API, accounts, activity data, storage, listeners, tokens, private files, counters, launcher destinations or tunnel changes. The scratch Homepage checkout began clean. Owner Mac-local uncommitted/unpushed work is inaccessible and must be checked before acceptance.

**Verification and next:** `git diff --check` passed. The scratch environment's `go` executable is not the Go compiler (`go test` reports `Go: Unknown option: test`), so local Go test/build are **unverified**. Check exact-head GitHub CI. Owner Mac/Safari/live acceptance is unverified. Draft only: no merge, central Issue #47 closure, branch deletion, tag, release, deployment, restart or suite-wide rollout is authorized.
