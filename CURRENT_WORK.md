# Homepage current work — 2026-11-04

GitHub `main` is the authoritative merged baseline. Active issue [#13](https://github.com/richard-mauri/homepage/issues/13), branch `fix/13-personal-signin-feedback`, draft PR pending against `main`. The issue was transferred from misattributed NewsGo issue #48; its draft PR #51 was closed without merge. No NewsGo functionality belongs in this Homepage fix.

Scope: Homepage's header Sign in opens the **Personal settings** dialog. Its inner Sign in submit now has an accent hover state, clear keyboard focus outline, pressed state, disabled busy styling, progress text, `aria-busy` and an existing polite live status. Repeated submits are blocked; the approved credential request has a 15-second abort timeout, with distinct rejected/expired-token, offline and timeout feedback. A successful login still uses the existing session, one-hour handoff, and refresh flow; credentials remain out of persistent browser storage. Browser-only app settings are unchanged.

Code: `static/app.js`, `static/styles.css`; regression: `signin_feedback_test.go`. No app version change, user-data migration, account authorization change, tunnel or launcher edit.

Next: confirm Go tests and build through CI at exact PR head, then owner runs Mac local `go test -count=1 ./...` and `./devbuild.sh` and checks Safari dark/light, mouse hover, keyboard focus/Enter, slow network, invalid token, successful login, return from app and logout. Mac runtime, browser and live tunnel acceptance are **not performed**. No merge, release, deployment or restart is authorized.
