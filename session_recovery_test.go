package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "net/url"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"

 "themauris.org/homepage/internal/suitestore"
)

func TestNewsGoRecoveryFromPersistentHomepageSession(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks")
 if err!=nil {t.Fatal(err)}
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

 recover:=func(method string, withCookie bool)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"https://home.themauris.org/launch/newsgo?next=https://evil.example",nil)
  if withCookie{req.AddCookie(cookie)}
  recorder:=httptest.NewRecorder()
  svc.newsGoRecoveryHandler(recorder,req)
  return recorder
 }
 if got:=recover(http.MethodGet,false);got.Code!=302 || got.Header().Get("Location")!="https://newsgo.themauris.org/#suite_mode=local"{
  t.Fatalf("signed-out recovery must be browser-only: %d %s",got.Code,got.Header().Get("Location"))
 }
 if got:=recover(http.MethodPost,true);got.Code!=405{t.Fatalf("mutation allowed: %d",got.Code)}
 result:=recover(http.MethodGet,true)
 if result.Code!=302{t.Fatalf("recover: %d %s",result.Code,result.Body.String())}
 if result.Header().Get("Cache-Control")!="no-store" || result.Header().Get("Referrer-Policy")!="no-referrer"{
  t.Fatal("recovery redirect must not be cached or leak referrer")
 }
 dest,err:=url.Parse(result.Header().Get("Location"))
 if err!=nil{t.Fatal(err)}
 if dest.Scheme!="https" || dest.Host!="newsgo.themauris.org" || dest.RawQuery!=""{
  t.Fatalf("unsafe recovery destination: %q",result.Header().Get("Location"))
 }
 fragment,err:=url.ParseQuery(dest.Fragment)
 if err!=nil{t.Fatal(err)}
 bearer:=fragment.Get("suite_token")
 if len(bearer)!=64 || strings.Contains(result.Header().Get("Location"),parent.SessionToken) {
  t.Fatal("missing child handoff or leaked parent token")
 }
 // A fresh child bearer survives a Homepage process restart, and the parent
 // expiration is unchanged by handoff recovery.
 restarted:=newSessionService(store)
 if restarted.persistenceError!=nil{t.Fatal(restarted.persistenceError)}
 restarted.now=func()time.Time{return now}
 profileReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
 profileReq.Header.Set("Origin","https://newsgo.themauris.org")
 profileReq.Header.Set("Authorization","Bearer "+bearer)
 profile:=httptest.NewRecorder()
 restarted.profileHandler(profile,profileReq)
 if profile.Code!=200{t.Fatalf("handoff lost on restart: %d",profile.Code)}
 parentReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 parentReq.Header.Set("Authorization","Session "+parent.SessionToken)
 current:=httptest.NewRecorder()
 restarted.handler(current,parentReq)
 if current.Code!=200 || !strings.Contains(current.Body.String(),parent.ExpiresAt){
  t.Fatal("recovery silently renewed or lost parent")
 }

 // Credential revocation blocks *new* handoffs and invalidates old ones.
 if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
 if got:=recover(http.MethodGet,true);got.Header().Get("Location")!="https://newsgo.themauris.org/#suite_mode=local"{
  t.Fatal("revoked credential recovered a login")
 }
 if err:=os.WriteFile(authFile,credentials,0600);err!=nil{t.Fatal(err)}
 now=now.Add(12*time.Hour+time.Second)
 if got:=recover(http.MethodGet,true);got.Header().Get("Location")!="https://newsgo.themauris.org/#suite_mode=local"{
  t.Fatal("expired Homepage session recovered or extended")
 }
}

func TestNewsGoRecoveryStopsOnDurableStorageFailure(t *testing.T) {
 store,err:=suitestore.New(t.TempDir(),"newsgo")
 if err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 svc.persistenceError=os.ErrPermission
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/launch/newsgo",nil)
 w:=httptest.NewRecorder()
 svc.newsGoRecoveryHandler(w,req)
 if w.Code!=http.StatusServiceUnavailable{t.Fatalf("storage error hidden: %d",w.Code)}
}
