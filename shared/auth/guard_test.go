package auth

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestSuiteAccessMatrix(t *testing.T){
 cases:=[]struct{name string;who Identity;app,settings bool}{
  {"anonymous",Identity{},false,false},
  {"client-claimed user",Identity{User:"reader",Role:RoleUser},false,false},
  {"verified no username",Identity{Role:RoleUser,Verified:true},false,false},
  {"verified unknown role",Identity{User:"reader",Role:"root",Verified:true},false,false},
  {"personal",Identity{User:"reader",Role:RoleUser,Verified:true},true,true},
  {"admin",Identity{User:"operator",Role:RoleAdmin,Verified:true},true,true},
  {"guest",Identity{User:"guest-1",Role:RoleGuest,Verified:true},true,false},
 }
 for _,tc:=range cases {t.Run(tc.name,func(t *testing.T){
  if got:=Allowed(tc.who,UseApplication);got!=tc.app{t.Errorf("app=%v want %v",got,tc.app)}
  if got:=Allowed(tc.who,UsePersonalSettings);got!=tc.settings{t.Errorf("settings=%v want %v",got,tc.settings)}
 })}
}

func TestFailClosedGuard(t *testing.T){
 testCases:=[]struct{
  name string
  identity Identity
  verifyErr error
  permission Permission
  want int
 }{
  {"anonymous",Identity{},nil,UseApplication,401},
  {"spoofed role but unverified",Identity{User:"admin",Role:RoleAdmin},nil,UseApplication,401},
  {"personal app",Identity{User:"reader",Role:RoleUser,Verified:true},nil,UseApplication,200},
  {"personal settings",Identity{User:"reader",Role:RoleUser,Verified:true},nil,UsePersonalSettings,200},
  {"guest app",Identity{User:"guest",Role:RoleGuest,Verified:true},nil,UseApplication,200},
  {"guest settings",Identity{User:"guest",Role:RoleGuest,Verified:true},nil,UsePersonalSettings,403},
  {"unknown role",Identity{User:"admin",Role:"unknown",Verified:true},nil,UseApplication,401},
  {"profile outage",Identity{},ErrVerificationUnavailable,UseApplication,503},
  {"other verifier error",Identity{},errors.New("timeout"),UseApplication,503},
 }
 for _,tc:=range testCases{t.Run(tc.name,func(t *testing.T){
  called:=false
  next:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){called=true;w.WriteHeader(200)})
  verify:=func(_ context.Context,_ *http.Request)(Identity,error){return tc.identity,tc.verifyErr}
  gate:=Guard(verify,tc.permission,next)
  req:=httptest.NewRequest("GET","http://app.example/api/private",nil)
  // A forged role label must be ignored unless the real verifier accepts it.
  req.Header.Set("X-Suite-Role","admin")
  rec:=httptest.NewRecorder()
  gate.ServeHTTP(rec,req)
  if rec.Code!=tc.want{t.Fatalf("status=%d want=%d",rec.Code,tc.want)}
  if called!=(tc.want==200){t.Fatal("downstream access violated")}
  if rec.Header().Get("Cache-Control")!="no-store"{t.Fatal("gate must not be cached")}
 })}
 for _,tc:=range []struct{name string;verify Verifier;handler http.Handler}{
  {"missing verifier",nil,http.HandlerFunc(func(http.ResponseWriter,*http.Request){t.Fatal("called")})},
  {"missing target",func(context.Context,*http.Request)(Identity,error){return Identity{},nil},nil},
 } {t.Run(tc.name,func(t *testing.T){
  r:=httptest.NewRecorder()
  Guard(tc.verify,UseApplication,tc.handler).ServeHTTP(r,httptest.NewRequest("GET","/",nil))
  if r.Code!=503{t.Fatalf("unexpected status %d",r.Code)}
 })}
}
