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
