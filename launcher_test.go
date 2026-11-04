package main

import (
 "os"
 "strings"
 "testing"
)

// These are static launcher-contract checks. Safari behavior with named windows
// and popup settings still requires owner-side desktop/mobile acceptance.
func TestHomepageNamedTabLauncherContract(t *testing.T){
 code,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(code)
 for _,required:=range []string{
  "return 'mauri_app_'+url.hostname",
  "a.target=appTabName(app.url)",
  "const appTabHandles=new Map();",
  "if(!tab.closed){tab.focus();return tab}",
  "if(existingAppTab(tabName)){event.preventDefault();return;}",
  "appTabHandles.set(tabName,opened);",
  "const opened=window.open('',tabName);",
  "event.preventDefault();",
  "opened.location.href=url.toString();",
  "opened.opener=null",
  "event.metaKey || event.ctrlKey",
  "if(link.dataset.opening==='true'){event.preventDefault();return;}",
 } {
  if !strings.Contains(script,required){t.Errorf("launcher missing %q",required)}
 }
 if strings.Contains(script,"location.assign(url.toString())"){
  t.Error("suite launcher still replaces Homepage tab")
 }
 if strings.Contains(script,"a.target='_blank'"){
  t.Error("app cards still create unlimited new tabs")
 }
}

func TestHomepageOptionalAccountLaunchContract(t *testing.T){
 code,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(code)
 for _,required:=range []string{
  "if(isSuite)url.hash='';",
  "if(isSuite && homeSessionToken && Date.now()<homeSessionExpires){",
  "const issued=await renewHandoff();",
  "url.hash='suite_token='+encodeURIComponent(suiteHandoffToken);",
 } {
  if !strings.Contains(script,required){t.Errorf("optional account handoff missing %q",required)}
 }
 for _,forbidden:=range []string{
  "alert('Sign in to Homepage before launching",
  "alert('Homepage session expired. Please sign in again.",
 }{
  if strings.Contains(script,forbidden){t.Errorf("app launch blocked by %q",forbidden)}
 }
}
