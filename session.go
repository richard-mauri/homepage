package main

import (
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "strings"
 "sync"
 "time"

 "themauris.org/homepage/internal/suitestore"
)

const sessionCookieName = "mauri_suite_session"
const sessionLifetime = 12*time.Hour

type sessionRecord struct { User string; Role string; Expires time.Time }
type sessionService struct {
 mu sync.Mutex
 sessions map[string]sessionRecord
 store *suitestore.Store
 now func() time.Time
}

func newSessionService(store *suitestore.Store) *sessionService {
 return &sessionService{sessions:make(map[string]sessionRecord),store:store,now:time.Now}
}
func (s *sessionService) cookie(r *http.Request, value string, maxAge int) *http.Cookie {
 cookie:=&http.Cookie{Name:sessionCookieName,Value:value,Path:"/",Secure:true,HttpOnly:true,SameSite:http.SameSiteLaxMode,MaxAge:maxAge}
 // Set the parent domain only for the configured public suite hosts.
 host:=r.Host
 if strings.HasSuffix(host,".themauris.org") && !strings.Contains(host,":") {cookie.Domain="themauris.org"}
 return cookie
}
func validSessionOrigin(r *http.Request) bool {
 origin:=r.Header.Get("Origin")
 if origin=="" {return false}
 // Session mutation must originate on the same host; sister subdomains are
 // intentionally not trusted to create or revoke sessions.
 return origin=="https://"+r.Host || (strings.HasPrefix(r.Host,"127.0.0.1:") && origin=="http://"+r.Host)
}
func (s *sessionService) handler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method==http.MethodPost && r.URL.Path=="/api/session" {
  if !validSessionOrigin(r){http.Error(w,"invalid origin",http.StatusForbidden);return}
  if r.ContentLength>1024 {http.Error(w,"request too large",413);return}
  token,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
  if !ok {http.Error(w,"bearer token required",401);return}
  user,role,err:=s.store.Authenticate(token)
  if err!=nil{http.Error(w,"authentication failed",401);return}
  var nonce [32]byte
  if _,err=rand.Read(nonce[:]);err!=nil {http.Error(w,"session unavailable",500);return}
  id:=hex.EncodeToString(nonce[:])
  expires:=s.now().Add(sessionLifetime)
  s.mu.Lock()
  for k,v:=range s.sessions {if !s.now().Before(v.Expires){delete(s.sessions,k)}}
  s.sessions[id]=sessionRecord{User:user,Role:role,Expires:expires}
  s.mu.Unlock()
  http.SetCookie(w,s.cookie(r,id,int(sessionLifetime.Seconds())))
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":user,"role":role,"expiresAt":expires.UTC().Format(time.RFC3339)})
  return
 }
 if r.Method==http.MethodGet && r.URL.Path=="/api/session" {
  c,err:=r.Cookie(sessionCookieName)
  if err!=nil {http.Error(w,"not authenticated",401);return}
  s.mu.Lock();v,found:=s.sessions[c.Value];s.mu.Unlock()
  if !found || !s.now().Before(v.Expires){http.Error(w,"session expired",401);return}
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":v.User,"role":v.Role,"expiresAt":v.Expires.UTC().Format(time.RFC3339)})
  return
 }
 if r.Method==http.MethodDelete && r.URL.Path=="/api/session" {
  if !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
  if c,err:=r.Cookie(sessionCookieName);err==nil {
   s.mu.Lock();delete(s.sessions,c.Value);s.mu.Unlock()
  }
  http.SetCookie(w,s.cookie(r,"",-1))
  w.WriteHeader(http.StatusNoContent)
  return
 }
 w.Header().Set("Allow","GET, POST, DELETE")
 http.Error(w,"method not allowed",405)
}
