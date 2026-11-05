package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "io"
 "log"
 "net/http"
 "regexp"
 "strings"
 "time"
)

var authTraceEnabled bool

var traceIDPattern=regexp.MustCompile(`^[a-f0-9]{16,32}$`)

func safeTraceID(v string) string {
 if traceIDPattern.MatchString(v) {return v}
 return "-"
}
func traceAuthf(format string,args ...any) {
 if authTraceEnabled {log.Printf("[AUTH-TRACE] "+format,args...)}
}

// Trace only the status, auth transport, and failure classification.
// Never log Authorization values, Cookie values, URL fragments or user data.
type traceResponseWriter struct {
 http.ResponseWriter
 status int
}
func (w *traceResponseWriter) WriteHeader(status int) {
 if w.status!=0{return}
 w.status=status
 w.ResponseWriter.WriteHeader(status)
}
func (w *traceResponseWriter) Write(b []byte)(int,error) {
 if w.status==0{w.status=http.StatusOK}
 return w.ResponseWriter.Write(b)
}
func authTraceHTTP(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if !authTraceEnabled {next.ServeHTTP(w,r);return}
  started:=time.Now()
  tw:=&traceResponseWriter{ResponseWriter:w}
  next.ServeHTTP(tw,r)
  status:=tw.status
  if status==0{status=http.StatusOK}
  kind:="none"
  if strings.HasPrefix(r.Header.Get("Authorization"),"Session "){kind="session"}
  if strings.HasPrefix(r.Header.Get("Authorization"),"Bearer "){kind="bearer"}
  _,cookieErr:=r.Cookie(sessionCookieName)
  reason:="-"
  if status==http.StatusUnauthorized || status==http.StatusForbidden {
   reason=traceAuthFailureReason(r)
  }
  traceAuthf("http endpoint=%s method=%s status=%d auth=%s cookie=%t reason=%s page=%s elapsed_ms=%d",
   r.URL.Path,r.Method,status,kind,cookieErr==nil,reason,safeTraceID(r.Header.Get("X-Homepage-Trace")),time.Since(started).Milliseconds())
 })
}
func traceAuthFailureReason(r *http.Request) string {
 if r.URL.Path!="/api/session" || r.Method!=http.MethodGet {return "request_rejected"}
 // The exact identity of a session is never written to the log.
 id,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Session ")
 if !ok {
  c,err:=r.Cookie(sessionCookieName)
  if err!=nil{return "no_session_header_or_cookie"}
  id=c.Value
 }
 if len(id)!=64 {return "malformed_session"}
 sum:=sha256.Sum256([]byte(id))
 key:=hex.EncodeToString(sum[:])
 // Failure classification is intentionally limited to the server's
 // already-observed request state. Do not reveal credentials or identity.
 if sessionTraceLookup==nil{return "unavailable"}
 return sessionTraceLookup(key)
}

// Set only during Homepage setup; the session handler itself reports the
// underlying status. This read-only callback is used only in trace mode.
var sessionTraceLookup func(string)string

type clientTraceEvent struct {
 Event string `json:"event"`
 Phase string `json:"phase"`
 Status int `json:"status"`
 SessionStored bool `json:"sessionStored"`
 Page string `json:"page"`
}
var allowedTraceEvents=map[string]bool{
 "page_init":true,"page_show":true,"page_focus":true,"page_online":true,
 "refresh_begin":true,"refresh_response":true,"refresh_error":true,
 "ui_signed_in":true,"ui_signed_out":true,"ui_unavailable":true,
 "launch_begin":true,"handoff_response":true,"launch_local":true,"launch_handoff":true,
 "signin_response":true,"signin_error":true,"signout_response":true,
}
var allowedTracePhases=map[string]bool{
 "-":true,"initial":true,"session":true,"cookie_fallback":true,
 "valid":true,"invalid":true,"unavailable":true,"network":true,
 "stocks":true,"newsgo":true,"none":true,"expired":true,"success":true,
 "rejected":true,"storage_error":true,
}
func clientAuthTraceHandler(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if !authTraceEnabled {http.NotFound(w,r);return}
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return}
 if !validSessionOrigin(r){http.Error(w,"invalid origin",403);return}
 r.Body=http.MaxBytesReader(w,r.Body,1024)
 dec:=json.NewDecoder(r.Body)
 dec.DisallowUnknownFields()
 var event clientTraceEvent
 if err:=dec.Decode(&event);err!=nil {http.Error(w,"invalid trace",400);return}
 var extra any
 if err:=dec.Decode(&extra);err!=io.EOF {http.Error(w,"trailing trace",400);return}
 if !allowedTraceEvents[event.Event] || !allowedTracePhases[event.Phase] ||
  event.Status<0 || event.Status>599 || !traceIDPattern.MatchString(event.Page) {
  http.Error(w,"invalid trace event",400);return
 }
 traceAuthf("browser event=%s phase=%s status=%d stored=%t page=%s",
  event.Event,event.Phase,event.Status,event.SessionStored,event.Page)
 w.WriteHeader(http.StatusNoContent)
}

func (s *sessionService) traceSessionLookup(key string)string{
 s.mu.Lock()
 rec,found:=s.sessions[key]
 s.mu.Unlock()
 if !found{return "session_not_found"}
 if !s.now().Before(rec.Expires){return "session_expired"}
 if _,_,err:=s.store.AuthenticateDigest(rec.Digest);err!=nil{return "credential_unapproved"}
 return "session_valid"
}

func setSessionTraceLookup(s *sessionService){
 if authTraceEnabled {
  sessionTraceLookup=s.traceSessionLookup
  traceAuthf("enabled durable_sessions=%t session_duration=%s",s.persistencePath!="",sessionLifetime)
 }else{sessionTraceLookup=nil}
}

