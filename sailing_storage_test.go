package main

import (
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestSailingAccountStorageRegisteredAndCORS(t *testing.T) {
 mux := http.NewServeMux()
 if err := configureSuiteStorage(mux, filepath.Join(t.TempDir(), "suite")); err != nil { t.Fatal(err) }
 if !allowedStorageOrigin("https://weather.themauris.org") { t.Fatal("SailingGo origin must be trusted") }
 if allowedStorageOrigin("https://attacker.themauris.org") { t.Fatal("unexpected trusted subdomain") }
 request := httptest.NewRequest(http.MethodOptions, "https://home.themauris.org/api/storage/sailing", nil)
 request.Header.Set("Origin", "https://weather.themauris.org")
 request.Header.Set("Access-Control-Request-Method", "PUT")
 request.Header.Set("Access-Control-Request-Headers", "content-type, authorization")
 response := httptest.NewRecorder()
 mux.ServeHTTP(response, request)
 if response.Code != http.StatusNoContent { t.Fatalf("preflight: %d %s", response.Code, response.Body.String()) }
 if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://weather.themauris.org" { t.Fatalf("unexpected CORS origin %q", got) }
 unauth := httptest.NewRequest(http.MethodGet, "https://home.themauris.org/api/storage/sailing", nil)
 unauth.Header.Set("Origin", "https://weather.themauris.org")
 response = httptest.NewRecorder()
 mux.ServeHTTP(response, unauth)
 if response.Code != http.StatusUnauthorized { t.Fatalf("unauthenticated sailing settings: %d", response.Code) }
}

func TestSailingHomepageLauncherUsesSessionHandoff(t *testing.T) {
 js, err := os.ReadFile("static/app.js")
 if err != nil { t.Fatal(err) }
 source := string(js)
 if !strings.Contains(source, "url.origin!=='https://weather.themauris.org'") { t.Fatal("SailingGo must use authenticated launch path") }
 if !strings.Contains(source, "url.hash='suite_token='") { t.Fatal("expected fragment-only handoff") }
 if !strings.Contains(source, "url.hash='suite_mode=local'") { t.Fatal("expected signed-out browser-only launch") }
}
