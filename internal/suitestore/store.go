// Package suitestore provides a centralized, per-user JSON settings store.
// It is intentionally not mounted by Homepage until authentication and migration
// have been reviewed. No token values are written to disk.
package suitestore

import (
 "context"
 "crypto/sha256"
 "crypto/subtle"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "regexp"
 "strings"
 "sync"
 "time"
)

var safeName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
var legacyUserName = regexp.MustCompile(`^[a-zA-Z0-9_ -]{1,64}$`)

type Credential struct {
 User string `json:"user"`
 Digest string `json:"digest"`
 Label string `json:"label,omitempty"`
 Role string `json:"role,omitempty"`
}
type Record struct {
 Revision uint64 `json:"revision"`
 UpdatedAt string `json:"updatedAt"`
 Data json.RawMessage `json:"data"`
}
type Store struct {
 mu sync.Mutex
 root string
 namespaces map[string]bool
}
func New(root string, namespaces ...string) (*Store,error) {
 if root=="" { return nil, errors.New("storage root required") }
 allowed:=make(map[string]bool)
 for _, n:=range namespaces {
  if !safeName.MatchString(n) { return nil, fmt.Errorf("invalid namespace %q",n) }
  allowed[n]=true
 }
 if len(allowed)==0 { return nil,errors.New("at least one namespace required") }
 if err:=os.MkdirAll(root,0700);err!=nil{return nil,err}
 return &Store{root:root,namespaces:allowed},nil
}
func (s *Store) credentialPath() string {return filepath.Join(s.root,"auth","approved-tokens.json")}
func (s *Store) recordPath(user, app string) string {return filepath.Join(s.root,"users",user,app+".json")}

// Authenticate reads the existing NewsGo approved-token JSON schema and maps
// the bearer token to its stable user and role. Credentials are never logged.
func (s *Store) Authenticate(token string) (string,string,error) {
 if len(token)!=64 {return "","",errors.New("invalid token")}
 raw,err:=hex.DecodeString(token)
 if err!=nil || len(raw)!=32 {return "","",errors.New("invalid token")}
 sum:=sha256.Sum256([]byte(token))
 return s.AuthenticateDigest(hex.EncodeToString(sum[:]))
}
// AuthenticateDigest validates a stored token digest against the current
// approved-token registry, so revocation is checked on every session request.
func (s *Store) AuthenticateDigest(wanted string) (string,string,error) {
 if len(wanted)!=64 {return "","",errors.New("invalid digest")}
 b,err:=os.ReadFile(s.credentialPath());if err!=nil{return "","",err}
 var entries []Credential
 if err=json.Unmarshal(b,&entries);err!=nil{return "","",err}
 for _,v:=range entries {
  if !legacyUserName.MatchString(v.User) {continue}
  if len(v.Digest)==64 && subtle.ConstantTimeCompare([]byte(strings.ToLower(v.Digest)),[]byte(wanted))==1 {
   role:="user";if v.Role=="admin"{role="admin"}
   return v.User,role,nil
  }
 }
 return "","",errors.New("unapproved token")
}
type identityKey struct{}
// WithIdentity allows the trusted in-process Homepage session validator to
// establish an identity; it cannot be set by an HTTP header or client input.
func WithIdentity(r *http.Request,user string) *http.Request {
 return r.WithContext(context.WithValue(r.Context(),identityKey{},user))
}

func atomicWrite(path string, contents []byte) error {
 dir:=filepath.Dir(path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 file,err:=os.CreateTemp(dir,".suite-*.tmp");if err!=nil{return err}
 defer os.Remove(file.Name())
 if err=file.Chmod(0600);err!=nil{file.Close();return err}
 if _,err=file.Write(contents);err!=nil{file.Close();return err}
 if err=file.Sync();err!=nil{file.Close();return err}
 if err=file.Close();err!=nil{return err}
 return os.Rename(file.Name(),path)
}
func (s *Store) load(user,app string) (Record,bool,error) {
 b,err:=os.ReadFile(s.recordPath(user,app))
 if errors.Is(err,os.ErrNotExist){return Record{},false,nil}
 if err!=nil{return Record{},false,err}
 var rec Record
 if err=json.Unmarshal(b,&rec);err!=nil{return Record{},false,err}
 return rec,true,nil
}
func (s *Store) Save(user,app string,expected uint64,data json.RawMessage)(Record,error){
 if !legacyUserName.MatchString(user)||!s.namespaces[app]{return Record{},errors.New("invalid identity or namespace")}
 if len(data)==0 || len(data)>128<<10 || !json.Valid(data){return Record{},errors.New("invalid data")}
 s.mu.Lock();defer s.mu.Unlock()
 current,exists,err:=s.load(user,app);if err!=nil{return Record{},err}
 if exists && current.Revision!=expected || !exists && expected!=0{return Record{},ErrConflict}
 next:=Record{Revision:expected+1,UpdatedAt:time.Now().UTC().Format(time.RFC3339),Data:append(json.RawMessage(nil),data...)}
 b,err:=json.MarshalIndent(next,"","  ");if err!=nil{return Record{},err}
 if err=atomicWrite(s.recordPath(user,app),append(b,'\n'));err!=nil{return Record{},err}
 return next,nil
}
var ErrConflict=errors.New("revision conflict")

// Handler exposes GET/PUT /api/storage/{app}. It does not accept user IDs in
// paths, query strings or bodies. Cross-origin writes are rejected.
func (s *Store) Handler() http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  if r.Method!=http.MethodGet && r.Method!=http.MethodPut {w.Header().Set("Allow","GET, PUT");http.Error(w,"method not allowed",405);return}
  app,ok:=strings.CutPrefix(r.URL.Path,"/api/storage/")
  if !ok || !s.namespaces[app] {http.NotFound(w,r);return}
  user,_:=r.Context().Value(identityKey{}).(string)
  if user=="" {
   auth:=r.Header.Get("Authorization")
   if !strings.HasPrefix(auth,"Bearer "){http.Error(w,"authentication required",401);return}
   var err error
   user,_,err=s.Authenticate(strings.TrimPrefix(auth,"Bearer "))
   if err!=nil {http.Error(w,"authentication failed",401);return}
  }
  if r.Method==http.MethodPut {
   origin:=r.Header.Get("Origin")
   if origin!="" && origin!="https://"+r.Host && origin!="http://"+r.Host {http.Error(w,"cross-origin write rejected",403);return}
   r.Body=http.MaxBytesReader(w,r.Body,128<<10)
   var input struct {Revision uint64 `json:"revision"`;Data json.RawMessage `json:"data"`}
   dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
   if err:=dec.Decode(&input);err!=nil{http.Error(w,"invalid JSON",400);return}
   var extra any
   if err:=dec.Decode(&extra);err!=io.EOF {http.Error(w,"trailing JSON",400);return}
   rec,err:=s.Save(user,app,input.Revision,input.Data)
   if errors.Is(err,ErrConflict){http.Error(w,"revision conflict",409);return}
   if err!=nil{http.Error(w,"save failed",400);return}
   w.Header().Set("Content-Type","application/json");_ = json.NewEncoder(w).Encode(rec);return
  }
  s.mu.Lock();rec,exists,err:=s.load(user,app);s.mu.Unlock()
  if err!=nil{http.Error(w,"storage unavailable",500);return}
  if !exists{http.Error(w,"no saved settings",404);return}
  w.Header().Set("Content-Type","application/json");_ = json.NewEncoder(w).Encode(rec)
 })
}
