package main

import (
 "os"
 "strings"
 "testing"
)

// Browser-side regression contract for the Homepage/MauriStocks return path.
// Runtime acceptance still requires Safari and the live tunnel.
func TestSessionRefreshDistinguishesOutageFromSignout(t *testing.T) {
 source,err:=os.ReadFile("static/app.js")
 if err!=nil {t.Fatal(err)}
 js:=string(source)
 for _,required:=range []string{
  "response.status===401||response.status===403",
  "return {kind:'unavailable'}",
  "if(state.kind==='invalid'){showSignedOut();return false;}",
  "showUnavailable();",
  "retryAuth.textContent='Connection issue — Retry'",
  "retryAuth.addEventListener('click'",
  "window.addEventListener('pageshow'",
  "window.addEventListener('online'",
  "window.addEventListener('focus'",
  "if(homeSessionToken)return await request({},'cookie_fallback');",
 }{
  if !strings.Contains(js,required){t.Errorf("session resilience behavior missing %q",required)}
 }
 if strings.Contains(js,"Date.now()>suiteHandoffExpires-5*60*1000"){
  t.Error("Homepage refresh must not be gated on app handoff expiration")
 }
}
