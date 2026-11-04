package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"

 "themauris.org/homepage/internal/suitestore"
)

// Sign out revokes every issued or renewed bearer from the same Homepage
// session. An independently authenticated second device stays signed in.
func TestSignOutRevokesAppHandoffsButNotOtherSessions(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks")
 if err!=nil {t.Fatal(err)}
 approved:=strings.Repeat("d",64)
 digest:=sha256.Sum256([]byte(approved))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 creds,_:=json.Marshal([]suitestore.Credential{{User:"reader",Role:"user",Digest:hex.EncodeToString(digest[:])}})
 if err:=os.WriteFile(authFile,creds,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)

 login:=func()string{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  req.Header.Set("Authorization","Bearer "+approved)
  w:=httptest.NewRecorder();svc.handler(w,req)
  if w.Code!=200{t.Fatalf("login: %d %s",w.Code,w.Body.String())}
  var result struct{SessionToken string `json:"sessionToken"`}
  if err:=json.Unmarshal(w.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
  return result.SessionToken
 }
 sessionA,sessionB:=login(),login()
 issue:=func(session string)string{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  req.Header.Set("Authorization","Session "+session)
  w:=httptest.NewRecorder();svc.handoffHandler(w,req)
  if w.Code!=200{t.Fatalf("handoff: %d %s",w.Code,w.Body.String())}
  var result struct{Token string `json:"token"`}
  if err:=json.Unmarshal(w.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
  return result.Token
 }
 // Two apps from the same Homepage login; a separate device stays independent.
 newsToken:=issue(sessionA)
 oldStocksToken:=issue(sessionA)
 otherDeviceToken:=issue(sessionB)
 profile:=func(token,origin string)int{
  req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
  req.Header.Set("Origin",origin)
  req.Header.Set("Authorization","Bearer "+token)
  w:=httptest.NewRecorder();svc.profileHandler(w,req);return w.Code
 }
 if profile(newsToken,"https://newsgo.themauris.org")!=200 ||
  profile(oldStocksToken,"https://stocks.themauris.org")!=200 ||
  profile(otherDeviceToken,"https://newsgo.themauris.org")!=200 {
  t.Fatal("a valid handoff was rejected before logout")
 }
 renewal:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/renew",nil)
 renewal.Header.Set("Origin","https://stocks.themauris.org")
 renewal.Header.Set("Authorization","Bearer "+oldStocksToken)
 out:=httptest.NewRecorder();svc.renewHandoffHandler(out,renewal)
 if out.Code!=200 {t.Fatalf("renewal: %d %s",out.Code,out.Body.String())}
 var rotated struct{Token string `json:"token"`}
 if err:=json.Unmarshal(out.Body.Bytes(),&rotated);err!=nil{t.Fatal(err)}
 if profile(rotated.Token,"https://stocks.themauris.org")!=200{t.Fatal("rotated handoff not usable")}
 logout:=httptest.NewRequest(http.MethodDelete,"https://home.themauris.org/api/session",nil)
 logout.Header.Set("Origin","https://home.themauris.org")
 logout.Header.Set("Authorization","Session "+sessionA)
 out=httptest.NewRecorder();svc.handler(out,logout)
 if out.Code!=204 {t.Fatalf("logout: %d %s",out.Code,out.Body.String())}
 for _,tc:=range []struct{token,origin string}{
  {newsToken,"https://newsgo.themauris.org"},
  {oldStocksToken,"https://stocks.themauris.org"},
  {rotated.Token,"https://stocks.themauris.org"},
 }{
  if got:=profile(tc.token,tc.origin);got!=401{t.Fatalf("signed-out app still has account access: %d",got)}
 }
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/storage/newsgo",nil)
 req.Header.Set("Origin","https://newsgo.themauris.org")
 req.Header.Set("Authorization","Bearer "+newsToken)
 out=httptest.NewRecorder();svc.suiteStorageHandler(store.Handler()).ServeHTTP(out,req)
 if out.Code!=401{t.Fatalf("signed-out bearer accessed personal storage: %d",out.Code)}
 out=httptest.NewRecorder();svc.renewHandoffHandler(out,renewal)
 if out.Code!=401{t.Fatalf("signed-out app bearer renewed: %d",out.Code)}
 if got:=profile(otherDeviceToken,"https://newsgo.themauris.org");got!=200 {
  t.Fatalf("another device was signed out: %d",got)
 }
 if got:=profile(issue(sessionB),"https://stocks.themauris.org");got!=200 {
  t.Fatalf("independent session cannot issue a handoff: %d",got)
 }
 if got:=profile(issueAfterSignOut(t,svc,sessionA),"https://newsgo.themauris.org");got!=401 {
  t.Fatalf("signed-out login issued a new bearer: %d",got)
 }
}

func issueAfterSignOut(t *testing.T,svc *sessionService,session string)string{
 t.Helper()
 req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
 req.Header.Set("Origin","https://home.themauris.org")
 req.Header.Set("Authorization","Session "+session)
 w:=httptest.NewRecorder();svc.handoffHandler(w,req)
 if w.Code!=401 {t.Fatalf("signed-out session could issue a bearer: %d",w.Code)}
 return ""
}

func TestHomepageSignInHandoffUsesRevocableSession(t *testing.T){
 content,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 js:=string(content)
 if !strings.Contains(js,"await renewHandoff();") {t.Fatal("Homepage sign in must mint app handoffs with its session")}
 if strings.Contains(js,"await renewHandoff(approvedToken)") {t.Fatal("direct token handoff escapes logout's session ownership")}
}
