package main

import (
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "testing"
)

func TestStorageDisabledByDefault(t *testing.T) {
 mux:=http.NewServeMux()
 if err:=configureSuiteStorage(mux,"");err!=nil {t.Fatal(err)}
 w:=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/newsgo",nil))
 if w.Code!=http.StatusNotFound {t.Fatalf("expected 404, got %d",w.Code)}
}
func TestStorageOptInRequiresBearer(t *testing.T) {
 mux:=http.NewServeMux()
 root:=filepath.Join(t.TempDir(),"suite")
 if err:=configureSuiteStorage(mux,root);err!=nil{t.Fatal(err)}
 w:=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/mauristocks",nil))
 if w.Code!=http.StatusUnauthorized {t.Fatalf("expected 401, got %d",w.Code)}
 w=httptest.NewRecorder()
 mux.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/storage/not-allowed",nil))
 if w.Code!=http.StatusNotFound{t.Fatalf("expected 404, got %d",w.Code)}
}
