// Package auth defines the shared Mauris Suite access contract.
//
// Homepage remains the sole identity issuer. Consumers must supply a verifier
// that checks the live Homepage bearer/server session. Browser labels, query
// parameters, unverified headers, and cookies from unrelated origins are never
// identities. This helper does NOT authenticate a bearer on its own.
package auth

import (
 "context"
 "errors"
 "net/http"
)

type Role string

const (
 RoleUser Role = "user"
 RoleAdmin Role = "admin"
 RoleGuest Role = "guest"
)

type Identity struct {
 User string
 Role Role
 Verified bool
}

type Permission uint8

const (
 UseApplication Permission = iota
 UsePersonalSettings
)

// ErrVerificationUnavailable means the identity authority cannot currently
// give a trustworthy answer. Authentication errors must never authorize use.
var ErrVerificationUnavailable = errors.New("Homepage verification unavailable")

// Verifier is implemented by the consumer using Homepage's verified session
// profile endpoint. It must NEVER trust caller-provided role/user labels.
type Verifier func(context.Context, *http.Request) (Identity, error)

func Allowed(identity Identity, permission Permission) bool {
 if !identity.Verified || identity.User=="" {return false}
 switch identity.Role {
 case RoleUser,RoleAdmin:
  return permission==UseApplication || permission==UsePersonalSettings
 case RoleGuest:
  return permission==UseApplication
 default:
  return false
 }
}

// Guard is intended for protected application API handlers. It checks the
// verified identity on EVERY request; a one-time front-end check is not a
// server-side security boundary. An anonymous request is 401; a verified guest
// forbidden from personal settings is 403. Outages are 503, never fallback.
func Guard(verify Verifier, permission Permission, next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  if verify==nil || next==nil {
   http.Error(w,"suite authentication unavailable",http.StatusServiceUnavailable)
   return
  }
  identity,err:=verify(r.Context(),r)
  if err!=nil{
   http.Error(w,"Homepage verification unavailable",http.StatusServiceUnavailable)
   return
  }
  if !identity.Verified || identity.User=="" ||
    (identity.Role!=RoleUser && identity.Role!=RoleAdmin && identity.Role!=RoleGuest) {
   http.Error(w,"Homepage sign-in or guest access required",http.StatusUnauthorized)
   return
  }
  if !Allowed(identity,permission){
   http.Error(w,"personal settings unavailable to guests",http.StatusForbidden)
   return
  }
  next.ServeHTTP(w,r)
 })
}
