package main

import (
 "os"
 "strings"
 "testing"
)

func TestPersonalSignInVisualFeedback(t *testing.T) {
 css,err:=os.ReadFile("static/styles.css")
 if err!=nil {t.Fatal(err)}
 for _,want:=range []string{
  ".suite-signin-submit:hover:not(:disabled)",
  ".suite-signin-submit:focus-visible",
  ".suite-signin-submit:active:not(:disabled)",
  ".suite-signin-submit:disabled",
  "cursor:progress",
 } {
  if !strings.Contains(string(css),want) {t.Errorf("missing sign-in visual state %q",want)}
 }
}

func TestPersonalSignInProgressAndErrors(t *testing.T) {
 js,err:=os.ReadFile("static/app.js")
 if err!=nil {t.Fatal(err)}
 for _,want:=range []string{
  "submit.className='suite-signin-submit'",
  "if(signInPending)return;",
  "submit.textContent='Signing in…'",
  "submit.setAttribute('aria-busy','true')",
  "message.textContent='Checking your personal token…'",
  "message.textContent='Token accepted. Preparing application access…'",
  "response.status===401||response.status===403",
  "signIn timed out",
  "Network error.",
  "clearTimeout(timeout)",
  "submit.removeAttribute('aria-busy')",
  "if(authenticated)hideDialog()",
  "signal:controller.signal",
 } {
  if !strings.Contains(string(js),want) {t.Errorf("missing sign-in behavior %q",want)}
 }
 if !strings.Contains(string(js),"message.setAttribute('aria-live','polite')") {t.Error("progress messages must be announced")}
}
