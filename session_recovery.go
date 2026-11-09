package main

import (
 "net/http"
)

// newsGoRecoveryHandler serves a same-origin handoff bridge. Only the
// Homepage origin can access its localStorage session ID or host-only cookie.
// The page issues a POST /api/handoff/token, never a credential-creating GET.
// Its destination is fixed at the trusted NewsGo origin.
func (s *sessionService) newsGoRecoveryHandler(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Referrer-Policy","no-referrer")
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.Header().Set("Content-Security-Policy","default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
 if r.Method!=http.MethodGet {
  w.Header().Set("Allow","GET")
  http.Error(w,"method not allowed",http.StatusMethodNotAllowed)
  return
 }
 if s.persistenceError!=nil {
  http.Error(w,"Homepage session storage unavailable",http.StatusServiceUnavailable)
  return
 }
 page,err:=assets.ReadFile("static/launch-newsgo.html")
 if err!=nil{
  http.Error(w,"Session recovery unavailable",http.StatusInternalServerError)
  return
 }
 w.Header().Set("Content-Type","text/html; charset=utf-8")
 _,_=w.Write(page)
}
