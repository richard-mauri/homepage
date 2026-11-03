package main

import (
 "strings"
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "testing"
 "time"
)

func TestStorageDisabledByDefault(t *testing.T) {
 mux:=http.NewServeMux()
 if err:=configureSuiteStorage(mux,"",time.Hour);err!=nil {t.Fatal(err)}
 w:=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/newsgo",nil))
 if w.Code!=http.StatusNotFound {t.Fatalf("expected 404, got %d",w.Code)}
}
func TestStorageOptInRequiresBearer(t *testing.T) {
 mux:=http.NewServeMux()
 root:=filepath.Join(t.TempDir(),"suite")
 if err:=configureSuiteStorage(mux,root,time.Hour);err!=nil{t.Fatal(err)}
 w:=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/mauristocks",nil))
 if w.Code!=http.StatusUnauthorized {t.Fatalf("expected 401, got %d",w.Code)}
 w=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/not-allowed",nil))
 if w.Code!=http.StatusNotFound{t.Fatalf("expected 404, got %d",w.Code)}
}

func TestAdminPageOptIn(t *testing.T) {
 mux:=http.NewServeMux()
 if err:=configureSuiteStorage(mux,"",time.Hour);err!=nil{t.Fatal(err)}
 w:=httptest.NewRecorder();mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/suite-admin",nil))
 if w.Code!=404{t.Fatalf("disabled admin page: %d",w.Code)}
 mux=http.NewServeMux()
 if err:=configureSuiteStorage(mux,filepath.Join(t.TempDir(),"suite"),time.Hour);err!=nil{t.Fatal(err)}
 w=httptest.NewRecorder();mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/suite-admin",nil))
 if w.Code!=200{t.Fatalf("enabled admin page: %d",w.Code)}
 if w.Header().Get("Cache-Control")!="no-store"{t.Fatal("admin page must not be cached")}
 if w.Header().Get("Content-Security-Policy")==""{t.Fatal("CSP is missing")}
}

func TestOptionalPersistenceConfig(t *testing.T){
 t.Setenv("SUITE_DATA_DIR","")
 if config().PersistenceAvailable{t.Fatal("persistence must default off")}
 t.Setenv("SUITE_DATA_DIR",filepath.Join(t.TempDir(),"suite"))
 if !config().PersistenceAvailable{t.Fatal("persistence availability not advertised")}
}

func TestHomepageAssetURLsAreBuildVersioned(t *testing.T) {
 page,err:=assets.ReadFile("static/index.html");if err!=nil{t.Fatal(err)}
 html:=renderHomepage(page)
 if strings.Contains(html,"{{HOMEPAGE_ASSET_VERSION}}"){t.Fatal("asset placeholder unrendered")}
 if !strings.Contains(html,"/static/app.js?v="){t.Fatal("JS asset URL unversioned")}
 if !strings.Contains(html,"/static/styles.css?v="){t.Fatal("CSS asset URL unversioned")}
 if strings.Contains(html,"/static/app.js\" defer"){t.Fatal("old JS URL persists")}
}
