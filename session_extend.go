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

// extendNewsGoSessionHandler renews the parent Homepage login and rotates its
// NewsGo bearer in one durable transaction. Only a currently valid, parent-
// bound bearer can authorize the operation; neither cookies nor credentials
// supplied from a foreign origin are accepted.
func (s *sessionService) extendNewsGoSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Vary", "Origin")
	const origin = "https://newsgo.themauris.org"
	if r.Method == http.MethodOptions {
		if r.Header.Get("Origin") != origin ||
			r.Header.Get("Access-Control-Request-Method") != http.MethodPost ||
			strings.ToLower(strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))) != "authorization" {
			http.Error(w, "preflight not allowed", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "300")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Origin") != origin {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	// Credential lookup also verifies the current approved-token registry;
	// revocation and role changes take effect before any renewal.
	user, role, valid := s.authenticateHandoff(r)
	if !valid {
		http.Error(w, "session invalid or expired", http.StatusUnauthorized)
		return
	}
	old, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		http.Error(w, "bearer required", http.StatusUnauthorized)
		return
	}
	hash := sha256.Sum256([]byte(old))
	oldKey := hex.EncodeToString(hash[:])
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		http.Error(w, "renewal unavailable", http.StatusInternalServerError)
		return
	}
	replacement := hex.EncodeToString(nonce[:])
	newHash := sha256.Sum256([]byte(replacement))
	s.mu.Lock()
	if s.persistenceError != nil {
		s.mu.Unlock()
		http.Error(w, "session storage unavailable", http.StatusServiceUnavailable)
		return
	}
	handoff, found := s.handoffs[oldKey]
	now := s.now()
	if !found || !now.Before(handoff.Expires) || handoff.ParentSession == "" {
		s.mu.Unlock()
		http.Error(w, "parent-bound session required", http.StatusUnauthorized)
		return
	}
	parent, found := s.sessions[handoff.ParentSession]
	if !found || !now.Before(parent.Expires) || parent.Digest != handoff.Digest {
		s.mu.Unlock()
		http.Error(w, "parent session expired", http.StatusUnauthorized)
		return
	}
	beforeSessions, beforeHandoffs := copySessionRecords(s.sessions), copySessionRecords(s.handoffs)
	expiry := now.Add(sessionLifetime)
	parent.Expires = expiry
	s.sessions[handoff.ParentSession] = parent
	delete(s.handoffs, oldKey)
	s.handoffs[hex.EncodeToString(newHash[:])] = sessionRecord{
		Digest: handoff.Digest, ParentSession: handoff.ParentSession, Expires: expiry,
	}
	if err := s.commitLocked(beforeSessions, beforeHandoffs); err != nil {
		s.mu.Unlock()
		http.Error(w, "session storage unavailable", http.StatusServiceUnavailable)
		return
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": replacement, "expiresAt": expiry.UTC().Format(time.RFC3339),
		"user": user, "role": role,
	})
}
