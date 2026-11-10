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
 "time"

 "themauris.org/homepage/internal/suitestore"
)

func TestSuiteRecoveryBridgeDoesNotIssueTokensOnGet(t *testing.T) {
 store,err:=suitestore.New(t.TempDir(),"newsgo")
 if err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 request:=func(method string)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"https://home.themauris.org/launch/newsgo?next=https://evil.example",nil)
  rec:=httptest.NewRecorder()
  svc.suiteRecoveryHandler(rec,req)
  return rec
 }
 if got:=request(http.MethodPost);got.Code!=http.StatusMethodNotAllowed{t.Fatalf("POST: %d",got.Code)}
 result:=request(http.MethodGet)
 if result.Code!=http.StatusOK{t.Fatalf("GET: %d",result.Code)}
 if result.Header().Get("Cache-Control")!="no-store" ||
    result.Header().Get("Referrer-Policy")!="no-referrer" ||
    !strings.Contains(result.Header().Get("Content-Security-Policy"),"script-src 'self'"){
  t.Fatal("recovery bridge lacks security headers")
 }
 if !strings.Contains(result.Body.String(),"/static/launch-suite.js") ||
    strings.Contains(result.Body.String(),"evil.example"){
  t.Fatal("recovery bridge must load trusted script without open redirect")
 }
 script,err:=assets.ReadFile("static/launch-suite.js")
 if err!=nil{t.Fatal(err)}
 text:=string(script)
 for _,must:=range []string{
  "'/launch/newsgo':'https://newsgo.themauris.org/'",
  "'/launch/stocks':'https://stocks.themauris.org/'",
  "'/launch/sailing':'https://weather.themauris.org/'",
  "const suiteDestination=destinations[location.pathname]",
  "localStorage.getItem(homeLoginKey)",
  "headers=sessionID?{Authorization:'Session '+sessionID}:{}",
  "method:'POST',credentials:'same-origin'",
  "if(storedSession && response.status===401)response=await issueHandoff('')",
  "location.replace(suiteDestination+'#suite_mode=local')",
  "location.replace(suiteDestination+'#suite_token='+data.token)",
 } {
  if !strings.Contains(text,must){t.Errorf("missing recovery invariant: %q",must)}
 }
 if strings.Contains(text,"?suite_token=") || strings.Contains(text,"/api/session/renew"){
  t.Fatal("handoff must not use query strings or extend Homepage login")
 }
}

func TestNewsGoRecoveryUsesExistingDurableHomepageLogin(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks")
 if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("a",64)
 digest:=sha256.Sum256([]byte(approved))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 credentials,_:=json.Marshal([]suitestore.Credential{{User:"reader",Role:"user",Digest:hex.EncodeToString(digest[:])}})
 if err:=os.WriteFile(authFile,credentials,0600);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,11,9,12,0,0,0,time.UTC)
 svc:=newSessionService(store)
 if svc.persistenceError!=nil{t.Fatal(svc.persistenceError)}
 svc.now=func()time.Time{return now}
 loginReq:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 loginReq.Header.Set("Origin","https://home.themauris.org")
 loginReq.Header.Set("Authorization","Bearer "+approved)
 login:=httptest.NewRecorder()
 svc.handler(login,loginReq)
 if login.Code!=200{t.Fatalf("login: %d",login.Code)}
 var parent struct{SessionToken,ExpiresAt string}
 if err:=json.Unmarshal(login.Body.Bytes(),&parent);err!=nil{t.Fatal(err)}
 cookie:=login.Result().Cookies()[0]

 // Equivalent to recovery page POST. The same-origin request works with
 // Homepage localStorage's session ID, even if its cookie has gone missing.
 handoff:=func(withSession,withCookie bool)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  if withSession{req.Header.Set("Authorization","Session "+parent.SessionToken)}
  if withCookie{req.AddCookie(cookie)}
  rec:=httptest.NewRecorder()
  svc.handoffHandler(rec,req)
  return rec
 }
 noCred:=handoff(false,false)
 if noCred.Code!=401{t.Fatalf("missing parent credentials minted handoff: %d",noCred.Code)}
 result:=handoff(true,false)
 if result.Code!=200{t.Fatalf("parent session handoff: %d %s",result.Code,result.Body.String())}
 var child struct{Token,ExpiresAt string}
 if err:=json.Unmarshal(result.Body.Bytes(),&child);err!=nil{t.Fatal(err)}
 if len(child.Token)!=64 || child.ExpiresAt!=parent.ExpiresAt || child.Token==parent.SessionToken{
  t.Fatal("recovery must issue a distinct bearer capped to original parent expiry")
 }
 cookieFallback:=handoff(false,true)
 if cookieFallback.Code!=200{t.Fatalf("host-only cookie recovery: %d",cookieFallback.Code)}
 // Persistent session and newly minted handoff both survive server restarts.
 svc=newSessionService(store)
 if svc.persistenceError!=nil{t.Fatal(svc.persistenceError)}
 svc.now=func()time.Time{return now}
 profileReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
 profileReq.Header.Set("Origin","https://newsgo.themauris.org")
 profileReq.Header.Set("Authorization","Bearer "+child.Token)
 profile:=httptest.NewRecorder();svc.profileHandler(profile,profileReq)
 if profile.Code!=200{t.Fatalf("handoff lost on restart: %d",profile.Code)}
 sessionReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 sessionReq.Header.Set("Authorization","Session "+parent.SessionToken)
 current:=httptest.NewRecorder();svc.handler(current,sessionReq)
 if current.Code!=200 || !strings.Contains(current.Body.String(),parent.ExpiresAt){
  t.Fatal("recovery extended or lost parent session")
 }
 // Revoked/expired sessions must never mint child access.
 if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
 if got:=handoff(true,false).Code;got!=401{t.Fatalf("revoked parent recovered: %d",got)}
 if err:=os.WriteFile(authFile,credentials,0600);err!=nil{t.Fatal(err)}
 now=now.Add(12*time.Hour+time.Second)
 if got:=handoff(true,false).Code;got!=401{t.Fatalf("expired parent recovered: %d",got)}
}

func TestNewsGoRecoveryStopsOnDurableStorageFailure(t *testing.T) {
 store,err:=suitestore.New(t.TempDir(),"newsgo")
 if err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 svc.persistenceError=os.ErrPermission
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/launch/newsgo",nil)
 w:=httptest.NewRecorder()
 svc.suiteRecoveryHandler(w,req)
 if w.Code!=http.StatusServiceUnavailable{t.Fatalf("storage error hidden: %d",w.Code)}
}
