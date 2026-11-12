package main

import (
 "encoding/json"
 "io"
 "net/http"
 "strings"
 "time"
)

type homepagePresence struct {
 Seen time.Time
 Interaction time.Time
 Visible bool
 SignedIn bool
}
type homepagePing struct {
 ID string `json:"id"`
 Visible bool `json:"visible"`
 Interacted bool `json:"interacted"`
}

// homepagePresenceHandler receives anonymous or authenticated browser presence
// on Homepage's own origin. It never accepts cross-origin suite telemetry or
// reads/extends a login session. The anonymous identifier stays ephemeral.
func (c *activityCollector) homepagePresenceHandler(s *sessionService) http.HandlerFunc {
 return func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
  if !validSessionOrigin(r) {http.Error(w,"origin required",403);return}
  if r.Header.Get("Content-Type")!="application/json" {http.Error(w,"JSON required",415);return}
  r.Body=http.MaxBytesReader(w,r.Body,256)
  dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
  var p homepagePing
  if err:=dec.Decode(&p);err!=nil {http.Error(w,"bad ping",400);return}
  var extra any
  if dec.Decode(&extra)!=io.EOF || len(p.ID)!=32 || strings.Trim(p.ID,"0123456789abcdef")!="" {
   http.Error(w,"invalid presence identifier",400);return
  }
  _,_,authenticated:=s.authenticateSession(r)
  now:=c.now()
  c.mu.Lock()
  defer c.mu.Unlock()
  for id,entry:=range c.homePresence {
   if now.Sub(entry.Seen)>time.Hour {delete(c.homePresence,id)}
  }
  prior,exists:=c.homePresence[p.ID]
  if !exists && len(c.homePresence)>=5000 {http.Error(w,"presence capacity reached",503);return}
  if exists && now.Sub(prior.Seen)<15*time.Second {w.WriteHeader(http.StatusNoContent);return}
  prior.Seen=now
  prior.Visible=p.Visible
  prior.SignedIn=authenticated
  if p.Interacted && p.Visible {prior.Interaction=now}
  c.homePresence[p.ID]=prior
  c.updateHomepageSnapshotLocked(now)
  w.WriteHeader(http.StatusNoContent)
 }
}

func (c *activityCollector) updateHomepageSnapshotLocked(now time.Time) {
 p:=activityReport{}
 var recent time.Time
 for _,record:=range c.homePresence {
  age:=now.Sub(record.Seen)
  if age<0 || age>time.Hour {continue}
  if age<=60*time.Minute {p.Active60++}
  if age<=15*time.Minute {p.Active15++}
  if age<=5*time.Minute {p.Active5++}
  if age<=2*time.Minute {
   p.Active2++
   if record.Visible {p.Foreground++}else{p.Background++}
   if record.SignedIn {p.Authenticated++}else{p.Anonymous++}
  }
  if record.Interaction.After(recent) {recent=record.Interaction}
 }
 if !recent.IsZero(){p.LastInteraction=recent.UTC().Format(time.RFC3339)}
 c.latest["homepage"]=activityEntry{activityReport:p,Service:"homepage",Status:"fresh",LastReport:now.UTC().Format(time.RFC3339)}
}
