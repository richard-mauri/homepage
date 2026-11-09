package main

import (
 "strings"
 "testing"
)

func TestRenewControlsOnlyInManageTokensCurrentSession(t *testing.T) {
 admin,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil{t.Fatal(err)}
 page:=string(admin)
 start:=strings.Index(page,`<section id="account"`)
 end:=start+strings.Index(page[start:],`</section>`)
 if start<0||end<start{t.Fatal("Current session panel missing")}
 panel:=page[start:end]
 for _,element:=range []string{`id="renew"`,`id="remaining"`,`id="logout"`,`class="session-actions"`} {
  if !strings.Contains(panel,element){t.Errorf("Current session panel missing %s",element)}
 }
 if !strings.Contains(page,`'/api/session/renew'`){t.Error("Manage Tokens does not call Homepage renewal endpoint")}
 home,err:=assets.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 if strings.Contains(string(home),`renew.addEventListener`)||strings.Contains(string(home),`signedIn.append(identity,manage,remaining`){
  t.Error("renewal control was not removed from Homepage account header")
 }
 if strings.Contains(string(home),`signOut.addEventListener`)||strings.Contains(string(home),`signOut.textContent='Sign out'`){
  t.Error("duplicate Sign out control remains in Homepage header")
 }
 if !strings.Contains(string(home),`manage.hidden=false; // Every signed-in user`){
  t.Error("Manage tokens must be available to non-admin signed-in users")
 }
 if !strings.Contains(page,`localStorage.removeItem('mauri-home-session-v1')`) ||
    !strings.Contains(page,`sessionStorage.removeItem('suite-handoff-v1')`){
  t.Error("Disconnect must clear the Homepage browser session and cached handoff")
 }
 if !strings.Contains(page,`Authorization:'Session '+sessionToken`){
  t.Error("Disconnect must revoke the correct Homepage session by explicit session ID")
 }
}
