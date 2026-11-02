package suitestore

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestPerUserStorageAndConflicts(t *testing.T){
 s,err:=New(t.TempDir(),"newsgo","mauristocks");if err!=nil{t.Fatal(err)}
 a,err:=s.Save("alice","mauristocks",0,json.RawMessage(`{"symbols":["NVDA"]}`))
 if err!=nil || a.Revision!=1{t.Fatalf("first save: %v, %+v",err,a)}
 if _,err=s.Save("alice","mauristocks",0,json.RawMessage(`{}`));!errors.Is(err,ErrConflict){t.Fatalf("expected conflict; got %v",err)}
 if _,exists,err:=s.load("bob","mauristocks");err!=nil||exists{t.Fatalf("cross-user leak: %v %v",exists,err)}
 if _,err=s.Save("../escape","newsgo",0,json.RawMessage(`{}`));err==nil{t.Fatal("unsafe identity accepted")}
 if _,err=s.Save("alice","unknown",0,json.RawMessage(`{}`));err==nil{t.Fatal("unsafe namespace accepted")}
 b,err:=os.ReadFile(s.recordPath("alice","mauristocks"));if err!=nil{t.Fatal(err)}
 var rec Record;if err=json.Unmarshal(b,&rec);err!=nil||rec.Revision!=1{t.Fatalf("bad persisted record: %v %+v",err,rec)}
}

func TestAuthenticationAndHTTP(t *testing.T){
 s,err:=New(t.TempDir(),"newsgo","mauristocks");if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("a",64)
 sum:=sha256.Sum256([]byte(token))
 if err=os.MkdirAll(filepath.Dir(s.credentialPath()),0700);err!=nil{t.Fatal(err)}
 b,_:=json.Marshal([]Credential{{User:"alice",Digest:hex.EncodeToString(sum[:]),Role:"admin"}})
 if err=os.WriteFile(s.credentialPath(),b,0600);err!=nil{t.Fatal(err)}
 if user,role,err:=s.Authenticate(token);err!=nil||user!="alice"||role!="admin"{t.Fatalf("auth: %q %q %v",user,role,err)}
 handler:=s.Handler()
 req:=httptest.NewRequest(http.MethodPut,"https://home.themauris.org/api/storage/mauristocks",strings.NewReader(`{"revision":0,"data":{"symbols":["CAT"]}}`))
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Origin","https://home.themauris.org")
 w:=httptest.NewRecorder();handler.ServeHTTP(w,req)
 if w.Code!=200{t.Fatalf("save status %d: %s",w.Code,w.Body.String())}
 req=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/storage/mauristocks",nil)
 req.Header.Set("Authorization","Bearer "+token)
 w=httptest.NewRecorder();handler.ServeHTTP(w,req)
 if w.Code!=200||!strings.Contains(w.Body.String(),"CAT"){t.Fatalf("load status %d: %s",w.Code,w.Body.String())}
 req=httptest.NewRequest(http.MethodPut,"https://home.themauris.org/api/storage/mauristocks",strings.NewReader(`{"revision":1,"data":{}}`))
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Origin","https://evil.example")
 w=httptest.NewRecorder();handler.ServeHTTP(w,req)
 if w.Code!=403{t.Fatalf("cross-origin write status: %d",w.Code)}
 req=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/storage/mauristocks",nil)
 w=httptest.NewRecorder();handler.ServeHTTP(w,req)
 if w.Code!=401{t.Fatalf("unauthenticated read status: %d",w.Code)}
}
