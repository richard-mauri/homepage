package main

import (
 "os"
 "strings"
 "testing"
)

// Homepage deliberately navigates in the current tab. Native links remain
// functional without optional account access.
func TestHomepageSameTabLauncherContract(t *testing.T){
 code,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(code)
 for _,required:=range []string{
  "a.href=app.url",
  "if(!homeSessionToken || Date.now()>=homeSessionExpires){",
  "url.hash='suite_mode=local';",
  "const issued=await renewHandoff();",
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
  "alert('Sign in to Homepage before launching",
  "alert('Homepage session expired. Please sign in again.",
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
  "url.hash='suite_mode=local';",
 }{
  if !strings.Contains(handler,required) {t.Errorf("ScraperGo launch handler missing %q",required)}
 }
}
