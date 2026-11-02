package main

import (
 "embed"
 "encoding/json"
 "flag"
 "fmt"
 "io/fs"
 "log"
 "net/http"
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
 Apps []App `json:"apps"`
}
func env(key, fallback string) string {
 if v:=strings.TrimSpace(os.Getenv(key)); v!="" {return v}
 return fallback
}
func config() Config {
 return Config{SiteName:"The Mauris",Version:appVersion,Build:appBuild,Tagline:"A private corner of the web. Built for the things we follow, build, and explore.",Apps:[]App{
 {"sailing","Sailing Outlook","WEATHER & WATER","Live marine conditions, wind trends, and the Delta-to-Bay outlook.",env("SAILING_URL","https://weather.themauris.org/"),"ON THE WATER","sail","cyan",true},
 {"stocks","Stocks","MARKETS & PORTFOLIO","Market views, watchlists, and the signals worth paying attention to.",env("STOCKS_URL","https://stocks.themauris.org/"),"IN THE MARKETS","chart","mint",true},
 {"newsgo","NewsGo","NEWS INTELLIGENCE","Your news stream, organized for a faster read on what matters.",env("NEWSGO_URL","https://newsgo.themauris.org/"),"STAY INFORMED","news","amber",true},
 {"scrapergo","ScraperGo","DATA & AUTOMATION","Collection tools, feeds, and the machinery behind the scenes.",env("SCRAPERGO_URL","https://scrapergo.themauris.org/"),"IN THE WORKSHOP","code","violet",true},
 }}
}
func main() {
 version:=flag.Bool("version",false,"show version and exit")
 flag.Parse()
 if *version {fmt.Printf("Homepage v%s · %s\n",appVersion,appBuild);return}
 addr:=env("LISTEN_ADDR","127.0.0.1:8787")
 sub,err:=fs.Sub(assets,"static");if err!=nil {log.Fatal(err)}
 mux:=http.NewServeMux()
 if err:=configureSuiteStorage(mux,os.Getenv("SUITE_DATA_DIR"));err!=nil {log.Fatalf("Suite storage configuration: %v",err)}
 mux.Handle("/static/",http.StripPrefix("/static/",http.FileServer(http.FS(sub))))
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
  _,_ = w.Write([]byte(strings.ReplaceAll(string(page),"{{HOMEPAGE_VERSION}}",fmt.Sprintf("v%s · %s",appVersion,appBuild))))
 })
 server:=&http.Server{Addr:addr,Handler:mux,ReadHeaderTimeout:5*time.Second,IdleTimeout:60*time.Second}
 log.Printf("The Mauris homepage listening at http://%s",addr)
 log.Fatal(server.ListenAndServe())
}

// configureSuiteStorage mounts the authenticated settings API only when an
// operator explicitly supplies a persistent directory. The default Homepage
// configuration remains a launcher with no storage or authentication endpoint.
// This is bearer-token-only; cross-subdomain browser sessions are not yet built.
func configureSuiteStorage(mux *http.ServeMux, root string) error {
 if strings.TrimSpace(root)=="" { return nil }
 store,err:=suitestore.New(root,"newsgo","mauristocks")
 if err!=nil{return err}
 mux.Handle("/api/storage/",store.Handler())
 return nil
}
