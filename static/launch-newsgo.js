'use strict';
// This code executes only on Homepage's origin. It never saves approved
// personal tokens or exports Homepage's parent session ID to a sister app.
const newsGoDestination='https://newsgo.themauris.org/';
const homeLoginKey='mauri-home-session-v1';
const status=document.getElementById('status');

async function issueHandoff(sessionID){
 const headers=sessionID?{Authorization:'Session '+sessionID}:{};
 return fetch('/api/handoff/token',{
  method:'POST',credentials:'same-origin',cache:'no-store',headers
 });
}
async function recoverNewsGo(){
 let storedSession='';
 try{
  const value=JSON.parse(localStorage.getItem(homeLoginKey)||'null');
  if(value && typeof value.token==='string' && /^[a-f0-9]{64}$/.test(value.token)){
   // Cached expiration is not authoritative: Homepage may have renewed it.
   storedSession=value.token;
  }
 }catch(_){/* Cookie fallback below is still valid. */}
 try{
  let response=await issueHandoff(storedSession);
  // An obsolete Homepage localStorage ID may coexist with a newer live
  // host-only session cookie; let Homepage revalidate either identity.
  if(storedSession && response.status===401)response=await issueHandoff('');
  if(response.status===401){
   location.replace(newsGoDestination+'#suite_mode=local');
   return;
  }
  if(!response.ok)throw Error('Homepage handoff unavailable (HTTP '+response.status+')');
  const data=await response.json();
  if(!data || typeof data.token!=='string' || !/^[a-f0-9]{64}$/.test(data.token) ||
      !Number.isFinite(Date.parse(data.expiresAt)) || Date.parse(data.expiresAt)<=Date.now()){
   throw Error('Invalid Homepage handoff response');
  }
  // The child token is a URL fragment, never a query string or parent ID.
  location.replace(newsGoDestination+'#suite_token='+data.token);
 }catch(e){
  // An unavailable server is not proof of sign-out; stay on Homepage
  // and offer explicit browser-only navigation rather than misreporting it.
  status.textContent='Could not verify the Homepage session: '+e.message+'. Retry by reloading this page or continue without an account.';
 }
}
void recoverNewsGo();
