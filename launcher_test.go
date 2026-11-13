package main

import (
 "os"
 "strings"
 "testing"
)

// Homepage deliberately navigates in the current tab only after the server
// issues a verified session handoff.
func TestHomepageSameTabLauncherContract(t *testing.T){
 code,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(code)
 for _,required:=range []string{
  "a.href=app.url",
  "const issued=await renewHandoff();",
  "traceClient('launch_blocked','unauthenticated');",
  "notice.textContent='Sign in on Homepage or choose Continue as Guest (1 hour)",
  "location.assign(url.toString());",
  "url.hash='suite_token='+encodeURIComponent(suiteHandoffToken);",
 }{
  if !strings.Contains(script,required){t.Errorf("launcher missing %q",required)}
 }
 for _,forbidden:=range []string{
  "window.open(",
  "a.target='_blank'",
  "a.target=appTabName(",
  "appTabHandles",
  "url.hash='suite_mode=local';",
 }{
  if strings.Contains(script,forbidden){t.Errorf("same-tab optional launch contains %q",forbidden)}
 }
}

func TestScraperGoCardUsesAuthenticatedLaunchHandler(t *testing.T) {
 code,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(code)
 start:=strings.Index(script,"async function openSuiteApp(event)")
 end:=start+strings.Index(script[start:],"async function init()")
 if start<0 || end<=start {t.Fatal("missing suite launch handler")}
 handler:=script[start:end]
 for _,required:=range []string{
  "url.origin!=='https://scrapergo.themauris.org'",
  "event.preventDefault();",
  "const issued=await renewHandoff();",
  "url.hash='suite_token='+encodeURIComponent(suiteHandoffToken);",
  "traceClient('launch_blocked','unauthenticated');",
 }{
  if !strings.Contains(handler,required) {t.Errorf("ScraperGo launch handler missing %q",required)}
 }
 if strings.Contains(handler,"suite_mode=local") {t.Fatal("unauthenticated launch must stay on Homepage")}
 if !strings.Contains(handler,"  return;\n }\n const notice") {t.Fatal("failed handoff must return before navigation")}
}
