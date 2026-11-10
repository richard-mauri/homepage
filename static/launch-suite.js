'use strict';
// This bridge runs on Homepage's origin, where its persistent session ID and
// host-only cookie can be accessed. Only explicitly allowlisted destinations
// are supported; neither an URL query nor a referrer controls the destination.
const destinations=Object.freeze({
 '/launch/newsgo':'https://newsgo.themauris.org/',
 '/launch/stocks':'https://stocks.themauris.org/',
 '/launch/sailing':'https://weather.themauris.org/'
});
const suiteDestination=destinations[location.pathname];
const homeLoginKey='mauri-home-session-v1';
const statusElement=document.getElementById('status');
const localLink=document.getElementById('local-link');

async function issueHandoff(sessionID){
 const headers=sessionID?{Authorization:'Session '+sessionID}:{};
 return fetch('/api/handoff/token',{
  method:'POST',credentials:'same-origin',cache:'no-store',headers
 });
}
async function recoverSuite(){
 if(!suiteDestination){
  statusElement.textContent='Unknown suite application. Return to Homepage.';
  return;
 }
 localLink.href=suiteDestination+'#suite_mode=local';
 let storedSession='';
 try{
  const value=JSON.parse(localStorage.getItem(homeLoginKey)||'null');
  if(value && typeof value.token==='string' && /^[a-f0-9]{64}$/.test(value.token)){
   // Cached expiration is not authoritative: Homepage may have renewed it.
   storedSession=value.token;
  }
 }catch(_){/* Cookie fallback is still available. */}
 try{
  let response=await issueHandoff(storedSession);
  // A stale localStorage ID may coexist with a current host-only cookie.
  if(storedSession && response.status===401)response=await issueHandoff('');
  if(response.status===401){
   location.replace(suiteDestination+'#suite_mode=local');
   return;
  }
  if(!response.ok)throw Error('Homepage handoff unavailable (HTTP '+response.status+')');
  const data=await response.json();
  if(!data || typeof data.token!=='string' || !/^[a-f0-9]{64}$/.test(data.token) ||
      !Number.isFinite(Date.parse(data.expiresAt)) || Date.parse(data.expiresAt)<=Date.now()){
   throw Error('Invalid Homepage handoff response');
  }
  // Only a child bearer is passed in a URL fragment, not a query or parent ID.
  location.replace(suiteDestination+'#suite_token='+data.token);
 }catch(e){
  // Transient network or persistence failures are not authoritative sign-out.
  statusElement.textContent='Could not verify the Homepage session: '+e.message+'. Retry by reloading this page or continue without an account.';
 }
}
void recoverSuite();
