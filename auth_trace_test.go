package main

import (
 "bytes"
 "encoding/json"
 "log"
 "net/http"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
)

func TestAuthTraceOptInAndRedaction(t *testing.T){
 previous:=authTraceEnabled
 previousLookup:=sessionTraceLookup
 defer func(){authTraceEnabled=previous;sessionTraceLookup=previousLookup}()
 var output bytes.Buffer
 original:=log.Writer()
 log.SetOutput(&output)
 defer log.SetOutput(original)
 wrapped:=authTraceHTTP(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  http.Error(w,"session invalid",http.StatusUnauthorized)
 }))
 request:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/session",nil)
 rawSession:=strings.Repeat("a",64)
 rawBearer:=strings.Repeat("b",64)
 request.Header.Set("Authorization","Session "+rawSession)
 request.AddCookie(&http.Cookie{Name:sessionCookieName,Value:rawBearer})
 authTraceEnabled=false
 response:=httptest.NewRecorder()
 wrapped.ServeHTTP(response,request)
 if output.Len()!=0 {t.Fatal("trace emitted when disabled")}
 authTraceEnabled=true
 output.Reset()
 response=httptest.NewRecorder()
 wrapped.ServeHTTP(response,request)
 if response.Code!=401 {t.Fatalf("status changed: %d",response.Code)}
 message:=output.String()
 for _,want:=range []string{"[AUTH-TRACE]","endpoint=/api/session","status=401","auth=session"} {
  if !strings.Contains(message,want){t.Errorf("missing diagnostic %q",want)}
 }
 for _,secret:=range []string{rawSession,rawBearer} {
  if strings.Contains(message,secret){t.Fatal("authentication secret leaked into logs")}
 }
}

func TestClientAuthTraceStrictAllowlist(t *testing.T){
 previous:=authTraceEnabled
 defer func(){authTraceEnabled=previous}()
 authTraceEnabled=true
 send:=func(body,origin string)int{
  request:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/trace",strings.NewReader(body))
  request.Header.Set("Origin",origin)
  response:=httptest.NewRecorder()
  clientAuthTraceHandler(response,request)
  return response.Code
 }
 valid,_:=json.Marshal(clientTraceEvent{Event:"page_init",Phase:"initial",Status:0,SessionStored:true,Page:"0123456789abcdef"})
 if got:=send(string(valid),"https://home.themauris.org");got!=204{t.Fatalf("valid trace status: %d",got)}
 if got:=send(string(valid),"https://stocks.themauris.org");got!=403{t.Fatalf("foreign-origin trace status: %d",got)}
 if got:=send(`{"event":"page_init","phase":"initial","page":"0123456789abcdef","token":"secret"}`,"https://home.themauris.org");got!=400 {
  t.Fatalf("unknown secret-bearing field accepted: %d",got)
 }
 if got:=send(`{"event":"arbitrary","phase":"initial","page":"0123456789abcdef"}`,"https://home.themauris.org");got!=400 {
  t.Fatalf("unknown event accepted: %d",got)
 }
 authTraceEnabled=false
 if got:=send(string(valid),"https://home.themauris.org");got!=404 {t.Fatalf("disabled trace endpoint status: %d",got)}
}

func TestTraceClientCoverage(t *testing.T){
 js,err:=os.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 for _,want:=range []string{
  "authTraceActive=!!c.traceEnabled",
  "traceClient('refresh_response'",
  "traceClient('refresh_error'",
  "traceClient('ui_signed_out'",
  "traceClient('ui_signed_in'",
  "traceClient('launch_begin'",
  "traceClient('handoff_response'",
  "traceClient('page_show'",
  "traceClient('page_focus'",
  "traceClient('page_online'",
  "headers:tracedHeaders(headers)",
 }{
  if !strings.Contains(string(js),want){t.Errorf("browser trace instrumentation missing %q",want)}
 }
}
