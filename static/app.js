const icons = {
 sail:'<path d="M12 3v19M12 5 4 17h8M14 8l7 9h-7M3 21c2-1 4-1 6 0s4 1 6 0 4-1 6 0"/>',
 chart:'<path d="M3 3v18h18M7 16l4-5 4 3 6-8"/><path d="M17 6h4v4"/>',
 news:'<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 7h8M8 11h8M8 15h3M14 15h2"/>',
 code:'<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m9 9-3 3 3 3m6-6 3 3-3 3m-4 5 2-10"/>'
};
const accent={cyan:['#68cde0','rgba(104,205,224,.12)'],mint:['#75ddb6','rgba(117,221,182,.12)'],amber:['#eac081','rgba(234,192,129,.12)'],violet:['#b7a1f6','rgba(183,161,246,.12)']};
function svg(name){return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[name]||icons.code}</svg>`}
// Stable named browsing contexts let a Homepage-launched app tab be reused.
// Rel=noopener is kept for the browser's native fallback; the normal launch
// explicitly clears opener before navigating the new/reused window.
function appTabName(address){
 const url=new URL(address);
 return 'mauri_app_'+url.hostname.replace(/[^a-z0-9]/gi,'_').toLowerCase();
}
// A named window is not always rediscoverable in Safari after its opener is
// cleared. Keep the WindowProxy in this Homepage tab and focus it on reuse.
// References live only in page memory; a Homepage reload may lose them.
const appTabHandles=new Map();
function existingAppTab(name){
 const tab=appTabHandles.get(name);
 if(!tab)return null;
 try{
  if(!tab.closed){tab.focus();return tab}
 }catch(_){}
 appTabHandles.delete(name);
 return null;
}
function makeCard(app){const a=document.createElement('a');a.className='app-card';a.href=app.url;a.rel='noopener noreferrer';if(app.url.startsWith('https://')||app.url.startsWith('http://'))a.target=appTabName(app.url);const [color,glow]=accent[app.accent]||accent.cyan;a.style.setProperty('--card-accent',color);a.style.setProperty('--card-glow',glow);const top=document.createElement('div');top.className='card-top';const icon=document.createElement('span');icon.className='card-icon';icon.innerHTML=svg(app.icon);const arrow=document.createElement('span');arrow.className='card-arrow';arrow.textContent='↗';top.append(icon,arrow);const content=document.createElement('div');content.className='card-content';for(const [cls,val,tag] of [['card-category',app.category,'div'],['card-name',app.name,'h3'],['card-subtitle',app.subtitle,'div'],['card-description',app.description,'p']]){let el=document.createElement(tag);el.className=cls;el.textContent=val;content.append(el)}const bottom=document.createElement('div');bottom.className='card-bottom';bottom.innerHTML='<span>OPEN APPLICATION</span><span class="direction">→</span>';a.append(top,content,bottom);return a}
// Handoff tokens are kept only in memory and put in URL fragments for
// approved suite destinations. The durable sign-in token is never forwarded.
// Homepage has its own twelve-hour sign-in credential, distinct from
// the short-lived app token. Store only this temporary session ID, never
// the approved personal sign-in token.
const homeLoginKey='mauri-home-session-v1';
let homeSessionToken='',homeSessionExpires=0;
try{
 const session=JSON.parse(localStorage.getItem(homeLoginKey)||'null');
 if(session && /^[a-f0-9]{64}$/.test(session.token) && session.expires>Date.now()){
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
 // Never leave a cached handoff credential on a card. Only attach a newly
 // issued token when actually launching from a signed-in Homepage session.
 for(const link of document.querySelectorAll('#cards a.app-card')){
  const url=new URL(link.href);
  if(url.origin!=='https://newsgo.themauris.org' && url.origin!=='https://stocks.themauris.org')continue;
  url.hash='';
  link.href=url.toString();
 }
}
async function renewHandoff(approvedToken=''){
 try{
  const response=await fetch('/api/handoff/token',{method:'POST',credentials:'same-origin',cache:'no-store',headers:approvedToken?{Authorization:'Bearer '+approvedToken}:homeSessionHeaders()});
  if(!response.ok)throw Error('not authenticated');
  const data=await response.json();
  if(!/^[a-f0-9]{64}$/.test(data.token)||!Date.parse(data.expiresAt))throw Error('bad handoff');
  suiteHandoffToken=data.token;suiteHandoffExpires=Date.parse(data.expiresAt);
  try{sessionStorage.setItem('suite-handoff-v1',JSON.stringify({token:suiteHandoffToken,expires:suiteHandoffExpires}))}catch(_){}
  applyHandoffLinks();return true;
 }catch(_){suiteHandoffToken='';suiteHandoffExpires=0;try{sessionStorage.removeItem('suite-handoff-v1')}catch(_){}applyHandoffLinks();return false}
}
// Open the named app tab synchronously during the click. Safari can block a
// popup created after awaiting the signed-in handoff request.
async function openSuiteApp(event){
 if(event.defaultPrevented || event.button!==0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)return;
 const link=event.currentTarget;
 if(link.dataset.opening==='true'){event.preventDefault();return;}
 const url=new URL(link.href);
 const isSuite=url.origin==='https://newsgo.themauris.org' || url.origin==='https://stocks.themauris.org';
 const tabName=appTabName(url.toString());
 // No token is ever attached to the persistent card href. Hand-off tokens
 // are created afresh only for signed-in suite launches.
 if(isSuite)url.hash='';
 // Prefer the actual handle to an existing app tab. Looking up the
 // named browsing context alone can create a duplicate in Safari.
 if(existingAppTab(tabName)){event.preventDefault();return;}
 const opened=window.open('',tabName);
 if(!opened)return; // Native link with noopener remains the fallback.
 event.preventDefault();
 appTabHandles.set(tabName,opened);
 try{
  // Initial about:blank inherits our origin, allowing removal of the
  // reverse opener reference before the application is navigated.
  try{if(opened.location.href==='about:blank')opened.opener=null}catch(_){}
  if(isSuite && homeSessionToken && Date.now()<homeSessionExpires){
   link.dataset.opening='true';
   const issued=await renewHandoff();
   if(issued && suiteHandoffToken && Date.now()<suiteHandoffExpires){
    url.hash='suite_token='+encodeURIComponent(suiteHandoffToken);
   }
  }
  // Assigning location through the preopened handle works after await,
  // even when the destination is cross-origin. Homepage stays in its tab.
  opened.location.href=url.toString();
  opened.focus();
 }catch(_){
  // Do not replace the Homepage tab or leak credentials on navigation errors.
  // User may retry the ordinary card link.
 }finally{
  delete link.dataset.opening;
 }
}
async function init(){const root=document.querySelector('#cards');try{const r=await fetch('/api/config',{cache:'no-store'});if(!r.ok)throw Error('Configuration unavailable');const c=await r.json();root.replaceChildren(...c.apps.map(makeCard));applyHandoffLinks();for(const link of root.querySelectorAll('a.app-card'))link.addEventListener('click',openSuiteApp);document.querySelector('.count').textContent='/ '+String(c.apps.length).padStart(2,'0');if(c.persistenceAvailable)enableOptionalAuth()}catch(e){root.textContent='Applications are temporarily unavailable. Please refresh to try again.'}}
function tick(){document.querySelector('#clock').textContent=new Intl.DateTimeFormat('en-US',{hour:'2-digit',minute:'2-digit',hour12:false,timeZone:'America/Los_Angeles'}).format(new Date())}
const themeKey='themauris-theme';try{if(localStorage.getItem(themeKey)==='light')document.documentElement.dataset.theme='light'}catch(_){}document.querySelector('#themeToggle').addEventListener('click',()=>{const next=document.documentElement.dataset.theme==='light'?'dark':'light';document.documentElement.dataset.theme=next;try{localStorage.setItem(themeKey,next)}catch(_){}});document.querySelector('#year').textContent=new Date().getFullYear();tick();setInterval(tick,30_000);init();

/* Optional account control: browser-only remains the default.
   Signing in does not upload or replace application localStorage. */
function enableOptionalAuth(){
 if(document.querySelector('#suiteAccount'))return;
 const top=document.querySelector('.topright');
 const container=document.createElement('div');container.id='suiteAccount';container.className='suite-account';
 const signIn=document.createElement('button');signIn.type='button';signIn.textContent='Sign in';
 const signedIn=document.createElement('div');signedIn.className='suite-account-auth';signedIn.hidden=true;
 const identity=document.createElement('span');
 const manage=document.createElement('a');manage.href='/suite-admin';manage.textContent='Manage tokens';manage.hidden=true;
 const signOut=document.createElement('button');signOut.type='button';signOut.textContent='Sign out';
 signedIn.append(identity,manage,signOut);
 container.append(signIn,signedIn);top.prepend(container);
 const overlay=document.createElement('div');overlay.className='suite-modal-backdrop';overlay.hidden=true;
 const dialog=document.createElement('section');dialog.className='suite-modal';dialog.setAttribute('role','dialog');dialog.setAttribute('aria-modal','true');dialog.setAttribute('aria-labelledby','suiteDialogTitle');
 const header=document.createElement('div');header.className='suite-modal-header';
 const title=document.createElement('h2');title.id='suiteDialogTitle';title.textContent='Personal settings';
 const close=document.createElement('button');close.type='button';close.textContent='Close';close.setAttribute('aria-label','Close sign-in dialog');
 header.append(title,close);
 const description=document.createElement('p');description.textContent='Sign in with your approved personal token to use your server account. Without signing in, applications continue using browser storage. Existing local settings will not be changed.';
 const form=document.createElement('form');
 const label=document.createElement('label');label.htmlFor='suiteToken';label.textContent='Approved personal token';
 const token=document.createElement('input');token.id='suiteToken';token.type='password';token.autocomplete='off';token.spellcheck=false;token.required=true;token.minLength=64;token.maxLength=64;
 const submit=document.createElement('button');submit.type='submit';submit.textContent='Sign in';
 form.append(label,token,submit);
 const message=document.createElement('p');message.className='suite-modal-message';message.setAttribute('role','status');message.setAttribute('aria-live','polite');
 dialog.append(header,description,form,message);overlay.append(dialog);document.body.append(overlay);
 let previousFocus=null;
 function hideDialog(){overlay.hidden=true;token.value='';message.textContent='';if(previousFocus)previousFocus.focus();}
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
 const refresh=async()=>{
  try{
   const response=await fetch('/api/session',{credentials:'same-origin',cache:'no-store',headers:homeSessionHeaders()});
   // A short-lived app bearer is not a Homepage login. Never display
   // a signed-in Homepage identity unless the 12-hour session is valid.
   if(!response.ok)throw Error('Homepage login session unavailable');
   const session=await response.json();
   rememberHomeLogin(session);
   identity.textContent=session.user+' ('+session.role+')';
   manage.hidden=session.role!=='admin';
   signedIn.hidden=false;signIn.hidden=true;
   await renewHandoff();return true;
  }catch(_){signedIn.hidden=true;signIn.hidden=false;manage.hidden=true;suiteHandoffToken='';suiteHandoffExpires=0;try{sessionStorage.removeItem('suite-handoff-v1')}catch(_){}applyHandoffLinks();return false;}
 };
 form.addEventListener('submit',async event=>{
  event.preventDefault();submit.disabled=true;message.textContent='';
  try{
   const response=await fetch('/api/session',{method:'POST',credentials:'same-origin',headers:{Authorization:'Bearer '+token.value.trim()},cache:'no-store'});
   if(!response.ok)throw Error('Token not approved (HTTP '+response.status+').');
   rememberHomeLogin(await response.json());
   await renewHandoff(token.value.trim());
   if(!suiteHandoffToken)throw Error('Unable to issue one-hour app access.');
   if(!await refresh())throw Error('Homepage sign-in session was not retained. Please enable cookies for home.themauris.org and sign in again.');
   hideDialog();
  }catch(error){message.textContent=error.message;}
  finally{submit.disabled=false;}
 });
 signOut.addEventListener('click',async()=>{
  signOut.disabled=true;
  try{
   const response=await fetch('/api/session',{method:'DELETE',credentials:'same-origin',cache:'no-store',headers:homeSessionHeaders()});
   if(!response.ok)throw Error('Sign out failed (HTTP '+response.status+').');
   clearHomeLogin();suiteHandoffToken='';suiteHandoffExpires=0;try{sessionStorage.removeItem('suite-handoff-v1')}catch(_){}
   applyHandoffLinks();await refresh();
  }catch(error){alert(error.message);}
  finally{signOut.disabled=false;}
 });
 refresh();
 window.addEventListener('focus',()=>{if(Date.now()>suiteHandoffExpires-5*60*1000)void refresh()});
}
