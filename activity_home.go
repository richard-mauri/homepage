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
 SessionKey string
 UserKey string
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
  // Use the authenticated server-side session digest for telemetry identity.
  // Raw session IDs and personal credentials never enter the presence map.
  _,_,authenticated:=s.authenticateSession(r)
  sessionKey,userKey:="",""
  if strings.HasPrefix(r.Header.Get("Authorization"),"Session ") {
   // Local Safari can use the page's ephemeral Session header when its
   // Secure cookie is unavailable on loopback HTTP.
   record,_,_,ok:=s.homepageSession(r)
   authenticated=ok
   if ok {
    raw,_:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
    sessionKey,userKey=sessionLookupKey(raw),record.Digest
   }
  }else if authenticated {
   if cookie,err:=r.Cookie(sessionCookieName);err==nil {
    key:=sessionLookupKey(cookie.Value)
    s.mu.Lock();record,exists:=s.sessions[key];s.mu.Unlock()
    if exists && s.now().Before(record.Expires) {sessionKey,userKey=key,record.Digest}
   }
  }
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
  prior.SignedIn=authenticated && sessionKey!=""
  prior.SessionKey=sessionKey
  prior.UserKey=userKey
  if p.Interacted && p.Visible {prior.Interaction=now}
  c.homePresence[p.ID]=prior
  c.updateHomepageSnapshotLocked(now)
  if c.homeLastSample.IsZero() || now.Sub(c.homeLastSample)>=time.Minute {
   if err:=c.appendHistoryLocked("homepage",c.latest["homepage"].activityReport,now);err!=nil {
    http.Error(w,"activity history unavailable",503);return
   }
   c.homeLastSample=now
  }
  w.WriteHeader(http.StatusNoContent)
 }
}

func (c *activityCollector) updateHomepageSnapshotLocked(now time.Time) {
 p:=activityReport{}
 var recent time.Time
 var lastSeen time.Time
 // One identity per authenticated Homepage session, regardless of how many
 // tabs share it. Anonymous tabs retain their separate ephemeral identities.
 identities:=make(map[string]homepagePresence,len(c.homePresence))
 for id,record:=range c.homePresence {
  key:="guest:"+id
  if record.SignedIn && record.SessionKey!="" {key="session:"+record.SessionKey}
  previous,ok:=identities[key]
  if !ok || record.Seen.After(previous.Seen) {identities[key]=record}
  if record.Interaction.After(recent) {recent=record.Interaction}
 }
 users:=make(map[string]bool)
 for _,record:=range identities {
  if record.Seen.After(lastSeen) {lastSeen=record.Seen}
  age:=now.Sub(record.Seen)
  if age<0 || age>time.Hour {continue}
  p.Active60++
  if age<=15*time.Minute {p.Active15++}
  if age<=5*time.Minute {p.Active5++}
  if age<=2*time.Minute {
   p.Active2++
   if record.Visible {p.Foreground++}else{p.Background++}
   if record.SignedIn {
    p.Authenticated++
    if record.UserKey!="" {users[record.UserKey]=true}
   }else{p.Anonymous++}
  }
 }
 p.UniqueUsers=len(users)
 if !recent.IsZero(){p.LastInteraction=recent.UTC().Format(time.RFC3339)}
 status:="stale"
 if now.Sub(lastSeen)<=150*time.Second {status="fresh"}
 c.latest["homepage"]=activityEntry{activityReport:p,Service:"homepage",Status:status,LastReport:lastSeen.UTC().Format(time.RFC3339)}
}
