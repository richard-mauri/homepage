package main

import (
 "crypto/subtle"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "strconv"
 "strings"
 "sync"
 "time"
)

var activityServices = []string{"homepage","newsgo","stocks","sailing","scrapergo"}

type activityReport struct {
 Active2 int `json:"active2"`
 Active5 int `json:"active5"`
 Active15 int `json:"active15"`
 Active60 int `json:"active60"`
 Foreground int `json:"foreground"`
 Background int `json:"background"`
 Anonymous int `json:"anonymous"`
 Authenticated int `json:"authenticated"`
 InFlight int `json:"inFlight"`
 LastInteraction string `json:"lastInteraction,omitempty"`
}
type activityEntry struct {
 activityReport
 Service string `json:"service"`
 Status string `json:"status"`
 LastReport string `json:"lastReport,omitempty"`
}
type activityCollector struct {
 mu sync.Mutex
 latest map[string]activityEntry
 seen map[string]time.Time
 secrets map[string]string
 history map[string]activityHour
 homePresence map[string]homepagePresence
 historyPath string
 historyError string
 now func() time.Time
}
func newActivityCollector() *activityCollector {
 c:=&activityCollector{latest:map[string]activityEntry{},seen:map[string]time.Time{},secrets:map[string]string{},history:map[string]activityHour{},homePresence:map[string]homepagePresence{},now:time.Now}
 for _,name:=range activityServices {
  key:="HOMEPAGE_ACTIVITY_KEY_"+strings.ToUpper(name)
  c.secrets[name]=os.Getenv(key)
 }
 return c
}
func validActivityReport(p activityReport) bool {
 nums:=[]int{p.Active2,p.Active5,p.Active15,p.Active60,p.Foreground,p.Background,p.Anonymous,p.Authenticated,p.InFlight}
 for _,n:=range nums {if n<0 || n>100000 {return false}}
 if p.Active2>p.Active5 || p.Active5>p.Active15 || p.Active15>p.Active60{return false}
 if p.Foreground>p.Active2 || p.Background>p.Active2 || p.Anonymous>p.Active2 || p.Authenticated>p.Active2{return false}
 if len(p.LastInteraction)>0 {t,e:=time.Parse(time.RFC3339,p.LastInteraction);if e!=nil || t.After(time.Now().Add(2*time.Minute)){return false}}
 return true
}
func (c *activityCollector) ingest(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
 // Service-to-service only: browsers are never trusted telemetry clients.
 if r.Header.Get("Origin")!="" {http.Error(w,"browser origin not permitted",403);return}
 service:=r.Header.Get("X-Activity-Service")
 expected,exists:=c.secrets[service]
 if !exists || len(expected)<32 {http.Error(w,"collector not configured for service",503);return}
 provided,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 if !ok || subtle.ConstantTimeCompare([]byte(provided),[]byte(expected))!=1 {http.Error(w,"unauthorized",401);return}
 if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")),"application/json") {http.Error(w,"JSON required",415);return}
 ts,err:=strconv.ParseInt(r.Header.Get("X-Activity-Timestamp"),10,64)
 nonce:=r.Header.Get("X-Activity-Nonce")
 if err!=nil || len(nonce)<16 || len(nonce)>128 {http.Error(w,"invalid report metadata",400);return}
 now:=c.now()
 if d:=now.Sub(time.Unix(ts,0));d>2*time.Minute || d< -2*time.Minute {http.Error(w,"stale report",400);return}
 r.Body=http.MaxBytesReader(w,r.Body,2048)
 dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
 var p activityReport
 if err=dec.Decode(&p);err!=nil {http.Error(w,"invalid report",400);return}
 var extra any
 if dec.Decode(&extra)!=io.EOF || !validActivityReport(p){http.Error(w,"invalid report",400);return}
 replayKey:=service+":"+nonce
 c.mu.Lock()
 defer c.mu.Unlock()
 for k,until:=range c.seen {if now.After(until){delete(c.seen,k)}}
 if len(c.seen)>10000 {http.Error(w,"collector overloaded",503);return}
 if _,exists:=c.seen[replayKey];exists {http.Error(w,"duplicate report",409);return}
 if err:=c.appendHistoryLocked(service,p,now);err!=nil {http.Error(w,"activity history unavailable",503);return}
 c.seen[replayKey]=now.Add(3*time.Minute)
 c.latest[service]=activityEntry{activityReport:p,Service:service,Status:"fresh",LastReport:now.UTC().Format(time.RFC3339)}
 w.WriteHeader(http.StatusNoContent)
}
func (c *activityCollector) snapshot() []activityEntry {
 c.mu.Lock();defer c.mu.Unlock()
 now:=c.now()
 if len(c.homePresence)>0 {c.updateHomepageSnapshotLocked(now)}
 out:=make([]activityEntry,0,len(activityServices))
 for _,service:=range activityServices {
  entry,ok:=c.latest[service]
  if !ok {out=append(out,activityEntry{Service:service,Status:"unknown"});continue}
  t,err:=time.Parse(time.RFC3339,entry.LastReport)
  if err!=nil || now.Sub(t)>150*time.Second {entry.Status="stale"}
  out=append(out,entry)
 }
 return out
}
func (c *activityCollector) readAdmin(s *sessionService) http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  if r.Method!=http.MethodGet {w.Header().Set("Allow","GET");http.Error(w,"method not allowed",405);return}
  if origin:=r.Header.Get("Origin");origin!="" && !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
  // Do not accept credential-only tokens or cross-app handoffs.
  var role string;var ok bool
  if strings.HasPrefix(r.Header.Get("Authorization"),"Session ") {_,_,role,ok=s.homepageSession(r)}else{_,role,ok=s.authenticateSession(r)}
  if !ok {http.Error(w,"session required",401);return}
  if role!="admin" {http.Error(w,"admin required",403);return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  history,status:=c.historySnapshot()
  _=json.NewEncoder(w).Encode(map[string]any{"services":c.snapshot(),"history":history,"historyStatus":status,"advisoryOnly":true})
 }
}
// activityStatePath is reserved for future aggregate persistence; raw presence
// and replay identities are intentionally not written to disk.
func activityStatePath(root string) (string,error) {
 if root=="" {return "",errors.New("persistent suite storage required")}
 return filepath.Join(root,"activity","hourly.json"),nil
}
