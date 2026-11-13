package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strconv"
 "strings"
 "testing"
 "time"
)

func TestActivityReportValidation(t *testing.T) {
 good:=activityReport{Active2:1,Active5:2,Active15:3,Active60:4,Foreground:1,Anonymous:1}
 if !validActivityReport(good){t.Fatal("good report rejected")}
 bad:=good;bad.Active5=0
 if validActivityReport(bad){t.Fatal("invalid rolling counts accepted")}
 bad=good;bad.InFlight=-1
 if validActivityReport(bad){t.Fatal("negative work accepted")}
}
func TestActivityIngestAndReplay(t *testing.T) {
 c:=newActivityCollectorWithRoot(t.TempDir())
 c.secrets["newsgo"]=strings.Repeat("s",40)
 now:=time.Now().Truncate(time.Second)
 c.now=func()time.Time{return now}
 send:=func(key,nonce,origin string)int{
  r:=httptest.NewRequest(http.MethodPost,"/api/admin/activity/ingest",strings.NewReader(`{"active2":1,"active5":1,"active15":1,"active60":1,"foreground":1,"anonymous":1,"inFlight":0}`))
  r.Header.Set("Content-Type","application/json")
  r.Header.Set("Authorization","Bearer "+key)
  r.Header.Set("X-Activity-Service","newsgo")
  r.Header.Set("X-Activity-Timestamp",strconv.FormatInt(now.Unix(),10))
  r.Header.Set("X-Activity-Nonce",nonce)
  if origin!=""{r.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder();c.ingest(w,r);return w.Code
 }
 if v:=send("wrong",strings.Repeat("a",16),"");v!=401 {t.Fatalf("unauthorized %d",v)}
 if v:=send(strings.Repeat("s",40),strings.Repeat("b",16),"https://newsgo.themauris.org");v!=403 {t.Fatalf("browser request %d",v)}
 if v:=send(strings.Repeat("s",40),strings.Repeat("c",16),"");v!=204 {t.Fatalf("authorized %d",v)}
 if v:=send(strings.Repeat("s",40),strings.Repeat("c",16),"");v!=409 {t.Fatalf("replay %d",v)}
 entries:=c.snapshot()
 for _,e:=range entries {
  if e.Service=="newsgo" && (e.Status!="fresh" || e.Active2!=1){t.Fatal("report missing")}
  if e.Service=="scrapergo" && e.Status!="unknown" {t.Fatal("unreported service incorrectly idle")}
 }
 c.now=func()time.Time{return now.Add(3*time.Minute)}
 for _,e:=range c.snapshot(){if e.Service=="newsgo" && e.Status!="stale"{t.Fatal("stale report not marked")}}
}
func TestActivityAdminRequiresSession(t *testing.T){
 c:=newActivityCollector()
 s:=newSessionService(nil)
 r:=httptest.NewRequest(http.MethodGet,"/api/admin/activity",nil)
 w:=httptest.NewRecorder();c.readAdmin(s)(w,r)
 if w.Code!=401 {t.Fatalf("anonymous access %d",w.Code)}
 if json.Valid(w.Body.Bytes()) {t.Fatal("unauthenticated request returned JSON metrics")}
}

func TestNewsGoPerUserReportAcceptedAndOtherServicesRemainRestricted(t *testing.T) {
 c:=newActivityCollectorWithRoot(t.TempDir())
 c.secrets["newsgo"]=strings.Repeat("n",40)
 c.secrets["stocks"]=strings.Repeat("k",40)
 now:=time.Now().UTC().Truncate(time.Second)
 c.now=func()time.Time{return now}
 report:=activityReport{Active2:1,Active5:1,Active15:1,Active60:1,Authenticated:1,UniqueUsers:1,
  Users:[]activityUserSummary{{User:"richard",Sessions:1,LastSeen:now.Format(time.RFC3339)}}}
 payload,err:=json.Marshal(report)
 if err!=nil{t.Fatal(err)}
 send:=func(service,secret,nonce string)int{
  r:=httptest.NewRequest(http.MethodPost,"/api/admin/activity/ingest",strings.NewReader(string(payload)))
  r.Header.Set("Content-Type","application/json")
  r.Header.Set("Authorization","Bearer "+secret)
  r.Header.Set("X-Activity-Service",service)
  r.Header.Set("X-Activity-Timestamp",strconv.FormatInt(now.Unix(),10))
  r.Header.Set("X-Activity-Nonce",nonce)
  w:=httptest.NewRecorder();c.ingest(w,r);return w.Code
 }
 if got:=send("newsgo",strings.Repeat("n",40),strings.Repeat("a",16));got!=204{
  t.Fatalf("NewsGo user report rejected: %d",got)
 }
 if got:=send("stocks",strings.Repeat("k",40),strings.Repeat("b",16));got!=400{
  t.Fatalf("unmigrated service accepted user report: %d",got)
 }
 for _,entry:=range c.snapshot(){
  if entry.Service=="newsgo" && (entry.Status!="fresh"||len(entry.Users)!=1||entry.Users[0].User!="richard"){
   t.Fatalf("NewsGo user activity missing: %+v",entry)
  }
 }
}
