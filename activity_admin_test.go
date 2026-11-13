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

func TestActivityAdminPanelEmbeddedAndReadOnly(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil{t.Fatal(err)}
 html:=string(page)
 for _,fragment:=range []string{
  `id="activityAdmin" hidden`,
  `id="activityServices"`,
  `id="activityHistoryMessage"`,
  `id="activityRange"`,
  `id="temporaryGuests"`,
  `id="guestUsers"`,
  `function renderTemporaryGuests(data)`,
  `Guests cannot save or restore settings`,
  `'/api/admin/activity'`,
  `session.role!=='admin'`,
  `No restart is initiated here.`,
  `No report since Homepage started`,
  "const created=await api(",
  "sessionToken=created.sessionToken;",
  "const headers=sessionToken?",
 }{
  if !strings.Contains(html,fragment){t.Errorf("missing admin panel invariant: %s",fragment)}
 }
 if strings.Contains(html,`/api/admin/activity/restart`) {t.Fatal("dashboard must not provide restart action")}
}

func TestActivityAdminReadRequiresVerifiedRole(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 adminToken:=strings.Repeat("a",64)
 userToken:=strings.Repeat("b",64)
 records:=[]suitestore.Credential{}
 for _,candidate:=range []struct{value,user,role string}{{adminToken,"owner","admin"},{userToken,"reader","user"}}{
  hash:=sha256.Sum256([]byte(candidate.value))
  records=append(records,suitestore.Credential{User:candidate.user,Role:candidate.role,Digest:hex.EncodeToString(hash[:])})
 }
 authDir:=filepath.Join(root,"auth")
 if err:=os.MkdirAll(authDir,0700);err!=nil{t.Fatal(err)}
 contents,err:=json.Marshal(records)
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(authDir,"approved-tokens.json"),contents,0600);err!=nil{t.Fatal(err)}
 sessions:=newSessionService(store)
 collector:=newActivityCollectorWithRoot(root)
 login:=func(approved string)*http.Cookie {
  r:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session",nil)
  r.Header.Set("Origin","https://home.themauris.org")
  r.Header.Set("Authorization","Bearer "+approved)
  w:=httptest.NewRecorder()
  sessions.handler(w,r)
  if w.Code!=http.StatusOK{t.Fatalf("login status %d",w.Code)}
  return w.Result().Cookies()[0]
 }
 check:=func(cookie *http.Cookie,origin string) *httptest.ResponseRecorder {
  r:=httptest.NewRequest(http.MethodGet,"https://home.themauris.org/api/admin/activity",nil)
  if cookie!=nil{r.AddCookie(cookie)}
  if origin!=""{r.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder()
  collector.readAdmin(sessions)(w,r)
  return w
 }
 if w:=check(nil,"");w.Code!=http.StatusUnauthorized{t.Fatalf("anonymous status %d",w.Code)}
 ordinary:=login(userToken)
 if w:=check(ordinary,"");w.Code!=http.StatusForbidden{t.Fatalf("ordinary user status %d",w.Code)}
 privileged:=login(adminToken)
 guestReq:=httptest.NewRequest(http.MethodPost,"https://home.themauris.org/api/session/guest",nil)
 guestReq.Header.Set("Origin","https://home.themauris.org")
 guestResp:=httptest.NewRecorder();sessions.guestSessionHandler(guestResp,guestReq)
 if guestResp.Code!=http.StatusOK{t.Fatalf("guest issue status %d",guestResp.Code)}
 var guestIssued struct {User string `json:"user"`}
 if err:=json.Unmarshal(guestResp.Body.Bytes(),&guestIssued);err!=nil{t.Fatal(err)}
 if w:=check(privileged,"https://stocks.themauris.org");w.Code!=http.StatusForbidden{t.Fatalf("cross origin status %d",w.Code)}
 result:=check(privileged,"https://home.themauris.org")
 if result.Code!=http.StatusOK{t.Fatalf("admin status %d: %s",result.Code,result.Body.String())}
 if result.Header().Get("Cache-Control")!="no-store"{t.Fatal("activity metrics must be uncacheable")}
 var data struct {
  Services []activityEntry `json:"services"`
  GuestUsers []guestRosterEntry `json:"guestUsers"`
  History []activityHour `json:"history"`
  HistoryStatus string `json:"historyStatus"`
 }
 if err:=json.Unmarshal(result.Body.Bytes(),&data);err!=nil{t.Fatal(err)}
 if len(data.GuestUsers)!=1 || data.GuestUsers[0].User!=guestIssued.User {t.Fatalf("admin guest roster missing: %+v",data.GuestUsers)}
 if len(data.Services)!=len(activityServices) || data.HistoryStatus!="available"{
  t.Fatalf("unexpected admin response: %+v",data)
 }
 for _,entry:=range data.Services {if entry.Status!="unknown"{t.Fatalf("unreported service must be unknown: %+v",entry)}}
}

func TestActivityAutomaticRefreshDoesNotFlashButton(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 for _,fragment:=range []string{
  "async function refreshActivity(manual=false)",
  "if(manual)el('activityRefresh').disabled=true;",
  "if(manual)el('activityRefresh').disabled=false;",
  "el('activityRefresh').addEventListener('click',()=>{void refreshActivity(true);});",
  "setInterval(()=>{if(!document.hidden&&!el('activityAdmin').hidden)void refreshActivity()},30000);",
 } {
  if !strings.Contains(html,fragment) {t.Errorf("missing manual-only refresh-button behavior: %s",fragment)}
 }
 if strings.Contains(html,"activityPending=true;el('activityRefresh').disabled=true;") {
  t.Fatal("automatic refresh must not dim the manual refresh button")
 }
}

func TestSuiteAdminUsesCollapsiblePanesAndManageSuiteLink(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 for _,fragment:=range []string{
  `<section id="login"><details class="suite-pane" open><summary>Connect</summary>`,
  `<section id="account" hidden><details class="suite-pane" open><summary class="session-summary"><span>Current Session</span><span id="remaining" class="session-time" role="timer" aria-label="Session time remaining">`,
  `<section id="admin" hidden><details class="suite-pane"><summary>Users &amp; Access</summary>`,
  `<details class="suite-pane nested"><summary>Create user</summary>`,
  `<title>Suite Management — The Mauris</title>`,
  `<h1>Suite Management</h1>`,
  `Manage your session, user access, and activity across the application suite.`,
  `<button id="renew" type="button">Renew</button>`,
  `<button id="logout" type="button">Disconnect</button>`,
  `<section id="activityAdmin" hidden aria-labelledby="activityHeading">`,
  `<details class="suite-pane" open><summary id="activityHeading">Session activity</summary>`,
  `<details class="suite-pane nested"><summary>Hourly activity patterns</summary>`,
  `id="loginForm"`,`id="createForm"`,`id="newToken"`,`id="activityServices"`,
  `el('admin').hidden=session.role!=='admin';`,
  `el('activityAdmin').hidden=session.role!=='admin';`,
 } {
  if !strings.Contains(html,fragment) {t.Errorf("missing collapsible suite section or access control: %s",fragment)}
 }
 if !strings.Contains(html,`Session --:--:--</span></summary>`) {
  t.Fatal("session countdown must remain visible in collapsed summary")
 }
 script,err:=assets.ReadFile("static/app.js")
 if err!=nil {t.Fatal(err)}
 if !strings.Contains(string(script),`manage.href='/suite-admin';manage.textContent='Manage Suite'`) {
  t.Fatal("Homepage management link must read Manage Suite and keep /suite-admin target")
 }
}

func TestActivityPanelLabelsObservedIdentifiersNotSessions(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 for _,fragment:=range []string{
  "grid-template-columns:minmax(0,1fr)",
  "Homepage and ScraperGo count verified signed-in sessions (with separate user counts) and guest tab identifiers;",
  "Presence means recently observed activity, not merely an unexpired login.",
  "Seen in last 2 min",
  "Seen in last 15 min",
  "Seen in last 60 min",
  "Guest IDs (last 2 min)",
  "Signed-in IDs (last 2 min)",
  "Signed-in sessions (last 2 min)",
  "Unique signed-in users (2 min)",
  "const sessionAware=key==='homepage'||key==='scrapergo';",
  "Requests in progress",
 } {
  if !strings.Contains(html,fragment) {t.Errorf("missing activity clarity invariant: %s",fragment)}
 }
 for _,old:=range []string{"Active in 2 minutes","Active in 15 minutes","Active in 60 minutes"} {
  if strings.Contains(html,old) {t.Errorf("ambiguous label remains: %s",old)}
 }
}

func TestActivityServiceCardsCollapseIndependentlyAndSurviveRefresh(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 for _,fragment:=range []string{
  "document.createElement('details');card.className='activity-service'",
  "card.dataset.service=key",
  "card.open=expanded.has(key)?expanded.get(key):key==='homepage';",
  "document.createElement('summary')",
  "const expanded=new Map(Array.from(parent.querySelectorAll('details[data-service]'),item=>[item.dataset.service,item.open]));",
  "title.append(name,state);",
  "card.append(title,stats);parent.append(card);",
  ".activity-service summary:focus-visible",
 } {
  if !strings.Contains(html,fragment) {t.Errorf("missing activity disclosure invariant: %s",fragment)}
 }
}

func TestAdminUserActivityIsScopedAndValidated(t *testing.T) {
 now:=time.Now().UTC().Format(time.RFC3339)
 base:=activityReport{Active2:1,Active5:1,Active15:1,Active60:1,Authenticated:1,UniqueUsers:1,
  Users:[]activityUserSummary{{User:"richard",Sessions:1,LastSeen:now}}}
 if !validActivityReport(base){t.Fatal("valid verified per-user aggregate rejected")}
 invalid:=base
 invalid.Users=[]activityUserSummary{{User:"richard",Sessions:2,LastSeen:now}}
 if validActivityReport(invalid){t.Fatal("over-counted per-user sessions accepted")}
 invalid.Users=[]activityUserSummary{{User:"richard",Sessions:1,LastSeen:now},{User:"richard",Sessions:0,LastSeen:now}}
 if validActivityReport(invalid){t.Fatal("duplicate user aggregate accepted")}
 html,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil{t.Fatal(err)}
 for _,required:=range []string{"Activity by application","function renderUserActivity(data)","data.homeUsers","scraper?.users","renderUserActivity(data);"} {
  if !strings.Contains(string(html),required){t.Errorf("missing user activity UI: %s",required)}
 }
}

func TestHomepagePerUserActivityGroupsSessionsWithoutLeakingDigests(t *testing.T) {
 root:=t.TempDir()
 store,err:=suitestore.New(root,"newsgo")
 if err!=nil{t.Fatal(err)}
 approved:=strings.Repeat("a",64)
 digest:=sha256.Sum256([]byte(approved))
 digestKey:=hex.EncodeToString(digest[:])
 authFile:=filepath.Join(root,"auth","approved-tokens.json")
 if err=os.MkdirAll(filepath.Dir(authFile),0700);err!=nil{t.Fatal(err)}
 credentials,_:=json.Marshal([]suitestore.Credential{{User:"richard",Role:"admin",Digest:digestKey}})
 if err=os.WriteFile(authFile,credentials,0600);err!=nil{t.Fatal(err)}
 service:=newSessionService(store)
 collector:=newActivityCollector()
 now:=time.Now().UTC().Truncate(time.Second)
 collector.now=func()time.Time{return now}
 collector.homePresence=map[string]homepagePresence{
  "tab1":{Seen:now,SignedIn:true,SessionKey:"one",UserKey:digestKey},
  "tab2":{Seen:now,SignedIn:true,SessionKey:"one",UserKey:digestKey},
  "tab3":{Seen:now,SignedIn:true,SessionKey:"two",UserKey:digestKey},
  "guest":{Seen:now},
 }
 users:=collector.homepageUsersSnapshot(service)
 if len(users)!=1 || users[0].User!="richard" || users[0].Sessions!=2 || users[0].LastSeen=="" {
  t.Fatalf("incorrect per-user session attribution: %+v",users)
 }
 if data,_:=json.Marshal(users);strings.Contains(string(data),digestKey) {
  t.Fatal("raw digest was exposed in admin user summary")
 }
}
