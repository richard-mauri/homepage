package main

import (
 "encoding/json"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestActivityHistoryPersistsWithoutIdentity(t *testing.T) {
 root:=t.TempDir()
 at:=time.Now().UTC().Truncate(time.Hour)
 c:=newActivityCollectorWithRoot(root)
 c.now=func()time.Time{return at}
 if err:=c.appendHistoryLocked("newsgo",activityReport{Active2:3,InFlight:2},at);err!=nil{t.Fatal(err)}
 if err:=c.appendHistoryLocked("newsgo",activityReport{Active2:1,InFlight:0},at.Add(time.Minute));err!=nil{t.Fatal(err)}
 if err:=c.flushHistory();err!=nil{t.Fatal(err)}
 history,status:=c.historySnapshot()
 if status!="available" || len(history)!=1 {t.Fatalf("history status=%s buckets=%d",status,len(history))}
 h:=history[0]
 if h.Samples!=2 || h.Active2Peak!=3 || h.Active2Sum!=4 || h.InFlightPeak!=2 || h.InFlightSum!=2 {
  t.Fatalf("unexpected aggregate %+v",h)
 }
 path:=filepath.Join(root,"activity","hourly.json")
 data,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 if strings.Contains(string(data),"token") || strings.Contains(string(data),"session") {t.Fatal("history contains private identifiers")}
 st,err:=os.Stat(path);if err!=nil{t.Fatal(err)}
 if st.Mode().Perm()!=0600 {t.Fatalf("history file permissions %v",st.Mode().Perm())}
 c2:=newActivityCollectorWithRoot(root)
 c2.now=func()time.Time{return at}
 recovered,status:=c2.historySnapshot()
 if status!="available" || len(recovered)!=1 || recovered[0].Samples!=2 {t.Fatal("history not recovered")}
 for _,v:=range c2.snapshot() {if v.Status!="unknown" {t.Fatal("stale presence restored as live presence")}}
}

func TestActivityHistoryPrunesOldBuckets(t *testing.T) {
 root:=t.TempDir()
 c:=newActivityCollectorWithRoot(root)
 at:=time.Now().UTC().Truncate(time.Hour)
 prior:=at.AddDate(0,0,-31)
 if err:=c.appendHistoryLocked("homepage",activityReport{Active2:5},prior);err!=nil{t.Fatal(err)}
 if err:=c.appendHistoryLocked("homepage",activityReport{Active2:2},at);err!=nil{t.Fatal(err)}
 hours,status:=c.historySnapshot()
 if status!="available" || len(hours)!=1 || hours[0].Active2Peak!=2 {t.Fatalf("prune failed: %v, %+v",status,hours)}
}

func TestActivityHistoryCorruptionDoesNotAffectSessionAvailability(t *testing.T) {
 root:=t.TempDir()
 path:=filepath.Join(root,"activity","hourly.json")
 if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(path,[]byte("{broken"),0600);err!=nil{t.Fatal(err)}
 c:=newActivityCollectorWithRoot(root)
 if _,status:=c.historySnapshot();status!="unavailable"{t.Fatal("corrupt history marked available")}
 if err:=c.appendHistoryLocked("newsgo",activityReport{Active2:1},time.Now());err==nil{t.Fatal("corrupt state silently overwritten")}
 stored,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 if string(stored)!="{broken" {t.Fatal("corrupt file unexpectedly destroyed")}
}

func TestActivityHistoryRejectsMalformedAndRepeatedBuckets(t *testing.T) {
 root:=t.TempDir()
 path:=filepath.Join(root,"activity","hourly.json")
 if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{t.Fatal(err)}
 at:=time.Now().UTC().Truncate(time.Hour).Format(time.RFC3339)
 duplicate:=activityHour{Service:"stocks",Hour:at,Samples:1,Active2Sum:1,Active2Peak:1}
 data,err:=json.Marshal(activityHistoryState{Version:1,Hours:[]activityHour{duplicate,duplicate}})
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(path,data,0600);err!=nil{t.Fatal(err)}
 if _,status:=newActivityCollectorWithRoot(root).historySnapshot();status!="unavailable" {t.Fatal("duplicate bucket accepted")}
}

func TestActivityHistoryWriteFailureDoesNotCommitSamples(t *testing.T) {
 c:=newActivityCollectorWithRoot(t.TempDir())
 // A directory path cannot be atomically replaced by the writer's file.
 c.historyPath=t.TempDir()
 at:=time.Now().UTC()
 if err:=c.appendHistoryLocked("sailing",activityReport{Active2:2},at);err==nil{t.Fatal("unwritable target accepted")}
 if h,_:=c.historySnapshot();len(h)!=0 {t.Fatal("failed write committed in-memory history")}
}
