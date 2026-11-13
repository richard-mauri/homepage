package main

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestHomepageAnonymousPresenceAndStaleness(t *testing.T) {
 c:=newActivityCollectorWithRoot(t.TempDir())
 s:=newSessionService(nil)
 now:=time.Now().UTC().Truncate(time.Second)
 c.now=func()time.Time{return now}
 id:=strings.Repeat("a",32)
 ping:=func(origin,body string)int {
  r:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8787/api/activity/ping",strings.NewReader(body))
  r.Header.Set("Origin",origin)
  r.Header.Set("Content-Type","application/json")
  w:=httptest.NewRecorder()
  c.homepagePresenceHandler(s)(w,r)
  return w.Code
 }
 body:=`{"id":"`+id+`","visible":true,"interacted":true}`
 if v:=ping("https://other.example",body);v!=403 {t.Fatalf("cross-origin %d",v)}
 if v:=ping("http://127.0.0.1:8787",body);v!=204 {t.Fatalf("anonymous ping %d",v)}
 var found bool
 for _,e:=range c.snapshot() {
  if e.Service=="homepage" {
   found=true
   if e.Status!="fresh" || e.Active2!=1 || e.Anonymous!=1 || e.Foreground!=1 || e.Authenticated!=0 || e.LastInteraction=="" {t.Fatalf("unexpected %+v",e)}
  }
 }
 if !found {t.Fatal("homepage missing")}
 hours,status:=c.historySnapshot()
 if status!="available" || len(hours)!=1 || hours[0].Samples!=1 {t.Fatalf("unexpected history %s %+v",status,hours)}
 now=now.Add(3*time.Minute)
 for _,e:=range c.snapshot() {
  if e.Service=="homepage" && (e.Status!="stale" || e.Active2!=0) {t.Fatalf("stale homepage looks busy or fresh: %+v",e)}
 }
}
func TestHomepagePresenceBounds(t *testing.T){
 c:=newActivityCollectorWithRoot(t.TempDir())
 s:=newSessionService(nil)
 r:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:8787/api/activity/ping",nil)
 w:=httptest.NewRecorder();c.homepagePresenceHandler(s)(w,r)
 if w.Code!=http.StatusMethodNotAllowed {t.Fatalf("GET %d",w.Code)}
 r=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8787/api/activity/ping",strings.NewReader(`{"id":"bad","visible":true}`))
 r.Header.Set("Origin","http://127.0.0.1:8787")
 r.Header.Set("Content-Type","application/json")
 w=httptest.NewRecorder();c.homepagePresenceHandler(s)(w,r)
 if w.Code!=400{t.Fatalf("invalid id %d",w.Code)}
}

func TestHomepageAdminPageIncludesPresenceHeartbeat(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 if strings.Count(html, `src="/static/activity.js"`)!=1 {
  t.Fatal("admin page must load the Homepage presence reporter exactly once")
 }
 if !strings.Contains(html,`<script src="/static/activity.js" defer></script>`) {
  t.Fatal("admin page must load the reporter without blocking the dashboard")
 }
 reporter,err:=assets.ReadFile("static/activity.js")
 if err!=nil {t.Fatal(err)}
 if !strings.Contains(string(reporter),"/api/activity/ping") {
  t.Fatal("presence script must target Homepage's same-origin activity route")
 }
}

func TestHomepageAdminCSPPermitsHeartbeatScript(t *testing.T) {
 if !strings.Contains(adminPageCSP,"script-src 'self' 'unsafe-inline'") {
  t.Fatal("administrator CSP blocks its same-origin presence script")
 }
 if !strings.Contains(adminPageCSP,"connect-src 'self'") {
  t.Fatal("administrator CSP blocks its same-origin presence ping")
 }
 if !strings.Contains(adminPageCSP,"default-src 'none'") || !strings.Contains(adminPageCSP,"frame-ancestors 'none'") {
  t.Fatal("administrator CSP lost restrictive baseline")
 }
}

func TestHomepageSessionDedupAndUniqueUsers(t *testing.T) {
 c:=newActivityCollectorWithRoot(t.TempDir())
 now:=time.Now().UTC().Truncate(time.Second)
 c.now=func()time.Time{return now}
 user:="same-approved-token-digest"
 c.homePresence=map[string]homepagePresence{
  "tab-computer-one":{Seen:now,Visible:true,SignedIn:true,SessionKey:"session-one",UserKey:user},
  "another-tab-on-one":{Seen:now,Visible:true,SignedIn:true,SessionKey:"session-one",UserKey:user},
  "tab-computer-two":{Seen:now,Visible:true,SignedIn:true,SessionKey:"session-two",UserKey:user},
  "guest-tab":{Seen:now,Visible:true},
 }
 var got activityEntry
 for _,entry:=range c.snapshot(){if entry.Service=="homepage"{got=entry}}
 if got.Active2!=3 || got.Active15!=3 || got.Active60!=3 ||
  got.Authenticated!=2 || got.UniqueUsers!=1 || got.Anonymous!=1 {
  t.Fatalf("two computers, one user and one guest should count as 2 sessions, 1 user, 1 guest: %+v",got)
 }
 // Presence windows expire without ending or extending server login state.
 now=now.Add(3*time.Minute)
 for _,entry:=range c.snapshot(){if entry.Service=="homepage"{
  if entry.Active2!=0 || entry.Authenticated!=0 || entry.UniqueUsers!=0 {
   t.Fatalf("old observations should leave two-minute window: %+v",entry)
  }
 }}
}

func TestSuiteAdminShowsNewsGoPerUserActivity(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil{t.Fatal(err)}
 html:=string(page)
 for _,required:=range []string{
  "const newsUsers=Array.isArray(news?.users)?news.users:[]",
  "['NewsGo',mergeObserved(newsUsers,'newsgo')]",
  "key==='homepage'||key==='newsgo'||key==='scrapergo'",
  "Verified Homepage launches are shown for all suite apps",
 } {
  if !strings.Contains(html,required){t.Errorf("NewsGo session-aware UI missing %q",required)}
 }
}
