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

func TestSuiteSubdomainStorageCORS(t *testing.T){
 root:=t.TempDir()
 store,err:=suitestore.New(root,"mauristocks");if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("f",64)
 sum:=sha256.Sum256([]byte(token))
 if err=os.MkdirAll(filepath.Join(root,"auth"),0700);err!=nil{t.Fatal(err)}
 b,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(sum[:])}})
 if err=os.WriteFile(filepath.Join(root,"auth","approved-tokens.json"),b,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 login:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
 login.Header.Set("Origin","https://home.themauris.org")
 login.Header.Set("Authorization","Bearer "+token)
 w:=httptest.NewRecorder();svc.handler(w,login)
 if w.Code!=200{t.Fatalf("login: %d",w.Code)}
 cookie:=w.Result().Cookies()[0]
 handler:=svc.suiteStorageHandler(store.Handler())
 preflight:=httptest.NewRequest(http.MethodOptions,"https://home.themauris.org/api/storage/mauristocks",nil)
 preflight.Header.Set("Origin","https://stocks.themauris.org")
 preflight.Header.Set("Access-Control-Request-Method","PUT")
 preflight.Header.Set("Access-Control-Request-Headers","content-type")
 w=httptest.NewRecorder();handler.ServeHTTP(w,preflight)
 if w.Code!=204||w.Header().Get("Access-Control-Allow-Origin")!="https://stocks.themauris.org"||w.Header().Get("Access-Control-Allow-Credentials")!="true"{t.Fatalf("preflight: %d %+v",w.Code,w.Header())}
 write:=func(origin string) *httptest.ResponseRecorder {
  r:=httptest.NewRequest(http.MethodPut,"https://home.themauris.org/api/storage/mauristocks",strings.NewReader(`{"revision":0,"data":{"symbols":["CAT"]}}`))
  r.AddCookie(cookie)
  if origin!=""{r.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder();handler.ServeHTTP(w,r);return w
 }
 if got:=write("https://evil.example").Code;got!=403{t.Fatalf("foreign origin: %d",got)}
 if got:=write("").Code;got!=403{t.Fatalf("missing origin: %d",got)}
 if got:=write("https://stocks.themauris.org").Code;got!=200{t.Fatalf("allowed origin: %d",got)}
 read:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/storage/mauristocks",nil)
 read.AddCookie(cookie);read.Header.Set("Origin","https://newsgo.themauris.org")
 w=httptest.NewRecorder();handler.ServeHTTP(w,read)
 if w.Code!=200||!strings.Contains(w.Body.String(),"CAT"){t.Fatalf("subdomain read: %d %s",w.Code,w.Body.String())}
 invalid:=httptest.NewRequest(http.MethodOptions,"https://home.themauris.org/api/storage/mauristocks",nil)
 invalid.Header.Set("Origin","https://attacker.themauris.org")
 invalid.Header.Set("Access-Control-Request-Method","PUT")
 w=httptest.NewRecorder();handler.ServeHTTP(w,invalid)
 if w.Code!=403||w.Header().Get("Access-Control-Allow-Origin")!=""{t.Fatalf("untrusted subdomain: %d",w.Code)}
}
