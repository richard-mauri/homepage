package main

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "net/http"
 "strings"
 "sync"
 "time"

 "themauris.org/homepage/internal/suitestore"
)

const sessionCookieName = "mauri_home_session"
const legacySessionCookieName = "mauri_suite_session"
const sessionLifetime = 12*time.Hour
const defaultHandoffLifetime = sessionLifetime

type sessionRecord struct { Digest string; Expires time.Time; ParentSession string }
type durableSessionState struct {
 Version int `json:"version"`
 Sessions map[string]sessionRecord `json:"sessions"`
 Handoffs map[string]sessionRecord `json:"handoffs"`
}

type sessionService struct {
 persistencePath string
 persistenceError error
 mu sync.Mutex
 sessions map[string]sessionRecord
 handoffs map[string]sessionRecord
 guests map[string]guestRecord
 guestHandoffs map[string]guestHandoffRecord
 store *suitestore.Store
 now func() time.Time
 handoffLifetime time.Duration
}

func newSessionService(store *suitestore.Store) *sessionService {
 s:=&sessionService{sessions:make(map[string]sessionRecord),handoffs:make(map[string]sessionRecord),guests:make(map[string]guestRecord),guestHandoffs:make(map[string]guestHandoffRecord),store:store,now:time.Now,handoffLifetime:defaultHandoffLifetime}
 if store!=nil {
  s.persistencePath=store.SessionStatePath()
  s.persistenceError=s.loadDurableState()
 }
 return s
}

// Never persist an actual browser session ID. Only its SHA-256 digest is
// used as the server-side lookup key and written into the protected state.
func sessionLookupKey(id string) string {
 sum:=sha256.Sum256([]byte(id))
 return hex.EncodeToString(sum[:])
}
func (s *sessionService) loadDurableState() error {
 b,err:=os.ReadFile(s.persistencePath)
 if errors.Is(err,os.ErrNotExist){return nil}
 if err!=nil{return err}
 var state durableSessionState
 if err=json.Unmarshal(b,&state);err!=nil{return err}
 if state.Version!=1 || state.Sessions==nil || state.Handoffs==nil {
  return errors.New("unsupported or incomplete durable session state")
 }
 for key,rec:=range state.Sessions {
  if len(key)!=64 || len(rec.Digest)!=64 || rec.Expires.IsZero() {
   return errors.New("invalid durable session record")
  }
 }
 for key,rec:=range state.Handoffs {
  if len(key)!=64 || len(rec.Digest)!=64 || rec.Expires.IsZero() {
   return errors.New("invalid durable handoff record")
  }
 }
 s.sessions=state.Sessions
 s.handoffs=state.Handoffs
 return nil
}

// Called under mu. Write a complete replacement atomically, with restrictive
// permissions, so failures cannot leave a partially written session database.
func (s *sessionService) saveDurableStateLocked() error {
 if s.persistencePath=="" {return nil}
 state:=durableSessionState{Version:1,Sessions:s.sessions,Handoffs:s.handoffs}
 b,err:=json.Marshal(state);if err!=nil{return err}
 dir:=filepath.Dir(s.persistencePath)
 if err=os.MkdirAll(dir,0700);err!=nil{return err}
 tmp,err:=os.CreateTemp(dir,".homepage-auth-*.tmp");if err!=nil{return err}
 defer os.Remove(tmp.Name())
 if err=tmp.Chmod(0600);err!=nil{tmp.Close();return err}
 if _,err=tmp.Write(b);err!=nil{tmp.Close();return err}
 if err=tmp.Sync();err!=nil{tmp.Close();return err}
 if err=tmp.Close();err!=nil{return err}
 if err=os.Rename(tmp.Name(),s.persistencePath);err!=nil{return err}
 // Persist the directory entry as well as the file contents.
 d,err:=os.Open(dir);if err!=nil{return err}
 defer d.Close()
 return d.Sync()
}
func copySessionRecords(src map[string]sessionRecord) map[string]sessionRecord {
 dst:=make(map[string]sessionRecord,len(src))
 for key,record:=range src {dst[key]=record}
 return dst
}
// Called under mu. Revert in-memory changes if the write fails.
func (s *sessionService) commitLocked(beforeSessions,beforeHandoffs map[string]sessionRecord) error {
 if err:=s.saveDurableStateLocked();err!=nil {
  s.sessions=beforeSessions;s.handoffs=beforeHandoffs
  return err
 }
 return nil
}

func (s *sessionService) cookie(r *http.Request, value string, maxAge int) *http.Cookie {
 cookie:=&http.Cookie{Name:sessionCookieName,Value:value,Path:"/",Secure:true,HttpOnly:true,SameSite:http.SameSiteLaxMode,MaxAge:maxAge}
 // Host-only cookie: Homepage's own UI may use it, but suite applications
 // never receive or authenticate through this cookie.
 return cookie
}
// Expire the former parent-domain cookie on sign-in and sign-out. Safari
// can otherwise retain both cookies with the same name and send either first.
func clearLegacyDomainCookie(w http.ResponseWriter,r *http.Request) {
 if r.Host!="home.themauris.org" {return}
 http.SetCookie(w,&http.Cookie{Name:legacySessionCookieName,Value:"",Path:"/",Domain:"themauris.org",Secure:true,HttpOnly:true,SameSite:http.SameSiteLaxMode,MaxAge:-1})
 // Also expire the old host-only variant; it must not coexist with
 // Homepage's new, distinctly named login cookie.
 http.SetCookie(w,&http.Cookie{Name:legacySessionCookieName,Value:"",Path:"/",Secure:true,HttpOnly:true,SameSite:http.SameSiteLaxMode,MaxAge:-1})
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
 case "https://home.themauris.org", "https://stocks.themauris.org", "https://newsgo.themauris.org", "https://weather.themauris.org":
  return true
 default:
  return false
 }
}
// ScraperGo can validate a parent-bound handoff without gaining account storage rights.
func allowedProfileOrigin(origin string) bool {
 return allowedStorageOrigin(origin) || activityAliasOrigin(origin)
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
 if s.persistenceError!=nil{return "","",false}
 c,err:=r.Cookie(sessionCookieName);if err!=nil{return "","",false}
 s.mu.Lock();v,found:=s.sessions[sessionLookupKey(c.Value)];s.mu.Unlock()
 if !found || !s.now().Before(v.Expires){return "","",false}
 user,role,err:=s.store.AuthenticateDigest(v.Digest)
 return user,role,err==nil
}
// A Homepage-only session is an independent 12-hour opaque credential,
// explicitly sent by the Homepage browser. App bearers never authorize it.
func (s *sessionService) homepageSession(r *http.Request) (sessionRecord,string,string,bool) {
 if s.persistenceError!=nil{return sessionRecord{},"","",false}
 id,has:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
 if !has || len(id)!=64{return sessionRecord{},"","",false}
 if _,err:=hex.DecodeString(id);err!=nil{return sessionRecord{},"","",false}
 s.mu.Lock();record,exists:=s.sessions[sessionLookupKey(id)];s.mu.Unlock()
 if !exists || !s.now().Before(record.Expires){return sessionRecord{},"","",false}
 user,role,err:=s.store.AuthenticateDigest(record.Digest)
 return record,user,role,err==nil
}
// authenticateHandoff checks the parent-session-bound bearer and revalidates
// its originating credential digest so role changes/revocations take effect.
func (s *sessionService) authenticateHandoff(r *http.Request) (string,string,bool) {
 if _,guest,ok:=s.guestHandoff(r);ok{return guest.User,"guest",true}
 if s.persistenceError!=nil{return "","",false}
 token,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 if !ok || len(token)!=64{return "","",false}
 digest:=sha256.Sum256([]byte(token))
 key:=hex.EncodeToString(digest[:])
 s.mu.Lock();record,found:=s.handoffs[key]
 if found && !s.now().Before(record.Expires){delete(s.handoffs,key);found=false}
 if found && record.ParentSession!="" {
  parent,exists:=s.sessions[record.ParentSession]
  if !exists || !s.now().Before(parent.Expires) || parent.Digest!=record.Digest {
   delete(s.handoffs,key);found=false
  }
 }
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
 // A Homepage guest uses an in-memory child bearer; never write it to the
 // durable named-session/handoff maps.
 if strings.HasPrefix(r.Header.Get("Authorization"),"Session ") {
  if parent,guest,ok:=s.guestSession(r);ok {
   token,expires,issued:=s.mintGuestHandoff(parent)
   if !issued {http.Error(w,"guest expired",401);return}
   w.Header().Set("Content-Type","application/json")
   _=json.NewEncoder(w).Encode(map[string]string{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339),"user":guest.User,"role":"guest"})
   return
  }
 } else if r.Header.Get("Authorization")=="" {
  if parent,guest,ok:=s.guestSession(r);ok {
   token,expires,issued:=s.mintGuestHandoff(parent)
   if !issued {http.Error(w,"guest expired",401);return}
   w.Header().Set("Content-Type","application/json")
   _=json.NewEncoder(w).Encode(map[string]string{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339),"user":guest.User,"role":"guest"})
   return
  }
 }
 // When a fresh approved token is provided, it takes precedence over
 // any older Homepage cookie (which might represent a different user).
 approved,hasBearer:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 var user,role string
 var source sessionRecord
 var parentSession string
 if hasBearer {
  var err error
  user,role,err=s.store.Authenticate(approved)
  if err!=nil{http.Error(w,"authentication failed",401);return}
  digest:=sha256.Sum256([]byte(approved))
  source=sessionRecord{Digest:hex.EncodeToString(digest[:])}
  // A credential-authenticated handoff issued immediately after sign-in
  // is also owned by that browser's active matching Homepage session.
  if c,err:=r.Cookie(sessionCookieName);err==nil {
   s.mu.Lock();v,exists:=s.sessions[sessionLookupKey(c.Value)];s.mu.Unlock()
   if exists && v.Digest==source.Digest && s.now().Before(v.Expires){parentSession=sessionLookupKey(c.Value)}
  }
 }else if strings.HasPrefix(r.Header.Get("Authorization"),"Session ") {
  var ok bool
  source,user,role,ok=s.homepageSession(r)
  if !ok{http.Error(w,"Homepage session invalid",401);return}
  rawSession,_:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
  parentSession=sessionLookupKey(rawSession)
 }else{
  var ok bool
  user,role,ok=s.authenticateSession(r)
  if !ok{http.Error(w,"session invalid",401);return}
  c,_:=r.Cookie(sessionCookieName)
  s.mu.Lock();source=s.sessions[sessionLookupKey(c.Value)];s.mu.Unlock()
  parentSession=sessionLookupKey(c.Value)
 }
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{http.Error(w,"token unavailable",500);return}
 token:=hex.EncodeToString(nonce[:])
 hash:=sha256.Sum256([]byte(token))
 expires:=s.now().Add(s.handoffLifetime)
 s.mu.Lock()
 if s.persistenceError!=nil {s.mu.Unlock();http.Error(w,"session storage unavailable",503);return}
 beforeSessions,beforeHandoffs:=copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
 if parentSession!="" {
  parent,exists:=s.sessions[parentSession]
  if !exists || !s.now().Before(parent.Expires) || parent.Digest!=source.Digest {
   s.mu.Unlock();http.Error(w,"Homepage session signed out",401);return
  }
  // A handoff never outlives its twelve-hour Homepage login.
  if parent.Expires.Before(expires){expires=parent.Expires}
 }
 for key,v:=range s.handoffs {if !s.now().Before(v.Expires){delete(s.handoffs,key)}}
 s.handoffs[hex.EncodeToString(hash[:])]=sessionRecord{Digest:source.Digest,Expires:expires,ParentSession:parentSession}
 if err:=s.commitLocked(beforeSessions,beforeHandoffs);err!=nil {
  s.mu.Unlock();http.Error(w,"session storage unavailable",503);return
 }
 s.mu.Unlock()
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]any{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339),"user":user,"role":role})
}
// renewHandoffHandler rotates an unexpired app bearer without a browser cookie.
// The new token supersedes the old one; an expired token cannot be renewed.
func (s *sessionService) renewHandoffHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Vary","Origin")
 origin:=r.Header.Get("Origin")
 if r.Method==http.MethodOptions {
  if !allowedStorageOrigin(origin) || origin=="https://home.themauris.org" || r.Header.Get("Access-Control-Request-Method")!=http.MethodPost ||
   strings.ToLower(strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers")))!="authorization" {
   http.Error(w,"preflight not allowed",403);return
  }
  w.Header().Set("Access-Control-Allow-Origin",origin)
  w.Header().Set("Access-Control-Allow-Methods","POST")
  w.Header().Set("Access-Control-Allow-Headers","Authorization")
  w.Header().Set("Access-Control-Max-Age","300")
  w.WriteHeader(http.StatusNoContent);return
 }
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST, OPTIONS");http.Error(w,"method not allowed",405);return}
 if origin!="https://newsgo.themauris.org" && origin!="https://stocks.themauris.org" {http.Error(w,"origin not allowed",403);return}
 w.Header().Set("Access-Control-Allow-Origin",origin)
 user,role,ok:=s.authenticateHandoff(r)
 if !ok{http.Error(w,"account token invalid or expired",401);return}
 old,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
 if !ok{http.Error(w,"token required",401);return}
 oldHash:=sha256.Sum256([]byte(old))
 oldKey:=hex.EncodeToString(oldHash[:])
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{http.Error(w,"renewal unavailable",500);return}
 newToken:=hex.EncodeToString(nonce[:])
 newHash:=sha256.Sum256([]byte(newToken))
 expires:=s.now().Add(s.handoffLifetime)
 s.mu.Lock()
 if s.persistenceError!=nil {s.mu.Unlock();http.Error(w,"session storage unavailable",503);return}
 beforeSessions,beforeHandoffs:=copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
 record,found:=s.handoffs[oldKey]
 if !found || !s.now().Before(record.Expires){s.mu.Unlock();http.Error(w,"token expired",401);return}
 if record.ParentSession!="" {
  parent,exists:=s.sessions[record.ParentSession]
  if !exists || !s.now().Before(parent.Expires) || parent.Digest!=record.Digest {
   delete(s.handoffs,oldKey);s.mu.Unlock();http.Error(w,"Homepage session signed out",401);return
  }
  // Compatibility-only renewal; never extend access beyond parent expiry.
  if parent.Expires.Before(expires){expires=parent.Expires}
 }
 delete(s.handoffs,oldKey)
 s.handoffs[hex.EncodeToString(newHash[:])]=sessionRecord{Digest:record.Digest,Expires:expires,ParentSession:record.ParentSession}
 if err:=s.commitLocked(beforeSessions,beforeHandoffs);err!=nil {
  s.mu.Unlock();http.Error(w,"session storage unavailable",503);return
 }
 s.mu.Unlock()
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"token":newToken,"expiresAt":expires.UTC().Format(time.RFC3339),"user":user,"role":role})
}
// recoverScraperHandoffHandler silently replaces a ScraperGo child bearer
// using only the existing, unexpired, host-only Homepage cookie. It never
// renews the parent login, accepts an approved personal token, or grants
// ScraperGo account-storage access.
func (s *sessionService) recoverScraperHandoffHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Vary","Origin")
 origin:=r.Header.Get("Origin")
 if origin!="https://scrapergo.themauris.org" {http.Error(w,"origin not allowed",403);return}
 w.Header().Set("Access-Control-Allow-Origin",origin)
 w.Header().Set("Access-Control-Allow-Credentials","true")
 if r.Method==http.MethodOptions {
  if r.Header.Get("Access-Control-Request-Method")!=http.MethodPost ||
   strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))!="" {
   http.Error(w,"preflight not allowed",403);return
  }
  w.Header().Set("Access-Control-Allow-Methods","POST")
  w.Header().Set("Access-Control-Max-Age","300")
  w.WriteHeader(http.StatusNoContent);return
 }
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST, OPTIONS");http.Error(w,"method not allowed",405);return}
 if s.persistenceError!=nil {http.Error(w,"session storage unavailable",503);return}
 // A temporary guest can silently obtain a child bearer using its valid
 // host-only Homepage guest cookie. No browser navigation is necessary.
 if parent,_,ok:=s.guestSession(r);ok {
  token,expires,issued:=s.mintGuestHandoff(parent)
  if !issued {http.Error(w,"guest expired",401);return}
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]string{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339)})
  return
 }
 c,err:=r.Cookie(sessionCookieName)
 if err!=nil || len(c.Value)!=64 {http.Error(w,"Homepage sign-in required",401);return}
 key:=sessionLookupKey(c.Value)
 s.mu.Lock()
 parent,found:=s.sessions[key]
 s.mu.Unlock()
 if !found || !s.now().Before(parent.Expires) {http.Error(w,"Homepage sign-in expired",401);return}
 if _,_,err=s.store.AuthenticateDigest(parent.Digest);err!=nil {http.Error(w,"Homepage sign-in revoked",401);return}
 var nonce [32]byte
 if _,err=rand.Read(nonce[:]);err!=nil {http.Error(w,"handoff unavailable",500);return}
 token:=hex.EncodeToString(nonce[:])
 hash:=sha256.Sum256([]byte(token))
 expires:=s.now().Add(s.handoffLifetime)
 s.mu.Lock()
 if s.persistenceError!=nil {s.mu.Unlock();http.Error(w,"session storage unavailable",503);return}
 // Recheck the parent under the write lock, including sign-out races.
 current,valid:=s.sessions[key]
 if !valid || !s.now().Before(current.Expires) || current.Digest!=parent.Digest {
  s.mu.Unlock();http.Error(w,"Homepage sign-in expired",401);return
 }
 if current.Expires.Before(expires) {expires=current.Expires}
 beforeSessions,beforeHandoffs:=copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
 s.handoffs[hex.EncodeToString(hash[:])]=sessionRecord{Digest:current.Digest,Expires:expires,ParentSession:key}
 if err=s.commitLocked(beforeSessions,beforeHandoffs);err!=nil {
  s.mu.Unlock();http.Error(w,"session storage unavailable",503);return
 }
 s.mu.Unlock()
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"token":token,"expiresAt":expires.UTC().Format(time.RFC3339)})
}

// storageHandler uses server-derived identity; cookie-authenticated writes
// require an exact same-origin Origin header to prevent cross-site mutations.
func (s *sessionService) storageHandler(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if _,guest,ok:=s.guestSession(r);ok && guest.User!="" {
   http.Error(w,"guest settings persistence forbidden",403);return
  }
  if _,guest,ok:=s.guestHandoff(r);ok && guest.User!="" {
   http.Error(w,"guest settings persistence forbidden",403);return
  }
  if user,_,ok:=s.authenticateHandoff(r);ok {next.ServeHTTP(w,suitestore.WithIdentity(r,user));return}
  // Suite browsers must present the session-bound handoff. Homepage's login
  // cookie is not an alternative identity for cross-origin application APIs.
  origin:=r.Header.Get("Origin")
  if origin!="" && origin!="https://home.themauris.org" {
   http.Error(w,"account handoff token required",401);return
  }
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
func activityAliasOrigin(origin string) bool {
 return origin=="https://scrapergo.themauris.org" || origin=="https://newsgo.themauris.org"
}

func (s *sessionService) profileHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Vary","Origin")
 origin:=r.Header.Get("Origin")
 if r.Method==http.MethodOptions {
  // An Authorization header on a cross-origin GET requires a browser
  // preflight. Reject untrusted origins, methods, or request headers.
  if !allowedProfileOrigin(origin) || r.Header.Get("Access-Control-Request-Method")!=http.MethodGet {
   http.Error(w,"preflight not allowed",403);return
  }
  headers:=strings.ToLower(strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers")))
  if headers!="" && headers!="authorization" {
   http.Error(w,"preflight headers not allowed",403);return
  }
  w.Header().Set("Access-Control-Allow-Origin",origin)
  w.Header().Set("Access-Control-Allow-Credentials","true")
  w.Header().Set("Access-Control-Allow-Methods","GET")
  w.Header().Set("Access-Control-Allow-Headers","Authorization")
  w.Header().Set("Access-Control-Max-Age","300")
  w.WriteHeader(http.StatusNoContent);return
 }
 if r.Method!=http.MethodGet {w.Header().Set("Allow","GET, OPTIONS");http.Error(w,"method not allowed",405);return}
 if origin!="" && !allowedProfileOrigin(origin){http.Error(w,"origin not allowed",403);return}
 if origin!="" {w.Header().Set("Access-Control-Allow-Origin",origin);w.Header().Set("Access-Control-Allow-Credentials","true")}
 user,role,ok:=s.authenticateHandoff(r)
 // The cookie is only valid for Homepage's own same-origin page, never
 // as fallback for NewsGo or MauriStocks.
 if !ok && (origin=="" || origin=="https://home.themauris.org") {
  user,role,ok=s.authenticateSession(r)
 }
 if !ok {http.Error(w,"account token invalid or expired",401);return}
 w.Header().Set("Content-Type","application/json")
 if role=="guest" {
  parent,guest,valid:=s.guestHandoff(r)
  if !valid {http.Error(w,"guest expired",401);return}
  result:=map[string]string{"user":guest.User,"role":"guest","expiresAt":guest.Expires.UTC().Format(time.RFC3339)}
  if activityAliasOrigin(origin) {
   result["activitySession"],result["activityUser"]=s.guestProfileTags(origin,parent)
  }
  _=json.NewEncoder(w).Encode(result);return
 }
 result:=map[string]string{"user":user,"role":role}
 if bearer,exists:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ");exists {
  sum:=sha256.Sum256([]byte(bearer))
  s.mu.Lock();record,found:=s.handoffs[hex.EncodeToString(sum[:])];s.mu.Unlock()
  if found && s.now().Before(record.Expires){
   result["expiresAt"]=record.Expires.UTC().Format(time.RFC3339)
   if activityAliasOrigin(origin) && record.ParentSession!="" {
    // Privacy-preserving, origin-scoped identifiers; never expose parent
    // session ID, approved credential digest or bearer in activity metrics.
    sessionTag:=sha256.Sum256([]byte("activity-session:"+origin+":"+record.ParentSession))
    userTag:=sha256.Sum256([]byte("activity-user:"+origin+":"+record.Digest))
    result["activitySession"]=hex.EncodeToString(sessionTag[:])
    result["activityUser"]=hex.EncodeToString(userTag[:])
   }
  }
 }
 _=json.NewEncoder(w).Encode(result)
}

func (s *sessionService) handler(w http.ResponseWriter,r *http.Request) {
 if s.persistenceError!=nil {http.Error(w,"session storage unavailable",503);return}
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
  beforeSessions,beforeHandoffs:=copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
  for k,v:=range s.sessions {if !s.now().Before(v.Expires){delete(s.sessions,k)}}
  digest:=sha256.Sum256([]byte(token))
  s.sessions[sessionLookupKey(id)]=sessionRecord{Digest:hex.EncodeToString(digest[:]),Expires:expires}
  if err:=s.commitLocked(beforeSessions,beforeHandoffs);err!=nil {
   s.mu.Unlock();http.Error(w,"session storage unavailable",503);return
  }
  s.mu.Unlock()
  s.revokeGuestSession(w,r)
  http.SetCookie(w,s.cookie(r,id,int(sessionLifetime.Seconds())))
  clearLegacyDomainCookie(w,r)
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":user,"role":role,"expiresAt":expires.UTC().Format(time.RFC3339),"sessionToken":id})
  return
 }
 if r.Method==http.MethodGet && r.URL.Path=="/api/session" {
  var user,role,id string
  var v sessionRecord
  var ok bool
  if strings.HasPrefix(r.Header.Get("Authorization"),"Session ") {
   v,user,role,ok=s.homepageSession(r)
   id,_=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
  }else{
   user,role,ok=s.authenticateSession(r)
   if ok{
    c,_:=r.Cookie(sessionCookieName)
    id=c.Value
    s.mu.Lock();v=s.sessions[sessionLookupKey(id)];s.mu.Unlock()
   }
  }
  if !ok{
   _,guest,valid:=s.guestSession(r)
   if !valid {http.Error(w,"session invalid",401);return}
   id,_=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
   if id=="" {if c,err:=r.Cookie(guestCookieName);err==nil {id=c.Value}}
   w.Header().Set("Content-Type","application/json")
   _=json.NewEncoder(w).Encode(map[string]any{"user":guest.User,"role":"guest","expiresAt":guest.Expires.UTC().Format(time.RFC3339),"sessionToken":id})
   return
  }
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]any{"user":user,"role":role,"expiresAt":v.Expires.UTC().Format(time.RFC3339),"sessionToken":id})
  return
 }
 if r.Method==http.MethodDelete && r.URL.Path=="/api/session" {
  if !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
  s.revokeGuestSession(w,r)
  // Revoke handoffs belonging to the signed-out Homepage login, not
  // handoffs belonging to independent sessions on other devices.
  revoked:=make(map[string]bool)
  if id,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ");ok{
   if _,_,_,valid:=s.homepageSession(r);valid {revoked[sessionLookupKey(id)]=true}
  }
  if c,err:=r.Cookie(sessionCookieName);err==nil {revoked[sessionLookupKey(c.Value)]=true}
  s.mu.Lock()
  beforeSessions,beforeHandoffs:=copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
  for id:=range revoked {delete(s.sessions,id)}
  for key,handoff:=range s.handoffs {
   if revoked[handoff.ParentSession] && handoff.ParentSession!="" {delete(s.handoffs,key)}
  }
  if err:=s.commitLocked(beforeSessions,beforeHandoffs);err!=nil {
   s.mu.Unlock();http.Error(w,"session storage unavailable",503);return
  }
  s.mu.Unlock()
  http.SetCookie(w,s.cookie(r,"",-1))
  clearLegacyDomainCookie(w,r)
  w.WriteHeader(http.StatusNoContent)
  return
 }
 w.Header().Set("Allow","GET, POST, DELETE")
 http.Error(w,"method not allowed",405)
}
