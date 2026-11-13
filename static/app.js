const icons = {
 sail:'<path d="M12 3v19M12 5 4 17h8M14 8l7 9h-7M3 21c2-1 4-1 6 0s4 1 6 0 4-1 6 0"/>',
 chart:'<path d="M3 3v18h18M7 16l4-5 4 3 6-8"/><path d="M17 6h4v4"/>',
 news:'<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 7h8M8 11h8M8 15h3M14 15h2"/>',
 code:'<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m9 9-3 3 3 3m6-6 3 3-3 3m-4 5 2-10"/>'
};
const accent={cyan:['#68cde0','rgba(104,205,224,.12)'],mint:['#75ddb6','rgba(117,221,182,.12)'],amber:['#eac081','rgba(234,192,129,.12)'],violet:['#b7a1f6','rgba(183,161,246,.12)']};
function svg(name){return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[name]||icons.code}</svg>`}
function makeCard(app){const a=document.createElement('a');a.className='app-card';a.href=app.url;a.rel='noopener noreferrer';const [color,glow]=accent[app.accent]||accent.cyan;a.style.setProperty('--card-accent',color);a.style.setProperty('--card-glow',glow);const top=document.createElement('div');top.className='card-top';const icon=document.createElement('span');icon.className='card-icon';icon.innerHTML=svg(app.icon);const arrow=document.createElement('span');arrow.className='card-arrow';arrow.textContent='↗';top.append(icon,arrow);const content=document.createElement('div');content.className='card-content';for(const [cls,val,tag] of [['card-category',app.category,'div'],['card-name',app.name,'h3'],['card-subtitle',app.subtitle,'div'],['card-description',app.description,'p']]){let el=document.createElement(tag);el.className=cls;el.textContent=val;content.append(el)}const bottom=document.createElement('div');bottom.className='card-bottom';bottom.innerHTML='<span>OPEN APPLICATION</span><span class="direction">→</span>';a.append(top,content,bottom);return a}
// Handoff tokens are kept only in memory and put in URL fragments for
// approved suite destinations. The durable sign-in token is never forwarded.
// Homepage has its own twelve-hour sign-in credential, distinct from
// the short-lived app token. Store only this temporary session ID, never
// the approved personal sign-in token.
// Trace is opt-in via the server's -trace flag. Never report credential or
// bearer values. The page ID is random diagnostics metadata, not a login key.
let authTraceActive=false;
const authTracePage=Array.from(crypto.getRandomValues(new Uint8Array(8)),n=>n.toString(16).padStart(2,'0')).join('');
function traceClient(event,phase='-',status=0){
 if(!authTraceActive)return;
 const body=JSON.stringify({event,phase,status,sessionStored:!!homeSessionToken,page:authTracePage});
 // Use keepalive so the navigation click can be observed before unload.
 try{void fetch('/api/trace',{method:'POST',credentials:'same-origin',headers:{'Content-Type':'application/json'},body,keepalive:true}).catch(()=>{});}catch(_){}
}
function tracedHeaders(headers={}){return {...headers,...(authTraceActive?{'X-Homepage-Trace':authTracePage}:{})};}
const homeLoginKey='mauri-home-session-v1';
let homeSessionToken='',homeSessionExpires=0;
try{
 const session=JSON.parse(localStorage.getItem(homeLoginKey)||'null');
 if(session && /^[a-f0-9]{64}$/.test(session.token) && Number.isFinite(session.expires)){
  // A sister app may extend the durable parent login while this page is away.
  // Retain the opaque ID for server validation even if the cached expiry passed.
  homeSessionToken=session.token;homeSessionExpires=session.expires;
 }
}catch(_){}
function clearHomeLogin(){
 homeSessionToken='';homeSessionExpires=0;
 try{localStorage.removeItem(homeLoginKey)}catch(_){}
}
function rememberHomeLogin(data){
 if(!/^[a-f0-9]{64}$/.test(data.sessionToken)||!Number.isFinite(Date.parse(data.expiresAt)))throw Error('Invalid Homepage session');
 homeSessionToken=data.sessionToken;homeSessionExpires=Date.parse(data.expiresAt);
 try{localStorage.setItem(homeLoginKey,JSON.stringify({token:homeSessionToken,expires:homeSessionExpires}))}
 catch(_){throw Error('Browser storage unavailable; Homepage cannot retain your sign-in across application navigation.')}
}
const homeSessionHeaders=()=>homeSessionToken?{Authorization:'Session '+homeSessionToken}:{};
let suiteHandoffToken='',suiteHandoffExpires=0;
try{const saved=JSON.parse(sessionStorage.getItem('suite-handoff-v1')||'null');if(saved&&/^[a-f0-9]{64}$/.test(saved.token)&&saved.expires>Date.now()){suiteHandoffToken=saved.token;suiteHandoffExpires=saved.expires}}catch(_){}
function applyHandoffLinks(){
 // Keep card destinations free of cached credentials. Only attach a fresh
 // bearer when navigating from an authenticated Homepage session.
 for(const link of document.querySelectorAll('#cards a.app-card')){
  const url=new URL(link.href);
  if(url.origin!=='https://newsgo.themauris.org' && url.origin!=='https://stocks.themauris.org' && url.origin!=='https://weather.themauris.org' && url.origin!=='https://scrapergo.themauris.org')continue;
  url.hash='';
  link.href=url.toString();
 }
}
async function renewHandoff(approvedToken=''){
 try{
  const response=await fetch('/api/handoff/token',{method:'POST',credentials:'same-origin',cache:'no-store',headers:tracedHeaders(approvedToken?{Authorization:'Bearer '+approvedToken}:homeSessionHeaders())});
  traceClient('handoff_response',response.ok?'valid':'rejected',response.status);
  if(!response.ok)throw Error('not authenticated');
  const data=await response.json();
  if(!/^[a-f0-9]{64}$/.test(data.token)||!Date.parse(data.expiresAt))throw Error('bad handoff');
  suiteHandoffToken=data.token;suiteHandoffExpires=Date.parse(data.expiresAt);
  try{sessionStorage.setItem('suite-handoff-v1',JSON.stringify({token:suiteHandoffToken,expires:suiteHandoffExpires}))}catch(_){}
  applyHandoffLinks();return true;
 }catch(_){traceClient('handoff_response','network');suiteHandoffToken='';suiteHandoffExpires=0;try{sessionStorage.removeItem('suite-handoff-v1')}catch(_){}applyHandoffLinks();return false}
}
// A verified launch is a last-seen observation, not a live heartbeat.
// Do not delay navigation indefinitely if optional telemetry is unavailable.
async function acknowledgeSuiteLaunch(service){
 if(!homeSessionToken)return;
 const controller=new AbortController();
 const timeout=setTimeout(()=>controller.abort(),1500);
 try{
  await fetch('/api/activity/launch',{method:'POST',credentials:'same-origin',cache:'no-store',headers:{...homeSessionHeaders(),'Content-Type':'application/json'},body:JSON.stringify({service}),signal:controller.signal});
 }catch(_){}finally{clearTimeout(timeout)}
}
// Mint a fresh handoff at application launch instead of reusing an older
// token created when Homepage obtained focus.
async function openSuiteApp(event){
 const link=event.currentTarget;
 const url=new URL(link.href);
 if(url.origin!=='https://newsgo.themauris.org' && url.origin!=='https://stocks.themauris.org' && url.origin!=='https://weather.themauris.org' && url.origin!=='https://scrapergo.themauris.org')return;
 event.preventDefault();
 traceClient('launch_begin',url.origin==='https://stocks.themauris.org'?'stocks':url.origin==='https://weather.themauris.org'?'sailing':url.origin==='https://scrapergo.themauris.org'?'scrapergo':'newsgo');
 // A successful server-issued handoff is required for every Homepage launch.
 // Do not rely on a cached expiry: the server may have revoked this session.
 const issued=await renewHandoff();
 if(issued && suiteHandoffToken && Date.now()<suiteHandoffExpires){
  await acknowledgeSuiteLaunch(url.origin==='https://newsgo.themauris.org'?'newsgo':url.origin==='https://scrapergo.themauris.org'?'scrapergo':url.origin==='https://stocks.themauris.org'?'stocks':'sailing');
  traceClient('launch_handoff','success');
  url.hash='suite_token='+encodeURIComponent(suiteHandoffToken);
 }else{
  traceClient('launch_blocked','unauthenticated');
  let notice=document.querySelector('#suiteLaunchNotice');
  if(!notice){
   notice=document.createElement('p');notice.id='suiteLaunchNotice';notice.className='suite-launch-notice';
   notice.setAttribute('role','alert');
   document.querySelector('#cards').before(notice);
  }
  notice.textContent='Sign in on Homepage or choose Sign in as Guest before opening a suite application.';
  notice.scrollIntoView({block:'center',behavior:'smooth'});
  return;
 }
 const notice=document.querySelector('#suiteLaunchNotice');if(notice)notice.remove();
 location.assign(url.toString());
}
async function init(){const root=document.querySelector('#cards');try{const r=await fetch('/api/config',{cache:'no-store'});if(!r.ok)throw Error('Configuration unavailable');const c=await r.json();authTraceActive=!!c.traceEnabled;traceClient('page_init','initial');root.replaceChildren(...c.apps.map(makeCard));applyHandoffLinks();for(const link of root.querySelectorAll('a.app-card'))link.addEventListener('click',openSuiteApp);document.querySelector('.count').textContent='/ '+String(c.apps.length).padStart(2,'0');if(c.persistenceAvailable)enableOptionalAuth()}catch(e){root.textContent='Applications are temporarily unavailable. Please refresh to try again.'}}
function tick(){document.querySelector('#clock').textContent=new Intl.DateTimeFormat('en-US',{hour:'2-digit',minute:'2-digit',hour12:false,timeZone:'America/Los_Angeles'}).format(new Date())}
const themeKey='themauris-theme';try{if(localStorage.getItem(themeKey)==='light')document.documentElement.dataset.theme='light'}catch(_){}document.querySelector('#themeToggle').addEventListener('click',()=>{const next=document.documentElement.dataset.theme==='light'?'dark':'light';document.documentElement.dataset.theme=next;try{localStorage.setItem(themeKey,next)}catch(_){}});document.querySelector('#year').textContent=new Date().getFullYear();tick();setInterval(tick,30_000);init();

/* A verified Homepage identity is required to launch suite applications.
   Signing in does not upload or replace application localStorage. */
function enableOptionalAuth(){
 if(document.querySelector('#suiteAccount'))return;
 const top=document.querySelector('.topright');
 const container=document.createElement('div');container.id='suiteAccount';container.className='suite-account';
 const signIn=document.createElement('button');signIn.type='button';signIn.textContent='Sign in';
 const guestButton=document.createElement('button');guestButton.type='button';guestButton.textContent='Sign in as Guest';
 const retryAuth=document.createElement('button');retryAuth.type='button';retryAuth.textContent='Connection issue — Retry';retryAuth.hidden=true;
 const signedIn=document.createElement('div');signedIn.className='suite-account-auth';signedIn.hidden=true;
 const identity=document.createElement('span');
 const manage=document.createElement('a');manage.href='/suite-admin';manage.textContent='Manage Suite';manage.hidden=true;
 signedIn.append(identity,manage);
 container.append(signIn,guestButton,retryAuth,signedIn);top.prepend(container);
 const overlay=document.createElement('div');overlay.className='suite-modal-backdrop';overlay.hidden=true;
 const dialog=document.createElement('section');dialog.className='suite-modal';dialog.setAttribute('role','dialog');dialog.setAttribute('aria-modal','true');dialog.setAttribute('aria-labelledby','suiteDialogTitle');
 const header=document.createElement('div');header.className='suite-modal-header';
 const title=document.createElement('h2');title.id='suiteDialogTitle';title.textContent='Personal settings';
 const close=document.createElement('button');close.type='button';close.textContent='Close';close.setAttribute('aria-label','Close sign-in dialog');
 header.append(title,close);
 const description=document.createElement('p');description.textContent='Sign in with your approved personal token to launch suite applications and use your personal settings. Or continue as a guest for one hour. Existing local settings will not be changed.';
 const form=document.createElement('form');
 const label=document.createElement('label');label.htmlFor='suiteToken';label.textContent='Approved personal token';
 const token=document.createElement('input');token.id='suiteToken';token.type='password';token.autocomplete='off';token.spellcheck=false;token.required=true;token.minLength=64;token.maxLength=64;
 const submit=document.createElement('button');submit.type='submit';submit.className='suite-signin-submit';submit.textContent='Sign in';
 form.append(label,token,submit);
 const message=document.createElement('p');message.className='suite-modal-message';message.setAttribute('role','status');message.setAttribute('aria-live','polite');
 dialog.append(header,description,form,message);overlay.append(dialog);document.body.append(overlay);
 let previousFocus=null,signInPending=false;
 function hideDialog(){if(signInPending)return;overlay.hidden=true;token.value='';message.textContent='';if(previousFocus)previousFocus.focus();}
 function showDialog(){previousFocus=document.activeElement;overlay.hidden=false;message.textContent='';token.focus();}
 signIn.addEventListener('click',showDialog);
 close.addEventListener('click',hideDialog);
 overlay.addEventListener('click',event=>{if(event.target===overlay)hideDialog();});
 dialog.addEventListener('keydown',event=>{if(event.key==='Escape'){hideDialog();return;}
  if(event.key==='Tab'){
   const controls=[close,token,submit].filter(x=>!x.disabled);
   const first=controls[0],last=controls[controls.length-1];
   if(event.shiftKey&&document.activeElement===first){event.preventDefault();last.focus();}
   else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus();}
  }
 });
 // Browser navigation and Safari page restoration can encounter a brief
 // unreachable tunnel. Only an authoritative 401/403 establishes sign-out.
 let refreshPending=null;
 async function refreshSession(){
  const request=async(headers,phase)=>{
   const response=await fetch('/api/session',{credentials:'same-origin',cache:'no-store',headers:tracedHeaders(headers)});
   traceClient('refresh_response',phase,response.status);
   if(response.ok)return {kind:'valid',session:await response.json()};
   if(response.status===401||response.status===403)return {kind:'invalid'};
   return {kind:'unavailable'};
  };
  try{
   traceClient('refresh_begin','session');
   const first=await request(homeSessionHeaders(),'session');
   if(first.kind==='valid'||first.kind==='unavailable')return first;
   // Recover from a stale/missing browser-side session ID using the
   // independent, host-only Homepage cookie before declaring sign-out.
   if(homeSessionToken)return await request({},'cookie_fallback');
   return first;
  }catch(_){traceClient('refresh_error','network');return {kind:'unavailable'};}
 }
 function showUnavailable(){
  traceClient('ui_unavailable','unavailable');
  signedIn.hidden=true;signIn.hidden=true;guestButton.hidden=true;retryAuth.hidden=false;
 }
 function showSignedOut(){
  traceClient('ui_signed_out','invalid');
  clearHomeLogin();
  signedIn.hidden=true;signIn.hidden=false;guestButton.hidden=false;retryAuth.hidden=true;manage.hidden=true;
  suiteHandoffToken='';suiteHandoffExpires=0;
  try{sessionStorage.removeItem('suite-handoff-v1')}catch(_){}
  applyHandoffLinks();
 }
 const refresh=()=>{
  if(refreshPending)return refreshPending;
  refreshPending=(async()=>{
   const state=await refreshSession();
   if(state.kind==='valid'){
    try{rememberHomeLogin(state.session)}catch(_){showUnavailable();return false;}
    identity.textContent=state.session.user+' ('+state.session.role+')';
    manage.hidden=state.session.role==='guest'; // Every signed-in user except temporary guests may manage tokens
    if(state.session.role!=='guest'){manage.hidden=false; // Every signed-in user may manage tokens
    }
    signedIn.hidden=false;signIn.hidden=state.session.role!=='guest';guestButton.hidden=true;retryAuth.hidden=true;
    // Returning from any suite app is also a verified Homepage arrival.
    void acknowledgeSuiteLaunch('homepage');
    traceClient('ui_signed_in','valid');
    return true;
   }
   if(state.kind==='invalid'){showSignedOut();return false;}
   showUnavailable();
   return false;
  })().finally(()=>{refreshPending=null;});
  return refreshPending;
 };
 guestButton.addEventListener('click',async()=>{
  if(guestButton.disabled)return;
  guestButton.disabled=true;
  guestButton.textContent='Creating guest session…';
  try{
   const response=await fetch('/api/session/guest',{method:'POST',credentials:'same-origin',cache:'no-store'});
   if(!response.ok)throw Error('Guest access unavailable (HTTP '+response.status+').');
   rememberHomeLogin(await response.json());
   if(!await refresh())throw Error('Guest session could not be verified.');
  }catch(error){traceClient('guest_signin_error','failed');alert(error.message||'Guest access unavailable.');}
  finally{guestButton.disabled=false;guestButton.textContent='Sign in as Guest';}
 });
 retryAuth.addEventListener('click',()=>{void refresh()});
 form.addEventListener('submit',async event=>{
  event.preventDefault();
  if(signInPending)return;
  signInPending=true;submit.disabled=true;submit.textContent='Signing in…';
  submit.setAttribute('aria-busy','true');token.disabled=true;close.disabled=true;
  message.textContent='Checking your personal token…';
  const approvedToken=token.value.trim();
  const controller=new AbortController();
  const timeout=setTimeout(()=>controller.abort(),15000);
  let authenticated=false;
  try{
   const response=await fetch('/api/session',{method:'POST',credentials:'same-origin',headers:tracedHeaders({Authorization:'Bearer '+approvedToken}),cache:'no-store',signal:controller.signal});
   clearTimeout(timeout);
   traceClient('signin_response',response.ok?'success':'rejected',response.status);
   if(response.status===401||response.status===403)throw Error('Personal token rejected or expired. Check your token and try again.');
   if(!response.ok)throw Error('Sign in failed (HTTP '+response.status+'). Please try again.');
   message.textContent='Token accepted. Preparing application access…';
   rememberHomeLogin(await response.json());
   if(!await refresh())throw Error('Homepage sign-in session was not retained. Please sign in again.');
   authenticated=true;
   message.textContent='Signed in successfully.';
  }catch(error){
   traceClient('signin_error',error.name==='AbortError'?'network':'rejected');
   message.textContent=error.name==='AbortError'?'Sign in timed out. Check your connection and try again.':error instanceof TypeError?'Network error. Check your connection and try again.':error.message;
  }finally{
   clearTimeout(timeout);
   signInPending=false;submit.disabled=false;submit.textContent='Sign in';
   submit.removeAttribute('aria-busy');token.disabled=false;close.disabled=false;
   if(authenticated)hideDialog();
  }
 });
 // Validate after ordinary navigation, BFCache restoration and reconnect;
 // never couple Homepage validation to the expiry of an app handoff.
 void void refresh();
 window.addEventListener('focus',()=>{traceClient('page_focus','initial');void refresh()});
 window.addEventListener('pageshow',()=>{traceClient('page_show','initial');void refresh()});
 window.addEventListener('online',()=>{traceClient('page_online','initial');void refresh()});
 document.addEventListener('visibilitychange',()=>{if(document.visibilityState==='visible'){traceClient('page_show','initial');void refresh()}});

}
