package main

import (
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestVerifiedGuestLaunchAcknowledgment(t *testing.T) {
 s:=newSessionService(nil)
 c:=newActivityCollector()
 now:=time.Now().UTC().Truncate(time.Second)
 s.now=func()time.Time{return now}
 c.now=func()time.Time{return now}
 token:=strings.Repeat("a",64)
 key:=sessionLookupKey(token)
 guestName:="guest-1234abcd"
 s.guests[key]=guestRecord{User:guestName,Expires:now.Add(time.Hour)}
 request:=func(service,origin,session string)int {
  r:=httptest.NewRequest("POST","https://home.themauris.org/api/activity/launch",strings.NewReader(`{"service":"`+service+`"}`))
  r.Header.Set("Origin",origin)
  r.Header.Set("Content-Type","application/json")
  if session!="" {r.Header.Set("Authorization","Session "+session)}
  w:=httptest.NewRecorder();c.launchHandler(s)(w,r);return w.Code
 }
 if code:=request("newsgo","https://evil.example",token);code!=403 {t.Fatalf("cross origin %d",code)}
 if code:=request("other","https://home.themauris.org",token);code!=400 {t.Fatalf("unknown service %d",code)}
 if code:=request("newsgo","https://home.themauris.org","");code!=401 {t.Fatalf("missing identity %d",code)}
 if code:=request("newsgo","https://home.themauris.org",token);code!=204 {t.Fatalf("guest launch %d",code)}
 launches:=c.launchUsersSnapshot(s)
 for _,service:=range []string{"homepage","newsgo"} {
  entries:=launches[service]
  if len(entries)!=1||entries[0].User!=guestName||entries[0].Sessions!=0||entries[0].LastSeen=="" {t.Fatalf("%s launch %+v",service,entries)}
 }
 if len(c.homePresence)!=0 {t.Fatal("launch must not manufacture active browser presence")}
 if len(c.snapshot())!=5 {t.Fatal("unexpected services snapshot")}
 delete(s.guests,key)
 if len(c.launchUsersSnapshot(s))!=0 {t.Fatal("revoked guest launch still shown")}
 s.guests[key]=guestRecord{User:guestName,Expires:now.Add(time.Minute)}
 if code:=request("scrapergo","https://home.themauris.org",token);code!=204 {t.Fatalf("new guest launch %d",code)}
 now=now.Add(2*time.Minute)
 if len(c.launchUsersSnapshot(s))!=0 {t.Fatal("expired guest launch still shown")}
 if code:=request("newsgo","https://home.themauris.org",token);code!=401 {t.Fatalf("expired guest %d",code)}
 now=now.Add(time.Hour)
 _=c.launchUsersSnapshot(s)
 if len(c.launches)!=0 {t.Fatal("expired launch memory retained")}
}

func TestHomepageLauncherAcknowledgesVerifiedNavigation(t *testing.T) {
 script,err:=assets.ReadFile("static/app.js");if err!=nil{t.Fatal(err)}
 js:=string(script)
 for _,snippet:=range []string{"'/api/activity/launch'","await acknowledgeSuiteLaunch(","signal:controller.signal","url.hash='suite_token='"} {
  if !strings.Contains(js,snippet) {t.Errorf("missing launch behavior %q",snippet)}
 }
 page,err:=assets.ReadFile("static/suite-admin.html");if err!=nil{t.Fatal(err)}
 if !strings.Contains(string(page),"data.launchUsers||{}") {t.Fatal("admin view lacks launch fallback")}
}
