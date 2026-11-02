package suitestore

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "strings"
)

var ErrUnknownUser = errors.New("unknown user")
var ErrExistingUser = errors.New("user already exists")
var ErrLastAdmin = errors.New("cannot remove the last administrator")

// ListUsers returns metadata only. No credential digests or plaintext tokens.
func (s *Store) ListUsers() ([]Credential,error) {
 s.mu.Lock();defer s.mu.Unlock()
 entries,err:=s.readCredentials();if err!=nil{return nil,err}
 users:=make([]Credential,0,len(entries))
 for _,v:=range entries{users=append(users,Credential{User:v.User,Label:v.Label,Role:normalizedRole(v.Role)})}
 return users,nil
}
func normalizedRole(role string) string {if role=="admin"{return "admin"};return "user"}
func (s *Store) readCredentials() ([]Credential,error) {
 b,err:=os.ReadFile(s.credentialPath());if err!=nil{return nil,err}
 var entries []Credential
 if err=json.Unmarshal(b,&entries);err!=nil{return nil,err}
 return entries,nil
}
func (s *Store) writeCredentials(entries []Credential) error {
 b,err:=json.MarshalIndent(entries,"","  ");if err!=nil{return err}
 return atomicWrite(s.credentialPath(),append(b,'\n'))
}
func (s *Store) IssueUserToken(user,label,role string,rotate bool)(string,error){
 if !legacyUserName.MatchString(user)||len(strings.TrimSpace(user))==0{return "",errors.New("invalid username")}
 if len(label)>128{return "",errors.New("label too long")}
 if role!="user"&&role!="admin"{return "",errors.New("invalid role")}
 var random [32]byte
 if _,err:=rand.Read(random[:]);err!=nil{return "",err}
 token:=hex.EncodeToString(random[:])
 hash:=sha256.Sum256([]byte(token))
 s.mu.Lock();defer s.mu.Unlock()
 entries,err:=s.readCredentials();if err!=nil{return "",err}
 idx:=-1
 for i,v:=range entries {if v.User==user {idx=i;break}}
 if rotate {
  if idx<0{return "",ErrUnknownUser}
  // Rotation keeps the existing role; it must not elevate a user.
  entries[idx].Digest=hex.EncodeToString(hash[:])
 }else{
  if idx>=0{return "",ErrExistingUser}
  entries=append(entries,Credential{User:user,Label:label,Role:role,Digest:hex.EncodeToString(hash[:])})
 }
 if err=s.writeCredentials(entries);err!=nil{return "",fmt.Errorf("token save: %w",err)}
 return token,nil
}
func (s *Store) SetUserRole(user,role string)error{
 if role!="user"&&role!="admin"{return errors.New("invalid role")}
 s.mu.Lock();defer s.mu.Unlock()
 entries,err:=s.readCredentials();if err!=nil{return err}
 idx:=-1;adminCount:=0
 for i,v:=range entries{if v.Role=="admin"{adminCount++};if v.User==user{idx=i}}
 if idx<0{return ErrUnknownUser}
 if entries[idx].Role=="admin"&&role!="admin"&&adminCount<=1{return ErrLastAdmin}
 entries[idx].Role=role
 return s.writeCredentials(entries)
}
func (s *Store) RevokeUser(user string)error{
 s.mu.Lock();defer s.mu.Unlock()
 entries,err:=s.readCredentials();if err!=nil{return err}
 idx:=-1;adminCount:=0
 for i,v:=range entries{if v.Role=="admin"{adminCount++};if v.User==user{idx=i}}
 if idx<0{return ErrUnknownUser}
 if entries[idx].Role=="admin"&&adminCount<=1{return ErrLastAdmin}
 entries=append(entries[:idx],entries[idx+1:]...)
 // Preferences deliberately remain owned by stable user identity.
 return s.writeCredentials(entries)
}
