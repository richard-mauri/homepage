package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "themauris.org/homepage/internal/suitestore"
)

func TestGuestSessionMemoryOnlyExpiryHandoffAndStorageDenial(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 now:=time.Date(2026,11,13,12,0,0,0,time.UTC)
 svc.now=func()time.Time{return now}
 create:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session/guest",nil)
 create.Header.Set("Origin","https://home.themauris.org")
 w:=httptest.NewRecorder();svc.guestSessionHandler(w,create)
 if w.Code!=200{t.Fatalf("guest issue: %d %s",w.Code,w.Body.String())}
 var issued struct{User,Role,SessionToken,ExpiresAt string}
 if err:=json.Unmarshal(w.Body.Bytes(),&issued);err!=nil{t.Fatal(err)}
 if !strings.HasPrefix(issued.User,"guest-") || issued.Role!="guest" || len(issued.SessionToken)!=64 {
  t.Fatalf("invalid guest identity: %+v",issued)
 }
 if issued.ExpiresAt!=now.Add(time.Hour).Format(time.RFC3339){t.Fatalf("guest expiry %s",issued.ExpiresAt)}
 if len(w.Result().Cookies())!=1 || w.Result().Cookies()[0].Name!=guestCookieName {
  t.Fatal("guest cookie missing")
 }
 req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 req.Header.Set("Authorization","Session "+issued.SessionToken)
 current:=httptest.NewRecorder();svc.handler(current,req)
 if current.Code!=200||!strings.Contains(current.Body.String(),issued.User){t.Fatal("guest session unavailable")}
 handoffReq:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
 handoffReq.Header.Set("Origin","https://home.themauris.org")
 handoffReq.Header.Set("Authorization","Session "+issued.SessionToken)
 handoff:=httptest.NewRecorder();svc.handoffHandler(handoff,handoffReq)
 if handoff.Code!=200{t.Fatalf("guest handoff: %d %s",handoff.Code,handoff.Body.String())}
 var child struct{Token,ExpiresAt string}
 if err=json.Unmarshal(handoff.Body.Bytes(),&child);err!=nil{t.Fatal(err)}
 if child.ExpiresAt!=issued.ExpiresAt{t.Fatal("guest handoff extended session")}
 profileReq:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
 profileReq.Header.Set("Origin","https://scrapergo.themauris.org")
 profileReq.Header.Set("Authorization","Bearer "+child.Token)
 profile:=httptest.NewRecorder();svc.profileHandler(profile,profileReq)
 if profile.Code!=200 || !strings.Contains(profile.Body.String(),`"role":"guest"`) ||
  !strings.Contains(profile.Body.String(),`"activitySession"`) {t.Fatalf("guest profile: %d %s",profile.Code,profile.Body.String())}
 passed:=false
 protected:=svc.storageHandler(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){passed=true;w.WriteHeader(204)}))
 blocked:=httptest.NewRecorder();protected.ServeHTTP(blocked,profileReq)
 if blocked.Code!=403 || passed {t.Fatal("guest accessed persisted user settings")}
 if len(svc.sessions)!=0 || len(svc.handoffs)!=0 {t.Fatal("guest entered durable named-user maps")}
 restarted:=newSessionService(store)
 if _,_,ok:=restarted.guestSession(req);ok {t.Fatal("guest survived Homepage restart")}
 now=now.Add(time.Hour+time.Second)
 expired:=httptest.NewRecorder();svc.handler(expired,req)
 if expired.Code!=401{t.Fatalf("expired guest still authenticated: %d",expired.Code)}
 svc.mu.Lock();svc.cleanupGuestsLocked();n,m:=len(svc.guests),len(svc.guestHandoffs);svc.mu.Unlock()
 if n!=0||m!=0 {t.Fatalf("expired guest records retained: %d / %d",n,m)}
}

func TestGuestIssuanceRequiresSameOriginAndNeverRenews(t *testing.T) {
 store,err:=suitestore.New(t.TempDir(),"newsgo")
 if err!=nil{t.Fatal(err)}
 svc:=newSessionService(store)
 now:=time.Now().UTC().Truncate(time.Second)
 svc.now=func()time.Time{return now}
 request:=func(origin string,cookie *http.Cookie)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session/guest",nil)
  req.Header.Set("Origin",origin)
  if cookie!=nil {req.AddCookie(cookie)}
  w:=httptest.NewRecorder();svc.guestSessionHandler(w,req);return w
 }
 if got:=request("https://evil.example",nil).Code;got!=403{t.Fatalf("cross-origin guest issue: %d",got)}
 first:=request("https://home.themauris.org",nil)
 if first.Code!=200{t.Fatalf("first guest: %d",first.Code)}
 cookie:=first.Result().Cookies()[0]
 now=now.Add(30*time.Minute)
 second:=request("https://home.themauris.org",cookie)
 if second.Code!=200 {t.Fatalf("existing guest reuse: %d",second.Code)}
 var a,b map[string]string
 _=json.Unmarshal(first.Body.Bytes(),&a);_=json.Unmarshal(second.Body.Bytes(),&b)
 if a["sessionToken"]!=b["sessionToken"] || a["expiresAt"]!=b["expiresAt"] {
  t.Fatal("guest token silently rotated or extended")
 }
}
