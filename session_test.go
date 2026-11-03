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

func TestSessionLifecycle(t *testing.T){
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo","mauristocks");if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("a",64)
 hash:=sha256.Sum256([]byte(token))
 if err:=os.MkdirAll(filepath.Join(root,"auth"),0700);err!=nil{t.Fatal(err)}
 credentials,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(hash[:]),Role:"user"}})
 if err:=os.WriteFile(filepath.Join(root,"auth","approved-tokens.json"),credentials,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+token)
 w:=httptest.NewRecorder();svc.handler(w,login)
 if w.Code!=200{t.Fatalf("login: %d %s",w.Code,w.Body.String())}
 cookies:=w.Result().Cookies()
 if len(cookies)!=1||!cookies[0].Secure||!cookies[0].HttpOnly||cookies[0].Domain!="" {t.Fatalf("unsafe cookie: %+v",cookies)}
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil);req.AddCookie(cookies[0])
 w=httptest.NewRecorder();svc.handler(w,req)
 if w.Code!=200||!strings.Contains(w.Body.String(),`"role":"user"`){t.Fatalf("session: %d %s",w.Code,w.Body.String())}
 logout:=httptest.NewRequest(http.MethodDelete,"https://home.themauris.org/api/session",nil)
 logout.Header.Set("Origin","https://home.themauris.org");logout.AddCookie(cookies[0])
 w=httptest.NewRecorder();svc.handler(w,logout)
 if w.Code!=204{t.Fatalf("logout: %d",w.Code)}
 w=httptest.NewRecorder();svc.handler(w,req)
 if w.Code!=401{t.Fatalf("session survived logout: %d",w.Code)}
}
func TestSessionRejectsCrossOrigin(t *testing.T){
 svc:=newSessionService(nil)
 req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 req.Header.Set("Origin","https://stocks.themauris.org")
 w:=httptest.NewRecorder();svc.handler(w,req)
 if w.Code!=403{t.Fatalf("expected 403, got %d",w.Code)}
}

func TestStorageSessionRevocationAndCSRF(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"mauristocks");if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("d",64)
 h:=sha256.Sum256([]byte(token))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 records,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(h[:])}})
 if err:=os.WriteFile(authFile,records,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+token)
 w:=httptest.NewRecorder();svc.handler(w,login)
 if w.Code!=200{t.Fatalf("login %d: %s",w.Code,w.Body.String())}
 cookie:=w.Result().Cookies()[0]
 handler:=svc.storageHandler(store.Handler())
 write:=func(origin string) int {
  req:=httptest.NewRequest(http.MethodPut,"https://home.themauris.org/api/storage/mauristocks",strings.NewReader(`{"revision":0,"data":{"symbols":["CAT"]}}`))
  req.AddCookie(cookie)
  if origin!=""{req.Header.Set("Origin",origin)}
  resp:=httptest.NewRecorder();handler.ServeHTTP(resp,req)
  return resp.Code
 }
 if got:=write("https://evil.example");got!=401{t.Fatalf("cross-origin cookie accepted: %d",got)}
 if got:=write("");got!=403{t.Fatalf("missing origin accepted: %d",got)}
 if got:=write("https://home.themauris.org");got!=200{t.Fatalf("same-origin write failed: %d",got)}
 if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/storage/mauristocks",nil)
 req.AddCookie(cookie)
 resp:=httptest.NewRecorder();handler.ServeHTTP(resp,req)
 if resp.Code!=401{t.Fatalf("revoked session accepted: %d",resp.Code)}
 sessionReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 sessionReq.AddCookie(cookie)
 resp=httptest.NewRecorder();svc.handler(resp,sessionReq)
 if resp.Code!=401{t.Fatalf("revoked session status accepted: %d",resp.Code)}
}

func TestSuiteProfileReadOnlyAndOriginPolicy(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"mauristocks")
 if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("b",64)
 digest:=sha256.Sum256([]byte(token))
 if err:=os.MkdirAll(filepath.Join(root,"auth"),0700);err!=nil{t.Fatal(err)}
 credentials,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(digest[:]),Role:"user"}})
 if err:=os.WriteFile(filepath.Join(root,"auth","approved-tokens.json"),credentials,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+token)
 w:=httptest.NewRecorder();svc.handler(w,login)
 if w.Code!=200{t.Fatalf("login failed: %d",w.Code)}
 cookie:=w.Result().Cookies()[0]
 fetch:=func(origin string,withCookie bool,method string)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"https://home.themauris.org/api/session/profile",nil)
  if origin!=""{req.Header.Set("Origin",origin)}
  if withCookie{req.AddCookie(cookie)}
  result:=httptest.NewRecorder();svc.profileHandler(result,req)
  return result
 }
 w=fetch("https://stocks.themauris.org",true,http.MethodGet)
 if w.Code!=401 {t.Fatalf("cross-origin cookie accepted: %d",w.Code)}
 w=fetch("https://home.themauris.org",true,http.MethodGet)
 if w.Code!=200 || !strings.Contains(w.Body.String(),`"user":"reader"`) || !strings.Contains(w.Body.String(),`"role":"user"`) {
  t.Fatalf("Homepage session profile: %d %s",w.Code,w.Body.String())
 }
 if w.Header().Get("Access-Control-Allow-Origin")!="https://home.themauris.org" || w.Header().Get("Access-Control-Allow-Credentials")!="true" {
  t.Fatal("missing trusted origin credentialed CORS response")
 }
 if w.Header().Get("Cache-Control")!="no-store"{t.Fatal("profile must not be cached")}
 if w=fetch("https://evil.example",true,http.MethodGet);w.Code!=403 || w.Header().Get("Access-Control-Allow-Origin")!="" {t.Fatal("untrusted origin not rejected")}
 if w=fetch("https://stocks.themauris.org",false,http.MethodGet);w.Code!=401 {t.Fatal("anonymous profile access")}
 if w=fetch("https://stocks.themauris.org",true,http.MethodPost);w.Code!=405 {t.Fatal("profile must be read only")}
}

func TestSuiteStorageRejectsCrossOriginCookieFallback(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo");if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("c",64)
 hash:=sha256.Sum256([]byte(approved))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 b,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(hash[:]),Role:"user"}})
 if err:=os.WriteFile(authFile,b,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+approved)
 w:=httptest.NewRecorder();svc.handler(w,login)
 if w.Code!=200{t.Fatalf("login: %d",w.Code)}
 cookie:=w.Result().Cookies()[0]
 handler:=svc.suiteStorageHandler(store.Handler())
 for _,method:=range []string{http.MethodGet,http.MethodPut}{
  req:=httptest.NewRequest(method,"https://home.themauris.org/api/storage/newsgo",nil)
  req.Header.Set("Origin","https://newsgo.themauris.org")
  req.AddCookie(cookie)
  out:=httptest.NewRecorder();handler.ServeHTTP(out,req)
  if out.Code!=401{t.Fatalf("cross-app cookie accepted for %s: %d",method,out.Code)}
 }
}

func TestHandoffBearerExpiryAndRevocation(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("f",64)
 digest:=sha256.Sum256([]byte(token))
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 entries,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(digest[:]),Role:"user"}})
 if err:=os.WriteFile(authFile,entries,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 start:=time.Now()
 svc.now=func()time.Time{return start}
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+token)
 response:=httptest.NewRecorder();svc.handler(response,login)
 if response.Code!=200{t.Fatalf("login: %d",response.Code)}
 cookie:=response.Result().Cookies()[0]
 issue:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
 issue.Header.Set("Origin","https://home.themauris.org");issue.AddCookie(cookie)
 response=httptest.NewRecorder();svc.handoffHandler(response,issue)
 if response.Code!=200{t.Fatalf("issue: %d %s",response.Code,response.Body.String())}
 var handed struct{Token string `json:"token"`}
 if err:=json.Unmarshal(response.Body.Bytes(),&handed);err!=nil{t.Fatal(err)}
 if len(handed.Token)!=64||handed.Token==token{t.Fatal("missing or durable handoff token")}
 profile:=func()int{
  req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
  req.Header.Set("Origin","https://newsgo.themauris.org")
  req.Header.Set("Authorization","Bearer "+handed.Token)
  out:=httptest.NewRecorder();svc.profileHandler(out,req)
  return out.Code
 }
 if got:=profile();got!=200{t.Fatalf("profile via handoff: %d",got)}
 handler:=svc.suiteStorageHandler(store.Handler())
 req:=httptest.NewRequest(http.MethodPut,"https://home.themauris.org/api/storage/newsgo",strings.NewReader(`{"revision":0,"data":{"subscriptions":[]}}`))
 req.Header.Set("Origin","https://newsgo.themauris.org")
 req.Header.Set("Authorization","Bearer "+handed.Token)
 out:=httptest.NewRecorder();handler.ServeHTTP(out,req)
 if out.Code!=200{t.Fatalf("storage via handoff: %d %s",out.Code,out.Body.String())}
 svc.now=func()time.Time{return start.Add(61*time.Minute)}
 if got:=profile();got!=401{t.Fatalf("expired handoff: %d",got)}
 svc.now=func()time.Time{return start}
 if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
 if got:=profile();got!=401{t.Fatalf("revoked handoff: %d",got)}
}

func TestHandoffExplicitIdentityOverridesStaleCookie(t *testing.T){
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 oldToken:=strings.Repeat("a",64)
 newToken:=strings.Repeat("b",64)
 digest:=func(token string)string{h:=sha256.Sum256([]byte(token));return hex.EncodeToString(h[:])}
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 entries,_:=json.Marshal([]suitestore.Credential{{User:"olduser",Digest:digest(oldToken)},{User:"newuser",Digest:digest(newToken)}})
 if err:=os.WriteFile(authFile,entries,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org");login.Header.Set("Authorization","Bearer "+oldToken)
 out:=httptest.NewRecorder();svc.handler(out,login)
 if out.Code!=200{t.Fatalf("login: %d",out.Code)}
 issue:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
 issue.Header.Set("Origin","https://home.themauris.org")
 issue.Header.Set("Authorization","Bearer "+newToken)
 issue.AddCookie(out.Result().Cookies()[0])
 out=httptest.NewRecorder();svc.handoffHandler(out,issue)
 if out.Code!=200{t.Fatalf("issue: %d",out.Code)}
 var result struct{Token string `json:"token"`}
 if err:=json.Unmarshal(out.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
 req.Header.Set("Authorization","Bearer "+result.Token)
 out=httptest.NewRecorder();svc.profileHandler(out,req)
 if out.Code!=200 || !strings.Contains(out.Body.String(),`"user":"newuser"`){
  t.Fatalf("handoff used stale cookie identity: %d %s",out.Code,out.Body.String())
 }
}

func TestProfileBearerCORSPreflight(t *testing.T) {
 svc:=newSessionService(nil)
 testCase:=func(origin, method, headers string, expected int) {
  t.Helper()
  req:=httptest.NewRequest(http.MethodOptions,"https://home.themauris.org/api/session/profile",nil)
  req.Header.Set("Origin",origin)
  req.Header.Set("Access-Control-Request-Method",method)
  if headers!=""{req.Header.Set("Access-Control-Request-Headers",headers)}
  w:=httptest.NewRecorder();svc.profileHandler(w,req)
  if w.Code!=expected{t.Fatalf("origin=%q method=%q headers=%q: got %d, want %d",origin,method,headers,w.Code,expected)}
  if expected==http.StatusNoContent {
   if w.Header().Get("Access-Control-Allow-Origin")!=origin ||
      w.Header().Get("Access-Control-Allow-Headers")!="Authorization" ||
      w.Header().Get("Access-Control-Allow-Methods")!="GET" ||
      w.Header().Get("Access-Control-Allow-Credentials")!="true"{
     t.Fatal("trusted bearer preflight missing required CORS response headers")
   }
  }else if w.Header().Get("Access-Control-Allow-Origin")!=""{
   t.Fatal("untrusted origin allowed")
  }
 }
 testCase("https://newsgo.themauris.org","GET","authorization",204)
 testCase("https://stocks.themauris.org","GET","Authorization",204)
 testCase("https://evil.example","GET","authorization",403)
 testCase("https://newsgo.themauris.org","PUT","authorization",403)
 testCase("https://newsgo.themauris.org","GET","x-secret",403)
}

func TestHandoffRejectsForeignOrigin(t *testing.T) {
 svc:=newSessionService(nil)
 req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
 req.Header.Set("Origin","https://evil.example")
 out:=httptest.NewRecorder();svc.handoffHandler(out,req)
 if out.Code!=403{t.Fatalf("foreign origin accepted: %d",out.Code)}
}
