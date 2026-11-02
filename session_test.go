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
 if len(cookies)!=1||!cookies[0].Secure||!cookies[0].HttpOnly||cookies[0].Domain!="themauris.org" {t.Fatalf("unsafe cookie: %+v",cookies)}
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
 if got:=write("https://evil.example");got!=403{t.Fatalf("foreign origin accepted: %d",got)}
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
 if w.Code!=200 || !strings.Contains(w.Body.String(),`"user":"reader"`) || !strings.Contains(w.Body.String(),`"role":"user"`) {
  t.Fatalf("profile: %d %s",w.Code,w.Body.String())
 }
 if w.Header().Get("Access-Control-Allow-Origin")!="https://stocks.themauris.org" || w.Header().Get("Access-Control-Allow-Credentials")!="true" {
  t.Fatal("missing trusted origin credentialed CORS response")
 }
 if w.Header().Get("Cache-Control")!="no-store"{t.Fatal("profile must not be cached")}
 if w=fetch("https://evil.example",true,http.MethodGet);w.Code!=403 || w.Header().Get("Access-Control-Allow-Origin")!="" {t.Fatal("untrusted origin not rejected")}
 if w=fetch("https://stocks.themauris.org",false,http.MethodGet);w.Code!=401 {t.Fatal("anonymous profile access")}
 if w=fetch("https://stocks.themauris.org",true,http.MethodPost);w.Code!=405 {t.Fatal("profile must be read only")}
}
