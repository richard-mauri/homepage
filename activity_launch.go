package main

import (
 "encoding/json"
 "io"
 "net/http"
 "strings"
 "time"
)

// launchObservation is an ephemeral, verified navigation event. It deliberately
// does not enter the presence/heartbeat map or count as an active session.
type launchObservation struct {
 Seen time.Time
 Service string
 UserDigest string
 GuestKey string
 GuestUser string
}

func (c *activityCollector) launchHandler(s *sessionService) http.HandlerFunc {
 return func(w http.ResponseWriter,r *http.Request) {
  w.Header().Set("Cache-Control","no-store")
  if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
  if !validSessionOrigin(r) {http.Error(w,"origin required",403);return}
  if r.Header.Get("Content-Type")!="application/json" {http.Error(w,"JSON required",415);return}
  r.Body=http.MaxBytesReader(w,r.Body,128)
  var request struct{Service string `json:"service"`}
  decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields()
  if decoder.Decode(&request)!=nil {http.Error(w,"invalid launch",400);return}
  var trailing any
  if decoder.Decode(&trailing)!=io.EOF {http.Error(w,"invalid launch",400);return}
  switch request.Service {case "homepage","newsgo","scrapergo","stocks","sailing":default:http.Error(w,"unknown service",400);return}

  observation:=launchObservation{Service:request.Service}
  identityKey:=""
  if record,_,_,ok:=s.homepageSession(r);ok {
   observation.UserDigest=record.Digest
   raw,_:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
   identityKey=sessionLookupKey(raw)
  }else if key,guest,ok:=s.guestSession(r);ok {
   observation.GuestKey=key
   observation.GuestUser=guest.User
   identityKey=key
  }else{http.Error(w,"session required",401);return}

  c.mu.Lock()
  defer c.mu.Unlock()
  now:=c.now()
  for key,entry:=range c.launches {if now.Sub(entry.Seen)>time.Hour {delete(c.launches,key)}}
  if c.launches==nil {c.launches=make(map[string]launchObservation)}
  if len(c.launches)>4998 {http.Error(w,"launch observation capacity reached",503);return}
  observation.Seen=now
  c.launches[request.Service+":"+identityKey]=observation
  c.launches["homepage:"+identityKey]=launchObservation{Seen:now,Service:"homepage",UserDigest:observation.UserDigest,GuestKey:observation.GuestKey,GuestUser:observation.GuestUser}
  w.WriteHeader(http.StatusNoContent)
 }
}

// launchUsersSnapshot exposes last-seen events only for identities still valid.
// Active-session counts remain exclusively sourced from application heartbeats.
func (c *activityCollector) launchUsersSnapshot(s *sessionService) map[string][]activityUserSummary {
 c.mu.Lock()
 now:=c.now()
 events:=make([]launchObservation,0,len(c.launches))
 for key,entry:=range c.launches {
  if age:=now.Sub(entry.Seen);age<0 || age>time.Hour {delete(c.launches,key);continue}
  events=append(events,entry)
 }
 c.mu.Unlock()
 s.mu.Lock()
 guests:=make(map[string]string,len(s.guests))
 for key,guest:=range s.guests {if s.now().Before(guest.Expires) {guests[key]=guest.User}}
 s.mu.Unlock()
 byService:=make(map[string]map[string]activityUserSummary)
 for _,entry:=range events {
  user:=""
  if entry.GuestKey!="" {
   if guests[entry.GuestKey]!=entry.GuestUser {continue}
   user=entry.GuestUser
  }else if entry.UserDigest!="" {
   name,_,err:=s.store.AuthenticateDigest(entry.UserDigest)
   if err!=nil {continue}
   user=name
  }
  if user=="" {continue}
  if byService[entry.Service]==nil {byService[entry.Service]=make(map[string]activityUserSummary)}
  prev:=byService[entry.Service][user]
  ts:=entry.Seen.UTC().Format(time.RFC3339)
  if ts>prev.LastSeen {byService[entry.Service][user]=activityUserSummary{User:user,LastSeen:ts}}
 }
 result:=make(map[string][]activityUserSummary)
 for service,entries:=range byService {for _,entry:=range entries {result[service]=append(result[service],entry)}}
 return result
}
