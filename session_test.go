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
