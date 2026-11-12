package main

import (
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "sort"
 "time"
)

const activityHistoryDays = 30

// An hourly bucket summarizes samples received from a service. It is not a
// count of unique people over the hour: a visitor can occur in many samples.
type activityHour struct {
 Service string `json:"service"`
 Hour string `json:"hour"`
 Samples int `json:"samples"`
 Active2Sum int64 `json:"active2Sum"`
 Active2Peak int `json:"active2Peak"`
 InFlightSum int64 `json:"inFlightSum"`
 InFlightPeak int `json:"inFlightPeak"`
}
type activityHistoryState struct {
 Version int `json:"version"`
 Hours []activityHour `json:"hours"`
}

func newActivityCollectorWithRoot(root string) *activityCollector {
 c:=newActivityCollector()
 path,err:=activityStatePath(root)
 if err!=nil {c.historyError=err.Error();return c}
 c.historyPath=path
 if err=c.loadHistory();err!=nil {c.historyError=err.Error()}
 return c
}

func (c *activityCollector) loadHistory() error {
 if c.historyPath=="" {return errors.New("activity history path unavailable")}
 data,err:=os.ReadFile(c.historyPath)
 if errors.Is(err,os.ErrNotExist) {return nil}
 if err!=nil{return err}
 if len(data)>2*1024*1024{return errors.New("activity history exceeds size limit")}
 var state activityHistoryState
 if err=json.Unmarshal(data,&state);err!=nil{return fmt.Errorf("activity history unreadable: %w",err)}
 if state.Version!=1 || len(state.Hours)>5*24*activityHistoryDays+10 {
  return errors.New("activity history has an unsupported version or size")
 }
 validServices:=map[string]bool{}
 for _,name:=range activityServices {validServices[name]=true}
 cutoff:=c.now().UTC().AddDate(0,0,-activityHistoryDays).Truncate(time.Hour)
 for _,h:=range state.Hours {
  hour,err:=time.Parse(time.RFC3339,h.Hour)
  if !validServices[h.Service] || err!=nil || !hour.Equal(hour.UTC().Truncate(time.Hour)) ||
   h.Samples<=0 || h.Samples>100000 || h.Active2Sum<0 || h.InFlightSum<0 ||
   h.Active2Peak<0 || h.InFlightPeak<0 || h.Active2Peak>100000 || h.InFlightPeak>100000 {
   return errors.New("activity history has an invalid bucket")
  }
  if hour.Before(cutoff){continue}
  key:=h.Service+"|"+h.Hour
  if _,exists:=c.history[key];exists{return errors.New("activity history contains duplicate buckets")}
  c.history[key]=h
 }
 return nil
}

// appendHistoryLocked mutates the candidate map only. Caller commits to
// memory after a successful durable write, so disk errors cannot claim success.
func (c *activityCollector) appendHistoryLocked(service string, report activityReport, now time.Time) error {
 if c.historyError!="" {return errors.New("activity history unavailable")}
 if c.historyPath=="" {return errors.New("activity history not configured")}
 bucketStart:=now.UTC().Truncate(time.Hour).Format(time.RFC3339)
 cutoff:=now.UTC().AddDate(0,0,-activityHistoryDays).Truncate(time.Hour)
 next:=make(map[string]activityHour,len(c.history)+1)
 for key,h:=range c.history {
  t,e:=time.Parse(time.RFC3339,h.Hour)
  if e==nil && !t.Before(cutoff){next[key]=h}
 }
 key:=service+"|"+bucketStart
 h:=next[key]
 h.Service=service;h.Hour=bucketStart;h.Samples++
 h.Active2Sum+=int64(report.Active2)
 h.InFlightSum+=int64(report.InFlight)
 if report.Active2>h.Active2Peak {h.Active2Peak=report.Active2}
 if report.InFlight>h.InFlightPeak {h.InFlightPeak=report.InFlight}
 next[key]=h
 hours:=make([]activityHour,0,len(next))
 for _,item:=range next {hours=append(hours,item)}
 sort.Slice(hours,func(i,j int)bool {
  if hours[i].Hour!=hours[j].Hour{return hours[i].Hour<hours[j].Hour}
  return hours[i].Service<hours[j].Service
 })
 if err:=writeActivityHistory(c.historyPath,activityHistoryState{Version:1,Hours:hours});err!=nil {return err}
 c.history=next
 return nil
}

func writeActivityHistory(path string,state activityHistoryState) error {
 b,err:=json.Marshal(state)
 if err!=nil{return err}
 dir:=filepath.Dir(path)
 if err=os.MkdirAll(dir,0700);err!=nil{return err}
 tmp,err:=os.CreateTemp(dir,".hourly-*.tmp")
 if err!=nil{return err}
 defer os.Remove(tmp.Name())
 if err=tmp.Chmod(0600);err!=nil{tmp.Close();return err}
 if _,err=tmp.Write(b);err!=nil{tmp.Close();return err}
 if err=tmp.Sync();err!=nil{tmp.Close();return err}
 if err=tmp.Close();err!=nil{return err}
 if err=os.Rename(tmp.Name(),path);err!=nil{return err}
 d,err:=os.Open(dir)
 if err!=nil{return err}
 defer d.Close()
 return d.Sync()
}

// historySnapshot returns only aggregated buckets. No browser IDs, credentials,
// per-user activity or replay nonces are persisted or returned.
func (c *activityCollector) historySnapshot() ([]activityHour,string) {
 c.mu.Lock();defer c.mu.Unlock()
 if c.historyError!="" {return []activityHour{},"unavailable"}
 hours:=make([]activityHour,0,len(c.history))
 for _,v:=range c.history {hours=append(hours,v)}
 sort.Slice(hours,func(i,j int)bool {
  if hours[i].Hour!=hours[j].Hour{return hours[i].Hour<hours[j].Hour}
  return hours[i].Service<hours[j].Service
 })
 return hours,"available"
}
