package main

import (
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "strings"

 "themauris.org/homepage/internal/suitestore"
)

// adminHandler is deliberately not exposed until the operator enables SUITE_DATA_DIR.
// An existing admin must authenticate; there is no unauthenticated bootstrap path.
func (s *sessionService) adminHandler(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if !validSessionOrigin(r) && r.Method!=http.MethodGet {http.Error(w,"invalid origin",403);return}
 role:=""
 if _,err:=r.Cookie(sessionCookieName);err==nil{
  _,resolved,ok:=s.authenticateSession(r)
  if !ok{http.Error(w,"authentication failed",401);return}
  role=resolved
 }else{
  if r.Method!=http.MethodGet && r.Header.Get("Origin")==""{http.Error(w,"origin required",403);return}
  token,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
  if !ok{http.Error(w,"authentication required",401);return}
  _,resolved,err:=s.store.Authenticate(token)
  if err!=nil{http.Error(w,"authentication failed",401);return}
  role=resolved
 }
 if role!="admin"{http.Error(w,"admin required",403);return}
 w.Header().Set("Content-Type","application/json")
 if r.URL.Path!="/api/admin/users"{http.NotFound(w,r);return}
 if r.Method==http.MethodGet {
  users,err:=s.store.ListUsers();if err!=nil{http.Error(w,"registry unavailable",500);return}
  _=json.NewEncoder(w).Encode(users);return
 }
 if r.Method!=http.MethodPost {w.Header().Set("Allow","GET, POST");http.Error(w,"method not allowed",405);return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var input struct {Action string `json:"action"`;User string `json:"user"`;Label string `json:"label"`;Role string `json:"role"`}
 decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields()
 if err:=decoder.Decode(&input);err!=nil{http.Error(w,"invalid JSON",400);return}
 var extra any
 if err:=decoder.Decode(&extra);err!=io.EOF{http.Error(w,"trailing JSON",400);return}
 switch input.Action {
 case "issue","rotate":
  role:=input.Role
  if input.Action=="rotate"{role="user"} // rotation retains original role in store
  if role==""{role="user"}
  token,err:=s.store.IssueUserToken(input.User,input.Label,role,input.Action=="rotate")
  if err!=nil{http.Error(w,"token operation failed: "+safeAdminError(err),400);return}
  _=json.NewEncoder(w).Encode(map[string]string{"user":input.User,"token":token})
 case "role":
  if err:=s.store.SetUserRole(input.User,input.Role);err!=nil{http.Error(w,safeAdminError(err),400);return}
  _=json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
 case "revoke":
  if err:=s.store.RevokeUser(input.User);err!=nil{http.Error(w,safeAdminError(err),400);return}
  _=json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
 default:
  http.Error(w,"unknown action",400)
 }
}
func safeAdminError(err error)string{
 switch {
 case errors.Is(err,suitestore.ErrExistingUser):return "user already exists"
 case errors.Is(err,suitestore.ErrUnknownUser):return "unknown user"
 case errors.Is(err,suitestore.ErrLastAdmin):return "cannot remove last admin"
 default:return "invalid request or storage failure"
 }
}
