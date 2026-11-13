package main

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "strings"
 "time"
)

const guestCookieName="mauri_guest_session"
const guestLifetime=time.Hour
const guestCapacity=5000

type guestRecord struct {
 User string
 Expires time.Time
}
type guestHandoffRecord struct {
 Parent string
 Expires time.Time
}

// All guest records are process-local. These maps must never be included in
// durableSessionState, suite storage, or on-disk activity history.
func (s *sessionService) cleanupGuestsLocked() {
 now:=s.now()
 for key,g:=range s.guests {if !now.Before(g.Expires){delete(s.guests,key)}}
 for key,h:=range s.guestHandoffs {
  g,ok:=s.guests[h.Parent]
  if !ok || !now.Before(h.Expires) || !now.Before(g.Expires) {delete(s.guestHandoffs,key)}
 }
}
func (s *sessionService) runGuestGC() {
 ticker:=time.NewTicker(5*time.Minute)
 defer ticker.Stop()
 for range ticker.C {
  s.mu.Lock();s.cleanupGuestsLocked();s.mu.Unlock()
 }
}
func (s *sessionService) guestCookie(value string,maxAge int)*http.Cookie {
 return &http.Cookie{Name:guestCookieName,Value:value,Path:"/",Secure:true,HttpOnly:true,SameSite:http.SameSiteLaxMode,MaxAge:maxAge}
}
func (s *sessionService) guestSession(r *http.Request)(string,guestRecord,bool) {
 var token string
 if id,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ");ok {token=id
 } else if c,err:=r.Cookie(guestCookieName);err==nil {token=c.Value}
 if len(token)!=64 || strings.Trim(token,"0123456789abcdef")!="" {return "",guestRecord{},false}
 key:=sessionLookupKey(token)
 s.mu.Lock()
 s.cleanupGuestsLocked()
 guest,ok:=s.guests[key]
 s.mu.Unlock()
 return key,guest,ok
}
func (s *sessionService) guestHandoff(r *http.Request)(string,guestRecord,bool) {
 bearer,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 if !ok || len(bearer)!=64 || strings.Trim(bearer,"0123456789abcdef")!="" {return "",guestRecord{},false}
 key:=sessionLookupKey(bearer)
 s.mu.Lock()
 s.cleanupGuestsLocked()
 handoff,ok:=s.guestHandoffs[key]
 guest,found:=s.guests[handoff.Parent]
 s.mu.Unlock()
 if !ok || !found {return "",guestRecord{},false}
 return handoff.Parent,guest,true
}
func (s *sessionService) mintGuestHandoff(parent string)(string,time.Time,bool) {
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{return "",time.Time{},false}
 token:=hex.EncodeToString(nonce[:])
 s.mu.Lock()
 defer s.mu.Unlock()
 s.cleanupGuestsLocked()
 guest,ok:=s.guests[parent]
 if !ok{return "",time.Time{},false}
 s.guestHandoffs[sessionLookupKey(token)]=guestHandoffRecord{Parent:parent,Expires:guest.Expires}
 return token,guest.Expires,true
}
// POST is a voluntary, same-origin guest sign-in. It never writes guest
// credentials, records or settings to disk. A live guest cookie is reused,
// never renewed or extended.
func (s *sessionService) guestSessionHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
 if !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
 if _,_,ok:=s.authenticateSession(r);ok {http.Error(w,"sign out of your named account before choosing guest",409);return}
 if _,guest,ok:=s.guestSession(r);ok {
  c,err:=r.Cookie(guestCookieName)
  if err==nil {
   w.Header().Set("Content-Type","application/json")
   _=json.NewEncoder(w).Encode(map[string]string{"user":guest.User,"role":"guest","sessionToken":c.Value,"expiresAt":guest.Expires.UTC().Format(time.RFC3339)})
   return
  }
 }
 var nonce [32]byte
 var idBytes [6]byte
 if _,err:=rand.Read(nonce[:]);err!=nil {http.Error(w,"guest unavailable",500);return}
 if _,err:=rand.Read(idBytes[:]);err!=nil {http.Error(w,"guest unavailable",500);return}
 token:=hex.EncodeToString(nonce[:])
 expires:=s.now().Add(guestLifetime)
 user:="guest-"+hex.EncodeToString(idBytes[:])
 s.mu.Lock()
 s.cleanupGuestsLocked()
 if len(s.guests)>=guestCapacity {s.mu.Unlock();http.Error(w,"guest capacity reached",503);return}
 s.guests[sessionLookupKey(token)]=guestRecord{User:user,Expires:expires}
 s.mu.Unlock()
 http.SetCookie(w,s.guestCookie(token,int(guestLifetime.Seconds())))
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"user":user,"role":"guest","sessionToken":token,"expiresAt":expires.UTC().Format(time.RFC3339)})
}
// revokeGuestSession explicitly disconnects a guest and all its child tokens.
func (s *sessionService) revokeGuestSession(w http.ResponseWriter,r *http.Request) {
 key,_,ok:=s.guestSession(r)
 if ok {
  s.mu.Lock()
  delete(s.guests,key)
  for id,h:=range s.guestHandoffs {if h.Parent==key {delete(s.guestHandoffs,id)}}
  s.mu.Unlock()
 }
 if _,err:=r.Cookie(guestCookieName);err==nil {http.SetCookie(w,s.guestCookie("",-1))}
}
// Caller already requires a trusted profile origin and normal bearer shape.
func (s *sessionService) guestProfileTags(origin,parent string)(string,string) {
 sid:=sha256.Sum256([]byte("activity-session:"+origin+":"+parent))
 uid:=sha256.Sum256([]byte("activity-user:"+origin+":"+parent))
 return hex.EncodeToString(sid[:]),hex.EncodeToString(uid[:])
}
