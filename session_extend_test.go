package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"themauris.org/homepage/internal/suitestore"
)

func TestNewsGoSessionExtension(t *testing.T) {
	root := t.TempDir()
	store, err := suitestore.New(root, "newsgo", "mauristocks")
	if err != nil { t.Fatal(err) }
	approved := strings.Repeat("d", 64)
	hash := sha256.Sum256([]byte(approved))
	authFile := filepath.Join(root, "auth", "approved-tokens.json")
	if err := os.MkdirAll(filepath.Dir(authFile), 0700); err != nil { t.Fatal(err) }
	credentials, _ := json.Marshal([]suitestore.Credential{{User:"reader", Role:"user", Digest:hex.EncodeToString(hash[:])}})
	if err := os.WriteFile(authFile, credentials, 0600); err != nil { t.Fatal(err) }
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	fresh := func() *sessionService {
		s := newSessionService(store)
		if s.persistenceError != nil { t.Fatal(s.persistenceError) }
		s.now = func() time.Time { return now }
		return s
	}
	svc := fresh()
	login := httptest.NewRequest(http.MethodPost, "https://home.themauris.org/api/session", nil)
	login.Header.Set("Origin", "https://home.themauris.org")
	login.Header.Set("Authorization", "Bearer "+approved)
	out := httptest.NewRecorder()
	svc.handler(out, login)
	if out.Code != 200 { t.Fatalf("login: %d %s",out.Code,out.Body.String()) }
	var parent struct { SessionToken, ExpiresAt string }
	if err := json.Unmarshal(out.Body.Bytes(), &parent); err != nil { t.Fatal(err) }
	issue := httptest.NewRequest(http.MethodPost, "https://home.themauris.org/api/handoff/token", nil)
	issue.Header.Set("Origin", "https://home.themauris.org")
	issue.Header.Set("Authorization", "Session "+parent.SessionToken)
	out = httptest.NewRecorder()
	svc.handoffHandler(out, issue)
	if out.Code != 200 { t.Fatalf("handoff: %d",out.Code) }
	var handoff struct { Token, ExpiresAt string }
	if err := json.Unmarshal(out.Body.Bytes(), &handoff); err != nil { t.Fatal(err) }
	oldBearer := handoff.Token
	extend := func(token, origin string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "https://home.themauris.org/api/handoff/extend-session", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Authorization", "Bearer "+token)
		result := httptest.NewRecorder()
		svc.extendNewsGoSessionHandler(result, req)
		return result
	}
	for _,tc := range []struct { origin string; want int }{
		{"https://evil.example",403},{"https://stocks.themauris.org",403},{"",403},
	} {
		if got:=extend(oldBearer,tc.origin).Code;got!=tc.want { t.Fatalf("origin %q accepted: %d",tc.origin,got) }
	}
	preflight := httptest.NewRequest(http.MethodOptions,"https://home.themauris.org/api/handoff/extend-session",nil)
	preflight.Header.Set("Origin","https://newsgo.themauris.org")
	preflight.Header.Set("Access-Control-Request-Method","POST")
	preflight.Header.Set("Access-Control-Request-Headers","authorization")
	out=httptest.NewRecorder();svc.extendNewsGoSessionHandler(out,preflight)
	if out.Code!=204 || out.Header().Get("Access-Control-Allow-Origin")!="https://newsgo.themauris.org" {t.Fatalf("preflight failed: %d",out.Code)}
	now=now.Add(4*time.Hour)
	out=extend(oldBearer,"https://newsgo.themauris.org")
	if out.Code!=200 {t.Fatalf("extension: %d %s",out.Code,out.Body.String())}
	var renewed struct {Token,ExpiresAt,User,Role string}
	if err:=json.Unmarshal(out.Body.Bytes(),&renewed);err!=nil{t.Fatal(err)}
	expected:=now.Add(12*time.Hour).Format(time.RFC3339)
	if renewed.Token==oldBearer || len(renewed.Token)!=64 || renewed.ExpiresAt!=expected || renewed.User!="reader" || renewed.Role!="user" {
		t.Fatalf("invalid session renewal response: %+v", struct { ExpiresAt, User, Role string }{renewed.ExpiresAt,renewed.User,renewed.Role})
	}
	if got:=extend(oldBearer,"https://newsgo.themauris.org").Code;got!=401 {t.Fatalf("replay accepted: %d",got)}
	svc=fresh()
	getParent:=func() *httptest.ResponseRecorder{
		req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
		req.Header.Set("Authorization","Session "+parent.SessionToken)
		result:=httptest.NewRecorder();svc.handler(result,req);return result
	}
	if result:=getParent();result.Code!=200 || !strings.Contains(result.Body.String(),expected) {
		t.Fatalf("extended parent lost after restart: %d",result.Code)
	}
	getProfile:=func(token string) int {
		req:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session/profile",nil)
		req.Header.Set("Origin","https://newsgo.themauris.org")
		req.Header.Set("Authorization","Bearer "+token)
		result:=httptest.NewRecorder();svc.profileHandler(result,req);return result.Code
	}
	if got:=getProfile(renewed.Token);got!=200 {t.Fatalf("new bearer rejected after restart: %d",got)}
	if got:=getProfile(oldBearer);got!=401 {t.Fatalf("old bearer survived rotation: %d",got)}
	now=now.Add(8*time.Hour+time.Second)
	if getParent().Code!=200 || getProfile(renewed.Token)!=200 {t.Fatal("parent/child expired at old deadline")}
	now=now.Add(4*time.Hour)
	if getParent().Code!=401 || getProfile(renewed.Token)!=401 {t.Fatal("renewal outlived new 12-hour deadline")}
	if got:=extend(renewed.Token,"https://newsgo.themauris.org").Code;got!=401 {t.Fatalf("expired parent renewed: %d",got)}
	now=now.Add(-time.Hour)
	if err:=os.WriteFile(authFile,[]byte("[]"),0600);err!=nil{t.Fatal(err)}
	if got:=extend(renewed.Token,"https://newsgo.themauris.org").Code;got!=401 {t.Fatalf("revoked credential renewed: %d",got)}
}

func TestNewsGoSessionExtensionRejectsUnboundBearer(t *testing.T) {
	root:=t.TempDir()
	store,err:=suitestore.New(root,"newsgo");if err!=nil{t.Fatal(err)}
	approved:=strings.Repeat("a",64)
	hash:=sha256.Sum256([]byte(approved))
	dir:=filepath.Join(root,"auth")
	if err:=os.MkdirAll(dir,0700);err!=nil{t.Fatal(err)}
	creds,_:=json.Marshal([]suitestore.Credential{{User:"reader",Digest:hex.EncodeToString(hash[:])}})
	if err:=os.WriteFile(filepath.Join(dir,"approved-tokens.json"),creds,0600);err!=nil{t.Fatal(err)}
	svc:=newSessionService(store)
	req:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/token",nil)
	req.Header.Set("Origin","https://home.themauris.org")
	req.Header.Set("Authorization","Bearer "+approved)
	out:=httptest.NewRecorder();svc.handoffHandler(out,req)
	if out.Code!=200{t.Fatalf("credential handoff: %d",out.Code)}
	var response struct{ Token string }
	if err:=json.Unmarshal(out.Body.Bytes(),&response);err!=nil{t.Fatal(err)}
	renew:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/handoff/extend-session",nil)
	renew.Header.Set("Origin","https://newsgo.themauris.org")
	renew.Header.Set("Authorization","Bearer "+response.Token)
	out=httptest.NewRecorder();svc.extendNewsGoSessionHandler(out,renew)
	if out.Code!=401{t.Fatalf("credential-only handoff extended a login: %d",out.Code)}
}
