package main

import (
 "embed"
 "encoding/json"
 "flag"
 "fmt"
 "io/fs"
 "log"
 "net/http"
 "net/url"
 "os"
 "strings"
 "time"

 "themauris.org/homepage/internal/suitestore"
)

var appVersion = "0.1.0"
var appBuild = "Build 000000"

//go:embed static/*
var assets embed.FS

type App struct {
 ID string `json:"id"`
 Name string `json:"name"`
 Subtitle string `json:"subtitle"`
 Description string `json:"description"`
 URL string `json:"url"`
 Category string `json:"category"`
 Icon string `json:"icon"`
 Accent string `json:"accent"`
 External bool `json:"external"`
}
type Config struct {
 SiteName string `json:"siteName"`
 Tagline string `json:"tagline"`
 Version string `json:"version"`
 Build string `json:"build"`
 PersistenceAvailable bool `json:"persistenceAvailable"`
 TraceEnabled bool `json:"traceEnabled"`
 Apps []App `json:"apps"`
}
func env(key, fallback string) string {
 if v:=strings.TrimSpace(os.Getenv(key)); v!="" {return v}
 return fallback
}
func config() Config {
 return Config{SiteName:"The Mauris",Version:appVersion,Build:appBuild,PersistenceAvailable:strings.TrimSpace(os.Getenv("SUITE_DATA_DIR"))!="",TraceEnabled:authTraceEnabled,Tagline:"A private corner of the web. Built for the things we follow, build, and explore.",Apps:[]App{
 {"sailing","Sailing Outlook","WEATHER & WATER","Live marine conditions, wind trends, and the Delta-to-Bay outlook.",env("SAILING_URL","https://weather.themauris.org/"),"ON THE WATER","sail","cyan",true},
 {"stocks","Stocks","MARKETS & PORTFOLIO","Market views, watchlists, and the signals worth paying attention to.",env("STOCKS_URL","https://stocks.themauris.org/"),"IN THE MARKETS","chart","mint",true},
 {"newsgo","NewsGo","NEWS INTELLIGENCE","Your news stream, organized for a faster read on what matters.",env("NEWSGO_URL","https://newsgo.themauris.org/"),"STAY INFORMED","news","amber",true},
 {"scrapergo","ScraperGo","DATA & AUTOMATION","Collection tools, feeds, and the machinery behind the scenes.",env("SCRAPERGO_URL","https://scrapergo.themauris.org/"),"IN THE WORKSHOP","code","violet",true},
 }}
}
func main() {
 version:=flag.Bool("version",false,"show version and exit")
 trace:=flag.Bool("trace",false,"trace authentication navigation without secrets to homepage.log")
 flag.Parse()
 if *version {fmt.Printf("Homepage v%s · %s\n",appVersion,appBuild);return}
 authTraceEnabled=*trace
 addr:=env("LISTEN_ADDR","127.0.0.1:8787")
 sub,err:=fs.Sub(assets,"static");if err!=nil {log.Fatal(err)}
 mux:=http.NewServeMux()
 mux.HandleFunc("/api/trace",clientAuthTraceHandler)
 if err:=configureSuiteStorage(mux,os.Getenv("SUITE_DATA_DIR"));err!=nil {log.Fatalf("Suite storage configuration: %v",err)}
 mux.Handle("/static/",http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-cache, must-revalidate")
  http.StripPrefix("/static/",http.FileServer(http.FS(sub))).ServeHTTP(w,r)
 }))
 mux.HandleFunc("/api/config",func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/api/config" {http.NotFound(w,r);return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  _=json.NewEncoder(w).Encode(config())
 })
 mux.HandleFunc("/healthz",func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Content-Type","text/plain")
  fmt.Fprintln(w,"ok")
 })
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/" {http.NotFound(w,r);return}
  w.Header().Set("Content-Type","text/html; charset=utf-8")
  w.Header().Set("Cache-Control","no-cache")
  page,err:=fs.ReadFile(sub,"index.html")
  if err!=nil {http.Error(w,"Homepage unavailable",http.StatusInternalServerError);return}
  _,_ = w.Write([]byte(renderHomepage(page)))
 })
 server:=&http.Server{Addr:addr,Handler:mux,ReadHeaderTimeout:5*time.Second,IdleTimeout:60*time.Second}
 log.Printf("The Mauris homepage listening at http://%s",addr)
 log.Fatal(server.ListenAndServe())
}

// configureSuiteStorage mounts the authenticated settings API only when an
// operator explicitly supplies a persistent directory. The default Homepage
// configuration remains a launcher with no storage or authentication endpoint.
// Session and bearer-token access are supported; cross-origin app integration remains disabled.
func configureSuiteStorage(mux *http.ServeMux, root string) error {
 if strings.TrimSpace(root)=="" { return nil }
 store,err:=suitestore.New(root,"newsgo","mauristocks","sailing")
 if err!=nil{return err}
 sessions:=newSessionService(store)
 if sessions.persistenceError!=nil {return fmt.Errorf("Homepage durable sessions unavailable: %w",sessions.persistenceError)}
 log.Printf("Homepage account and app handoff lifetime: %s",sessionLifetime)
 setSessionTraceLookup(sessions)
 mux.Handle("/api/storage/",sessions.suiteStorageHandler(store.Handler()))
 mux.Handle("/api/session",authTraceHTTP(http.HandlerFunc(sessions.handler)))
 mux.Handle("/api/session/renew",authTraceHTTP(http.HandlerFunc(sessions.renewHomepageSessionHandler)))
 mux.Handle("/api/session/profile",authTraceHTTP(http.HandlerFunc(sessions.profileHandler)))
 mux.Handle("/api/handoff/token",authTraceHTTP(http.HandlerFunc(sessions.handoffHandler)))
 mux.Handle("/launch/newsgo",authTraceHTTP(http.HandlerFunc(sessions.suiteRecoveryHandler)))
 mux.Handle("/launch/stocks",authTraceHTTP(http.HandlerFunc(sessions.suiteRecoveryHandler)))
 mux.Handle("/launch/sailing",authTraceHTTP(http.HandlerFunc(sessions.suiteRecoveryHandler)))
 mux.Handle("/api/handoff/renew",authTraceHTTP(http.HandlerFunc(sessions.renewHandoffHandler)))
 mux.Handle("/api/handoff/extend-session",authTraceHTTP(http.HandlerFunc(sessions.extendNewsGoSessionHandler)))
 activity:=newActivityCollectorWithRoot(root)
 mux.HandleFunc("/api/activity/ping",activity.homepagePresenceHandler(sessions))
 mux.HandleFunc("/api/admin/activity/ingest",activity.ingest)
 mux.HandleFunc("/api/admin/activity",activity.readAdmin(sessions))
 mux.HandleFunc("/api/admin/users",sessions.adminHandler)
 mux.HandleFunc("/suite-admin",func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/suite-admin" || r.Method!=http.MethodGet {http.NotFound(w,r);return}
  page,err:=assets.ReadFile("static/suite-admin.html")
  if err!=nil {http.Error(w,"Unavailable",500);return}
  w.Header().Set("Content-Type","text/html; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("Content-Security-Policy","default-src 'none'; script-src 'self' 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
  _,_=w.Write(page)
 })
 return nil
}

// renderHomepage stamps executable version and cache-busting asset URLs.
func renderHomepage(page []byte) string {
 html:=strings.ReplaceAll(string(page),"{{HOMEPAGE_VERSION}}",fmt.Sprintf("v%s · %s",appVersion,appBuild))
 return strings.ReplaceAll(html,"{{HOMEPAGE_ASSET_VERSION}}",url.QueryEscape(appVersion+"-"+appBuild))
}
