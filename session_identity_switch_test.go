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

// Switching identities is only possible after a successful server-side sign-out.
func TestSessionIdentitySwitchRequiresSignOut(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"mauristocks")
 if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("b",64)
 digest:=sha256.Sum256([]byte(token))
 if err:=os.MkdirAll(filepath.Join(root,"auth"),0700);err!=nil{t.Fatal(err)}
 records,err:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(digest[:]),Role:"user"}})
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(root,"auth","approved-tokens.json"),records,0600);err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 personal:=func(cookies ...*http.Cookie)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  req.Header.Set("Authorization","Bearer "+token)
  for _,cookie:=range cookies {req.AddCookie(cookie)}
  out:=httptest.NewRecorder();svc.handler(out,req);return out
 }
 guest:=func(cookies ...*http.Cookie)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session/guest",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  for _,cookie:=range cookies {req.AddCookie(cookie)}
  out:=httptest.NewRecorder();svc.guestSessionHandler(out,req);return out
 }
 signOut:=func(cookies ...*http.Cookie)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(http.MethodDelete,"https://home.themauris.org/api/session",nil)
  req.Header.Set("Origin","https://home.themauris.org")
  for _,cookie:=range cookies {req.AddCookie(cookie)}
  out:=httptest.NewRecorder();svc.handler(out,req);return out
 }
 status:=func(cookie *http.Cookie)int{
  req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
  req.AddCookie(cookie)
  out:=httptest.NewRecorder();svc.handler(out,req);return out.Code
 }

 // A current Guest session cannot silently become a personal identity.
 guestLogin:=guest()
 if guestLogin.Code!=200{t.Fatalf("guest login %d",guestLogin.Code)}
 guestCookie:=guestLogin.Result().Cookies()[0]
 if changed:=personal(guestCookie);changed.Code!=409{t.Fatalf("guest->personal without sign-out: %d",changed.Code)}
 if got:=status(guestCookie);got!=200{t.Fatalf("blocked switch revoked guest unexpectedly: %d",got)}
 if out:=signOut(guestCookie);out.Code!=204{t.Fatalf("guest sign-out %d",out.Code)}
 if got:=status(guestCookie);got!=401{t.Fatalf("guest survived sign-out: %d",got)}
 named:=personal()
 if named.Code!=200{t.Fatalf("personal login after guest sign-out: %d",named.Code)}
 namedCookie:=named.Result().Cookies()[0]
 // An authenticated personal user cannot silently create Guest or replace
 // the personal session, even if they present another valid personal token.
 if changed:=guest(namedCookie);changed.Code!=409{t.Fatalf("personal->guest without sign-out: %d",changed.Code)}
 if changed:=personal(namedCookie);changed.Code!=409{t.Fatalf("personal->personal without sign-out: %d",changed.Code)}
 if got:=status(namedCookie);got!=200{t.Fatalf("blocked switch revoked personal account: %d",got)}
 if out:=signOut(namedCookie);out.Code!=204{t.Fatalf("personal sign-out %d",out.Code)}
 if got:=status(namedCookie);got!=401{t.Fatalf("personal session survived sign-out: %d",got)}
 if fresh:=guest();fresh.Code!=200{t.Fatalf("guest login after personal sign-out %d",fresh.Code)}
}
