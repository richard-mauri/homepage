// Package banner renders the common Mauris Suite masthead without depending on
// Homepage's authentication or session implementation.
package banner

import (
 "bytes"
 "embed"
 "errors"
 "html/template"
 "io"
 "net/http"
 "net/url"
 "strings"
)

//go:embed banner.html banner.css status.js
var files embed.FS

var page = template.Must(template.ParseFS(files, "banner.html"))

// AccountState describes display only; callers must verify identity separately.
type AccountState string

const (
 SignedIn AccountState = "signed-in"
 Guest AccountState = "guest"
 BrowserOnly AccountState = "browser-only"
 Unavailable AccountState = "unavailable"
)

type Config struct {
 AppName string
 Version string
 Build string
 Description string
 HomeURL string
 TannerURL string
 AccountState AccountState
 AccountName string
}

type templateData struct {
 Config
 AccountLabel string
 BuildLabel string
}

// Render writes escaped markup. It does not supply account verification or any
// session controls. HomeURL permits a local path or the fixed suite Homepage
// origin; TannerURL must be a locally served absolute path.
func Render(w io.Writer, cfg Config) error {
 if strings.TrimSpace(cfg.AppName)=="" { return errors.New("banner: app name required") }
 if !safeHomeURL(cfg.HomeURL) || !safePath(cfg.TannerURL) { return errors.New("banner: invalid HomeURL or TannerURL") }
 data:=templateData{Config:cfg}
 data.BuildLabel=strings.TrimSpace(cfg.Build)
 if strings.HasPrefix(strings.ToLower(data.BuildLabel),"build ") {data.BuildLabel=strings.TrimSpace(data.BuildLabel[6:])}
 switch cfg.AccountState {
 case SignedIn:
  if strings.TrimSpace(cfg.AccountName)=="" { return errors.New("banner: signed-in identity required") }
  data.AccountLabel="Signed in as "+cfg.AccountName
 case Guest:
  data.AccountLabel="Guest access"
 case BrowserOnly:
  data.AccountLabel="Browser-only mode"
 case Unavailable:
  data.AccountLabel="Account status unavailable"
 default:
  return errors.New("banner: invalid account state")
 }
 var out bytes.Buffer
 if err:=page.Execute(&out,data);err!=nil{return err}
 _,err:=io.Copy(w,&out)
 return err
}

func safePath(value string) bool {
 if value=="" {return true}
 if !strings.HasPrefix(value,"/") || strings.HasPrefix(value,"//") || strings.ContainsAny(value,"\r\n\\") {return false}
 parsed,err:=url.Parse(value)
 return err==nil && parsed.Scheme=="" && parsed.Host=="" && parsed.User==nil
}

func safeHomeURL(value string) bool {
 if safePath(value) { return true }
 parsed,err:=url.Parse(value)
 return err==nil && parsed.Scheme=="https" && parsed.Host=="home.themauris.org" && parsed.User==nil && parsed.RawQuery=="" && parsed.Fragment=="" && (parsed.Path=="" || parsed.Path=="/")
}

// CSS combines the shared design tokens and banner-specific stylesheet.
// Both are distributed by the same Go module version.
func CSS(designCSS []byte) []byte {
 result:=make([]byte,0,len(designCSS)+1+len(bannerCSS()))
 result=append(result,designCSS...)
 result=append(result,'\n')
 return append(result,bannerCSS()...)
}

func bannerCSS() []byte {css,_:=files.ReadFile("banner.css");return css}

// ScriptJS is the canonical browser-side account-label formatter. Consumers
// must verify identity themselves before calling SuiteBannerStatus.signedIn.
func ScriptJS() []byte {code,_:=files.ReadFile("status.js");return append([]byte(nil),code...)}

// StatusScript serves the shared display-only formatter from each app's origin.
func StatusScript() http.Handler {
 contents:=ScriptJS()
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!=http.MethodGet && r.Method!=http.MethodHead {w.Header().Set("Allow","GET, HEAD");http.Error(w,"method not allowed",http.StatusMethodNotAllowed);return}
  w.Header().Set("Content-Type","application/javascript; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  if r.Method==http.MethodGet {_,_=w.Write(contents)}
 })
}

// Stylesheet returns a handler that consumers can mount on their own origin.
func Stylesheet(designCSS []byte) http.Handler {
 contents:=CSS(designCSS)
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!=http.MethodGet && r.Method!=http.MethodHead {w.Header().Set("Allow","GET, HEAD");http.Error(w,"method not allowed",http.StatusMethodNotAllowed);return}
  w.Header().Set("Content-Type","text/css; charset=utf-8")
  w.Header().Set("Cache-Control","public, max-age=3600")
  if r.Method==http.MethodGet {_,_=w.Write(contents)}
 })
}
