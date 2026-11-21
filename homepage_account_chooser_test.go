package main

import (
 "strings"
 "testing"
)

func TestHomepageAccountChooserAndSessionLifecycle(t *testing.T) {
 raw,err:=assets.ReadFile("static/app.js")
 if err!=nil{t.Fatal(err)}
 script:=string(raw)
 for _,want:=range []string{
  "title.textContent='Sign in'",
  "personalChoice.textContent='Sign in as another user'",
  "guestChoice.textContent='Sign in as Guest'",
  "signOutChoice.textContent='Sign out'",
  "personalChoice.addEventListener('click'",
  "guestChoice.addEventListener('click'",
  "signOutChoice.addEventListener('click'",
  "form.hidden=true;choices.hidden=!!sessionRole",
  "title.textContent=sessionRole?'Account':'Sign in'",
  "signOutChoice.hidden=!sessionRole",
  "if(sessionRole)signOutChoice.focus();else personalChoice.focus();",
  "if(sessionRole||signOutPending)return",
  "if(sessionRole){message.textContent='Sign out before signing in as another user.';return;}",
  "if(sessionRole){message.textContent='Sign out before choosing another identity.';return;}",
  "signIn.textContent='Account'",
  "signIn.textContent='Sign in'",
  "sessionRole=state.session.role;sessionUser=state.session.user",
  "signedIn.hidden=false;signIn.hidden=false",
  "method:'DELETE',credentials:'same-origin'",
  "if(!response.ok)throw Error('Could not sign out",
  "showSignedOut();success=true",
  "if(success)hideDialog()",
  "rememberHomeLogin(await response.json())",
  "if(!await refresh())throw Error('Homepage sign-in session was not retained.",
  "if(authenticated)hideDialog()",
 } {
  if !strings.Contains(script,want){t.Errorf("account-choice contract missing %q",want)}
 }
 if strings.Contains(script,"Currently signed in as "+'"'+"+"){
  t.Fatal("old account-choice prompt should not be shown to signed-in users")
 }
 if !strings.Contains(script,"const controls=[close,...(sessionRole?[signOutChoice]"){
  t.Fatal("keyboard focus must be trapped among visible authenticated controls only")
 }
 if strings.Contains(script,"suiteGuestSignInButton") || strings.Contains(script,"guestButton"){
  t.Fatal("guest authentication must be offered only inside the shared Sign in dialog")
 }
 if !strings.Contains(script,"container.append(signIn,retryAuth,signedIn)") ||
    !strings.Contains(script,"signedIn.hidden=true;signIn.hidden=false;retryAuth.hidden=true;manage.hidden=true;"){
  t.Fatal("signed-out Homepage must display just one Sign in button")
 }
 if strings.Contains(script,"signIn.hidden=state.session.role!=='guest'"){
  t.Fatal("personal accounts must retain access to the account chooser")
 }
 style,err:=assets.ReadFile("static/styles.css")
 if err!=nil{t.Fatal(err)}
 if !strings.Contains(string(style),".suite-modal [hidden]{display:none!important}"){
  t.Fatal("chooser/form hidden state must override author display styling")
 }
}
