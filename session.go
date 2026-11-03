package main

import (
 "crypto/rand"
 "crypto/sha256"
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
const handoffLifetime = time.Hour

type sessionRecord struct { Digest string; Expires time.Time }
type sessionService struct {
 mu sync.Mutex
 sessions map[string]sessionRecord
 handoffs map[string]sessionRecord
 store *suitestore.Store
 now func() time.Time
}

func newSessionService(store *suitestore.Store) *sessionService {
 return &sessionService{sessions:make(map[string]sessionRecord),handoffs:make(map[string]sessionRecord),store:store,now:time.Now}
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
// Only these trusted application origins may call the storage API using
// credentials. Other origins receive no CORS permission.
func allowedStorageOrigin(origin string) bool {
 switch origin {
 case "https://home.themauris.org", "https://stocks.themauris.org", "https://newsgo.themauris.org":
  return true
 default:
  return false
 }
}
func (s *sessionService) suiteStorageHandler(next http.Handler) http.Handler {
 secured:=s.storageHandler(next)
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  origin:=r.Header.Get("Origin")
  w.Header().Add("Vary","Origin")
  if origin!="" && allowedStorageOrigin(origin) {
   w.Header().Set("Access-Control-Allow-Origin",origin)
   w.Header().Set("Access-Control-Allow-Credentials","true")
  }
  if r.Method==http.MethodOptions {
   if !allowedStorageOrigin(origin) || r.Header.Get("Access-Control-Request-Method")!="GET" && r.Header.Get("Access-Control-Request-Method")!="PUT" {
    http.Error(w,"preflight not allowed",403);return
   }
   headers:=strings.ToLower(r.Header.Get("Access-Control-Request-Headers"))
   if headers!="" && headers!="content-type" && headers!="authorization" && headers!="content-type, authorization" && headers!="authorization, content-type"{
    http.Error(w,"preflight headers not allowed",403);return
   }
   w.Header().Set("Access-Control-Allow-Methods","GET, PUT")
   w.Header().Set("Access-Control-Allow-Headers","Content-Type, Authorization")
   w.Header().Set("Access-Control-Max-Age","300")
   w.WriteHeader(http.StatusNoContent);return
  }
  // A supplied Origin must be one of our trusted suite applications.
  // Requests without Origin remain available to non-browser bearer clients.
  if origin!="" && !allowedStorageOrigin(origin){http.Error(w,"origin not allowed",403);return}
  secured.ServeHTTP(w,r)
 })
}
// authenticateSession rechecks the credential registry on each request.
// Removing a token or changing its role takes effect immediately.
func (s *sessionService) authenticateSession(r *http.Request) (string,string,bool) {
 c,err:=r.Cookie(sessionCookieName);if err!=nil{return "","",false}
 s.mu.Lock();v,found:=s.sessions[c.Value];s.mu.Unlock()
 if !found || !s.now().Before(v.Expires){return "","",false}
 user,role,err:=s.store.AuthenticateDigest(v.Digest)
 return user,role,err==nil
}
// authenticateHandoff checks the independent one-hour bearer and revalidates
// its originating credential digest so role changes/revocations take effect.
func (s *sessionService) authenticateHandoff(r *http.Request) (string,string,bool) {
 token,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 if !ok || len(token)!=64{return "","",false}
 digest:=sha256.Sum256([]byte(token))
 key:=hex.EncodeToString(digest[:])
 s.mu.Lock();record,found:=s.handoffs[key]
 if found && !s.now().Before(record.Expires){delete(s.handoffs,key);found=false}
 s.mu.Unlock()
 if !found{return "","",false}
 user,role,err:=s.store.AuthenticateDigest(record.Digest)
 return user,role,err==nil
}
// handoffHandler issues a new browser-transfer bearer, but never the durable
// approved user credential. Link fragments are not transmitted to HTTP servers.
func (s *sessionService) handoffHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
 if !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
 user,role,ok:=s.authenticateSession(r)
 if !ok{http.Error(w,"session invalid",401);return}
 c,_:=r.Cookie(sessionCookieName)
 s.mu.Lock();source,found:=s.sessions[c.Value];s.mu.Unlock()
 if !found{http.Error(w,"session invalid",401);return}
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{http.Error(w,"token unavailable",500);return}
 token:=hex.EncodeToString(nonce[:])
 hash:=sha256.Sum256([]byte(token))
 expires:=s.now().Add(handoffLifetime)
 s.mu.Lock()
 for key,v:=range s.handoffs {if !s.now().Before(v.Expires){delete(s.handoffs,key)}}
 s.handoffs[hex.EncodeToString(hash[:])]=sessionRecord{Digest:source.Digest,Expires:expires}
 s.mu.Unlock()
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]any{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339),"user":user,"role":role})
}
// storageHandler uses server-derived identity; cookie-authenticated writes
// require an exact same-origin Origin header to prevent cross-site mutations.
func (s *sessionService) storageHandler(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if user,_,ok:=s.authenticateHandoff(r);ok {next.ServeHTTP(w,suitestore.WithIdentity(r,user));return}
  if _,err:=r.Cookie(sessionCookieName);err==nil {
   user,_,ok:=s.authenticateSession(r)
   if !ok{http.Error(w,"session invalid",401);return}
   if r.Method!=http.MethodGet && !allowedStorageOrigin(r.Header.Get("Origin")) {http.Error(w,"invalid origin",403);return}
   next.ServeHTTP(w,suitestore.WithIdentity(r,user));return
  }
  next.ServeHTTP(w,r)
 })
}
// profileHandler provides a read-only authenticated identity to trusted suite apps.
// It never creates or revokes sessions and does not expose token material.
func (s *sessionService) profileHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Vary","Origin")
 origin:=r.Header.Get("Origin")
 if origin!="" && allowedStorageOrigin(origin) {
  w.Header().Set("Access-Control-Allow-Origin",origin)
  w.Header().Set("Access-Control-Allow-Credentials","true")
 }
 if r.Method!=http.MethodGet {w.Header().Set("Allow","GET");http.Error(w,"method not allowed",405);return}
 if origin!="" && !allowedStorageOrigin(origin){http.Error(w,"origin not allowed",403);return}
 user,role,ok:=s.authenticateHandoff(r)
 if !ok {user,role,ok=s.authenticateSession(r)}
 if !ok {http.Error(w,"session invalid",401);return}
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"user":user,"role":role})
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
  digest:=sha256.Sum256([]byte(token))
  s.sessions[id]=sessionRecord{Digest:hex.EncodeToString(digest[:]),Expires:expires}
  s.mu.Unlock()
  http.SetCookie(w,s.cookie(r,id,int(sessionLifetime.Seconds())))
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":user,"role":role,"expiresAt":expires.UTC().Format(time.RFC3339)})
  return
 }
 if r.Method==http.MethodGet && r.URL.Path=="/api/session" {
  user,role,ok:=s.authenticateSession(r)
  if !ok{http.Error(w,"session invalid",401);return}
  c,_:=r.Cookie(sessionCookieName)
  s.mu.Lock();v:=s.sessions[c.Value];s.mu.Unlock()
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":user,"role":role,"expiresAt":v.Expires.UTC().Format(time.RFC3339)})
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
