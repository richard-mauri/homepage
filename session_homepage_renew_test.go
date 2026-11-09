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

func TestHomepageRenewSessionLifecycle(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks");if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("a",64)
 hash:=sha256.Sum256([]byte(approved))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 creds,_:=json.Marshal([]suitestore.Credential{{User:"reader",Role:"user",Digest:hex.EncodeToString(hash[:])}})
 if err:=os.WriteFile(authFile,creds,0600);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,11,9,10,0,0,0,time.UTC)
 svc:=newSessionService(store)
 if svc.persistenceError!=nil{t.Fatal(svc.persistenceError)}
 svc.now=func()time.Time{return now}
 call:=func(method,path,origin,auth string,handler http.HandlerFunc)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"https://home.themauris.org"+path,nil)
  if origin!=""{req.Header.Set("Origin",origin)}
  if auth!=""{req.Header.Set("Authorization",auth)}
  w:=httptest.NewRecorder();handler(w,req);return w
 }
 login:=call("POST","/api/session","https://home.themauris.org","Bearer "+approved,svc.handler)
 if login.Code!=200{t.Fatalf("login status %d: %s",login.Code,login.Body.String())}
 var parent struct{SessionToken,ExpiresAt string}
 if err:=json.Unmarshal(login.Body.Bytes(),&parent);err!=nil{t.Fatal(err)}
 handoff:=call("POST","/api/handoff/token","https://home.themauris.org","Session "+parent.SessionToken,svc.handoffHandler)
 if handoff.Code!=200{t.Fatalf("handoff %d",handoff.Code)}
 var child struct{Token,ExpiresAt string}
 if err:=json.Unmarshal(handoff.Body.Bytes(),&child);err!=nil{t.Fatal(err)}
 renew:=func(method,origin,auth string)*httptest.ResponseRecorder{return call(method,"/api/session/renew",origin,auth,svc.renewHomepageSessionHandler)}
 for _,tc:=range []struct{method,origin,auth string; want int}{
  {"GET","https://home.themauris.org","Session "+parent.SessionToken,405},
  {"POST","https://evil.example","Session "+parent.SessionToken,403},
  {"POST","https://home.themauris.org","Bearer "+child.Token,401},
  {"POST","https://home.themauris.org","",401},
 } {
  got:=renew(tc.method,tc.origin,tc.auth)
  if got.Code!=tc.want{t.Fatalf("%s %s expected %d, got %d",tc.method,tc.origin,tc.want,got.Code)}
 }
 now=now.Add(4*time.Hour)
 response:=renew("POST","https://home.themauris.org","Session "+parent.SessionToken)
 if response.Code!=200{t.Fatalf("renew %d: %s",response.Code,response.Body.String())}
 var updated struct{SessionToken,ExpiresAt,User,Role string}
 if err:=json.Unmarshal(response.Body.Bytes(),&updated);err!=nil{t.Fatal(err)}
 want:=now.Add(12*time.Hour).Format(time.RFC3339)
 if updated.SessionToken!=parent.SessionToken || updated.ExpiresAt!=want || updated.User!="reader" || updated.Role!="user"{
  t.Fatalf("unexpected renewal response: token stable=%v expiry=%s user=%q role=%q",updated.SessionToken==parent.SessionToken,updated.ExpiresAt,updated.User,updated.Role)
 }
 // Parent and existing live child survive a Homepage restart.
 svc=newSessionService(store);if svc.persistenceError!=nil{t.Fatal(svc.persistenceError)}
 svc.now=func()time.Time{return now}
 current:=call("GET","/api/session","","Session "+parent.SessionToken,svc.handler)
 if current.Code!=200 || !strings.Contains(current.Body.String(),want){t.Fatalf("parent not durable: %d",current.Code)}
 profile:=call("GET","/api/session/profile","https://newsgo.themauris.org","Bearer "+child.Token,svc.profileHandler)
 if profile.Code!=200{t.Fatalf("live child not extended: %d",profile.Code)}
 // Revoked approved credentials cannot renew otherwise valid sessions.
 if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
 if got:=renew("POST","https://home.themauris.org","Session "+parent.SessionToken).Code;got!=401{t.Fatalf("revoked credential renewed: %d",got)}
 if err:=os.WriteFile(authFile,creds,0600);err!=nil{t.Fatal(err)}
 now=now.Add(12*time.Hour)
 if got:=renew("POST","https://home.themauris.org","Session "+parent.SessionToken).Code;got!=401{t.Fatalf("expired session renewed: %d",got)}
}
