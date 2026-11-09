package main

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "net/http"
)

// newsGoRecoveryHandler is a top-level, same-browser navigation bridge.
// Only Homepage can read its host-only session cookie. The destination is
// fixed: an untrusted caller cannot supply a return URL or read the bearer.
// Neither a credential nor a Homepage session ID leaves Homepage.
func (s *sessionService) newsGoRecoveryHandler(w http.ResponseWriter, r *http.Request) {
 w.Header().Set("Cache-Control", "no-store")
 w.Header().Set("Referrer-Policy", "no-referrer")
 w.Header().Set("X-Content-Type-Options", "nosniff")
 if r.Method != http.MethodGet {
  w.Header().Set("Allow", "GET")
  http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
  return
 }
 const destination = "https://newsgo.themauris.org/"
 const local = destination + "#suite_mode=local"
 // A missing, expired, or revoked Homepage login is *not* renewed. NewsGo
 // still opens in browser-local mode with subscriptions/preferences intact.
 user, _, valid := s.authenticateSession(r)
 if !valid || user == "" {
  if s.persistenceError != nil {
   http.Error(w, "Homepage session storage unavailable", http.StatusServiceUnavailable)
   return
  }
  http.Redirect(w, r, local, http.StatusFound)
  return
 }
 cookie, err := r.Cookie(sessionCookieName)
 if err != nil {
  http.Redirect(w, r, local, http.StatusFound)
  return
 }
 key := sessionLookupKey(cookie.Value)
 var nonce [32]byte
 if _,err := rand.Read(nonce[:]); err != nil {
  http.Error(w, "handoff unavailable", http.StatusInternalServerError)
  return
 }
 token := hex.EncodeToString(nonce[:])
 hash := sha256.Sum256([]byte(token))
 s.mu.Lock()
 if s.persistenceError != nil {
  s.mu.Unlock()
  http.Error(w,"Homepage session storage unavailable",http.StatusServiceUnavailable)
  return
 }
 parent, exists := s.sessions[key]
 now := s.now()
 if !exists || !now.Before(parent.Expires) {
  s.mu.Unlock()
  http.Redirect(w, r, local, http.StatusFound)
  return
 }
 // The credential was authenticated before taking the lock. A revoked
 // credential still cannot authorize the bearer on later profile/storage
 // calls; this does not create a new login.
 beforeSessions,beforeHandoffs := copySessionRecords(s.sessions),copySessionRecords(s.handoffs)
 for k,v := range s.handoffs {
  if !now.Before(v.Expires) {delete(s.handoffs,k)}
 }
 expiry := now.Add(s.handoffLifetime)
 if parent.Expires.Before(expiry) {expiry = parent.Expires}
 s.handoffs[hex.EncodeToString(hash[:])] = sessionRecord{
  Digest: parent.Digest, Expires: expiry, ParentSession:key,
 }
 if err := s.commitLocked(beforeSessions,beforeHandoffs); err != nil {
  s.mu.Unlock()
  http.Error(w,"Homepage session storage unavailable",http.StatusServiceUnavailable)
  return
 }
 s.mu.Unlock()
 // Fragments are never sent in HTTP requests, and no token enters a query.
 http.Redirect(w, r, destination+"#suite_token="+token, http.StatusFound)
}
