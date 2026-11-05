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

func TestDurableSessionsSurviveRestartAndPreserveRevocation(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks")
 if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("e",64)
 hash:=sha256.Sum256([]byte(approved))
 authPath:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authPath),0700);err!=nil{t.Fatal(err)}
 credentials,_:=json.Marshal([]suitestore.Credential{{User:"reader",Role:"user",Digest:hex.EncodeToString(hash[:])}})
 if err:=os.WriteFile(authPath,credentials,0600);err!=nil{t.Fatal(err)}
 start:=time.Date(2026,11,5,12,0,0,0,time.UTC)
 now:=start
 fresh:=func()*sessionService{
  s:=newSessionService(store)
  if s.persistenceError!=nil{t.Fatal(s.persistenceError)}
  s.now=func()time.Time{return now}
  return s
 }
 svc:=fresh()
 signIn:=func()string{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  req.Header.Set("Authorization","Bearer "+approved)
  out:=httptest.NewRecorder();svc.handler(out,req)
  if out.Code!=200{t.Fatalf("sign in: %d %s",out.Code,out.Body.String())}
  var v struct{SessionToken string `json:"sessionToken"`; ExpiresAt string `json:"expiresAt"`}
  if err:=json.Unmarshal(out.Body.Bytes(),&v);err!=nil{t.Fatal(err)}
  if v.ExpiresAt!=start.Add(12*time.Hour).Format(time.RFC3339){t.Fatalf("expiry changed: %s",v.ExpiresAt)}
  return v.SessionToken
 }
 parent:=signIn()
 second:=signIn()
 handoff:=func(parent string)string{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  req.Header.Set("Authorization","Session "+parent)
  out:=httptest.NewRecorder();svc.handoffHandler(out,req)
  if out.Code!=200{t.Fatalf("handoff: %d %s",out.Code,out.Body.String())}
  var v struct{Token string `json:"token"`}
  if err:=json.Unmarshal(out.Body.Bytes(),&v);err!=nil{t.Fatal(err)}
  return v.Token
 }
 bearer:=handoff(parent)
 otherBearer:=handoff(second)
 checkSession:=func(id string)int{
  req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
  req.Header.Set("Authorization","Session "+id)
  out:=httptest.NewRecorder();svc.handler(out,req)
  return out.Code
 }
 checkHandoff:=func(token string)int{
  req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
  req.Header.Set("Origin","https://stocks.themauris.org")
  req.Header.Set("Authorization","Bearer "+token)
  out:=httptest.NewRecorder();svc.profileHandler(out,req)
  return out.Code
 }
 state,err:=os.ReadFile(store.SessionStatePath())
 if err!=nil{t.Fatal(err)}
 info,err:=os.Stat(store.SessionStatePath())
 if err!=nil{t.Fatal(err)}
 if info.Mode().Perm()!=0600{t.Fatalf("session file permissions: %o",info.Mode().Perm())}
 if strings.Contains(string(state),parent)||strings.Contains(string(state),bearer)||strings.Contains(string(state),approved){
  t.Fatal("durable session state contains a raw session ID, app bearer, or credential")
 }
 now=now.Add(10*time.Minute)
 svc=fresh() // Simulates the Homepage process restarting.
 if checkSession(parent)!=200||checkSession(second)!=200||checkHandoff(bearer)!=200||checkHandoff(otherBearer)!=200{
  t.Fatal("session or child handoff did not survive restart")
 }
 logout:=httptest.NewRequest(http.MethodDelete,"https://home.themauris.org/api/session",nil)
 logout.Header.Set("Origin","https://home.themauris.org")
 logout.Header.Set("Authorization","Session "+parent)
 out:=httptest.NewRecorder();svc.handler(out,logout)
 if out.Code!=204{t.Fatalf("signout: %d",out.Code)}
 svc=fresh()
 if checkSession(parent)!=401||checkHandoff(bearer)!=401{
  t.Fatal("signed-out parent or child bearer survived restart")
 }
 if checkSession(second)!=200||checkHandoff(otherBearer)!=200{
  t.Fatal("signout incorrectly revoked another browser session")
 }
 now=start.Add(12*time.Hour)
 svc=fresh()
 if checkSession(second)!=401||checkHandoff(otherBearer)!=401{
  t.Fatal("expired session or app bearer accepted after restart")
 }
}

func TestCorruptDurableSessionFileFailsClosed(t *testing.T){
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 statePath:=store.SessionStatePath()
 if err:=os.MkdirAll(filepath.Dir(statePath),0700);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(statePath,[]byte("{broken"),0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 if svc.persistenceError==nil{t.Fatal("corrupt durable session state silently discarded")}
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 out:=httptest.NewRecorder();svc.handler(out,req)
 if out.Code!=503{t.Fatalf("corrupt state returned %d, want 503",out.Code)}
}

func TestDurableSessionWriteFailureCannotClaimLogin(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("a",64)
 digest:=sha256.Sum256([]byte(approved))
 dir:=filepath.Join(root,"auth")
 if err:=os.MkdirAll(dir,0700);err!=nil{t.Fatal(err)}
 creds,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(digest[:])}})
 if err:=os.WriteFile(filepath.Join(dir,"approved-tokens.json"),creds,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 // Make the durable-session target a non-empty directory; rename must fail
 // even if the test process has elevated filesystem permissions.
 if err:=os.Mkdir(store.SessionStatePath(),0700);err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 req.Header.Set("Origin","https://home.themauris.org")
 req.Header.Set("Authorization","Bearer "+approved)
 out:=httptest.NewRecorder();svc.handler(out,req)
 if out.Code!=503{t.Fatalf("failed durable write incorrectly reported %d",out.Code)}
 if len(svc.sessions)!=0{t.Fatal("unpersisted login remained in memory")}
}
