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

func TestAdminTokenLifecycleAndIsolation(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"mauristocks");if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("a",64)
 sum:=sha256.Sum256([]byte(token))
 authPath:=filepath.Join(root,"auth","approved-tokens.json")
 if err:=os.MkdirAll(filepath.Dir(authPath),0700);err!=nil{t.Fatal(err)}
 entries,_:=json.Marshal([]suitestore.Credential{{User:"owner",Role:"admin",Digest:hex.EncodeToString(sum[:])}})
 if err:=os.WriteFile(authPath,entries,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 request:=func(token string,method string,body string)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"https://home.themauris.org/api/admin/users",strings.NewReader(body))
  req.Header.Set("Authorization","Bearer "+token)
  req.Header.Set("Origin","https://home.themauris.org")
  w:=httptest.NewRecorder();svc.adminHandler(w,req);return w
 }
 issued:=request(token,http.MethodPost,`{"action":"issue","user":"reader","role":"user"}`)
 if issued.Code!=200{t.Fatalf("issue: %d %s",issued.Code,issued.Body.String())}
 var newUser map[string]string
 if err:=json.Unmarshal(issued.Body.Bytes(),&newUser);err!=nil{t.Fatal(err)}
 readerToken:=newUser["token"]
 if len(readerToken)!=64{t.Fatal("no newly issued token")}
 if user,role,err:=store.Authenticate(readerToken);err!=nil||user!="reader"||role!="user"{t.Fatalf("new token: %s %s %v",user,role,err)}
 if got:=request(readerToken,http.MethodGet,"").Code;got!=403{t.Fatalf("nonadmin was authorized: %d",got)}
 list:=request(token,http.MethodGet,"")
 if list.Code!=200||strings.Contains(list.Body.String(),readerToken)||strings.Contains(list.Body.String(),hex.EncodeToString(sum[:])){t.Fatalf("credential leaked: %s",list.Body.String())}
 _,err=store.Save("reader","mauristocks",0,json.RawMessage(`{"symbols":["CAT"]}`));if err!=nil{t.Fatal(err)}
 rotated:=request(token,http.MethodPost,`{"action":"rotate","user":"reader"}`)
 if rotated.Code!=200{t.Fatalf("rotate: %d %s",rotated.Code,rotated.Body.String())}
 var updated map[string]string
 if err=json.Unmarshal(rotated.Body.Bytes(),&updated);err!=nil{t.Fatal(err)}
 if _,_,err=store.Authenticate(readerToken);err==nil{t.Fatal("old token still works")}
 if _,role,err:=store.Authenticate(updated["token"]);err!=nil||role!="user"{t.Fatalf("rotated token role: %s %v",role,err)}
 if saved,err:=store.Save("reader","mauristocks",1,json.RawMessage(`{}`));err!=nil||saved.Revision!=2{t.Fatalf("saved data lost after rotation: %v",err)}
 if got:=request(token,http.MethodPost,`{"action":"revoke","user":"owner"}`).Code;got!=400{t.Fatalf("last admin revoked: %d",got)}
 if got:=request(token,http.MethodPost,`{"action":"revoke","user":"reader"}`).Code;got!=200{t.Fatalf("revoke user: %d",got)}
 if _,_,err=store.Authenticate(updated["token"]);err==nil{t.Fatal("revoked token still works")}
}
