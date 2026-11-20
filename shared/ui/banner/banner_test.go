package banner

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "bytes"
 "testing"
)

func TestRenderEscapesAndShowsVerifiedState(t *testing.T) {
 var b bytes.Buffer
 err:=Render(&b,Config{AppName:`NewsGo <test>`,Version:"0.2.58",Build:"000160",HomeURL:"https://home.themauris.org/",AccountState:SignedIn,AccountName:`richard<script>`})
 if err!=nil {t.Fatal(err)}
 html:=b.String()
 for _,want:=range []string{`NewsGo &lt;test&gt;`,`richard&lt;script&gt;`,`Back to Homepage`,`Build 000160`} {
  if !strings.Contains(html,want){t.Errorf("missing %q",want)}
 }
 if strings.Contains(html,"<script>"){t.Fatal("unescaped HTML")}
}

func TestInvalidConfigurationFailsWithoutWriting(t *testing.T) {
 configs:=[]Config{
  {AppName:""},
  {AppName:"News",AccountState:SignedIn},
  {AppName:"News",AccountState:"unknown"},
  {AppName:"News",AccountState:Guest,HomeURL:"https://example.com"},
  {AppName:"News",AccountState:Guest,HomeURL:"https://home.themauris.org.evil.com"},
  {AppName:"News",AccountState:Guest,HomeURL:"//evil.example/"},
  {AppName:"News",AccountState:Guest,TannerURL:"javascript:alert(1)"},
 }
 for _,cfg:=range configs {
  var b bytes.Buffer
  if err:=Render(&b,cfg);err==nil {t.Errorf("expected error for %+v",cfg)}
  if b.Len()!=0 {t.Fatal("invalid input wrote partial markup")}
 }
}

func TestStatesAndFallback(t *testing.T) {
 for _,state:=range []AccountState{Guest,BrowserOnly,Unavailable} {
  var b bytes.Buffer
  if err:=Render(&b,Config{AppName:"ScraperGo",AccountState:state});err!=nil {t.Fatal(err)}
  if !strings.Contains(b.String(),`suite-banner__mark`) {t.Fatal("missing image fallback")}
 }
}

func TestStylesheetIsLocalAndReadOnly(t *testing.T) {
 h:=Stylesheet([]byte(".suite-ui{color:red}"))
 w:=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/suite-ui/banner.css",nil))
 if w.Code!=200 || !strings.Contains(w.Body.String(),".suite-banner") || !strings.Contains(w.Body.String(),".suite-ui") {t.Fatal("missing embedded styles")}
 if w.Header().Get("Content-Type")!="text/css; charset=utf-8" {t.Fatal("wrong CSS content type")}
 for _,method:=range []string{http.MethodPost,http.MethodPut} {
  w=httptest.NewRecorder()
  h.ServeHTTP(w,httptest.NewRequest(method,"/suite-ui/banner.css",nil))
  if w.Code!=http.StatusMethodNotAllowed {t.Errorf("%s returned %d",method,w.Code)}
 }
}
