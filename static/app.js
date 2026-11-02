const icons = {
 sail:'<path d="M12 3v19M12 5 4 17h8M14 8l7 9h-7M3 21c2-1 4-1 6 0s4 1 6 0 4-1 6 0"/>',
 chart:'<path d="M3 3v18h18M7 16l4-5 4 3 6-8"/><path d="M17 6h4v4"/>',
 news:'<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 7h8M8 11h8M8 15h3M14 15h2"/>',
 code:'<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m9 9-3 3 3 3m6-6 3 3-3 3m-4 5 2-10"/>'
};
const accent={cyan:['#68cde0','rgba(104,205,224,.12)'],mint:['#75ddb6','rgba(117,221,182,.12)'],amber:['#eac081','rgba(234,192,129,.12)'],violet:['#b7a1f6','rgba(183,161,246,.12)']};
function svg(name){return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[name]||icons.code}</svg>`}
function makeCard(app){const a=document.createElement('a');a.className='app-card';a.href=app.url;a.rel='noopener noreferrer';if(app.url.startsWith('https://')||app.url.startsWith('http://'))a.target='_blank';const [color,glow]=accent[app.accent]||accent.cyan;a.style.setProperty('--card-accent',color);a.style.setProperty('--card-glow',glow);const top=document.createElement('div');top.className='card-top';const icon=document.createElement('span');icon.className='card-icon';icon.innerHTML=svg(app.icon);const arrow=document.createElement('span');arrow.className='card-arrow';arrow.textContent='↗';top.append(icon,arrow);const content=document.createElement('div');content.className='card-content';for(const [cls,val,tag] of [['card-category',app.category,'div'],['card-name',app.name,'h3'],['card-subtitle',app.subtitle,'div'],['card-description',app.description,'p']]){let el=document.createElement(tag);el.className=cls;el.textContent=val;content.append(el)}const bottom=document.createElement('div');bottom.className='card-bottom';bottom.innerHTML='<span>OPEN APPLICATION</span><span class="direction">→</span>';a.append(top,content,bottom);return a}
async function init(){const root=document.querySelector('#cards');try{const r=await fetch('/api/config',{cache:'no-store'});if(!r.ok)throw Error('Configuration unavailable');const c=await r.json();root.replaceChildren(...c.apps.map(makeCard));document.querySelector('.count').textContent='/ '+String(c.apps.length).padStart(2,'0');if(c.persistenceAvailable)enableOptionalAuth()}catch(e){root.textContent='Applications are temporarily unavailable. Please refresh to try again.'}}
function tick(){document.querySelector('#clock').textContent=new Intl.DateTimeFormat('en-US',{hour:'2-digit',minute:'2-digit',hour12:false,timeZone:'America/Los_Angeles'}).format(new Date())}
const themeKey='themauris-theme';try{if(localStorage.getItem(themeKey)==='light')document.documentElement.dataset.theme='light'}catch(_){}document.querySelector('#themeToggle').addEventListener('click',()=>{const next=document.documentElement.dataset.theme==='light'?'dark':'light';document.documentElement.dataset.theme=next;try{localStorage.setItem(themeKey,next)}catch(_){}});document.querySelector('#year').textContent=new Date().getFullYear();tick();setInterval(tick,30_000);init();

/* Optional suite identity. Existing application browser storage remains the
   default, and no local settings are uploaded or replaced by signing in. */
function enableOptionalAuth(){
 if(document.querySelector('#suiteIdentity'))return;
 const section=document.createElement('section');
 section.id='suiteIdentity';section.className='suite-identity';
 const heading=document.createElement('div');heading.className='suite-identity-head';
 const title=document.createElement('strong');title.textContent='Personal settings';
 const status=document.createElement('span');status.id='suiteIdentityStatus';status.textContent='Browser-only mode';
 heading.append(title,status);
 const description=document.createElement('p');description.textContent='Without a token, your applications continue using their existing browser storage. Sign in to enable your private server account; synchronizing application settings will become available as each app is connected.';
 const form=document.createElement('form');form.id='suiteLogin';
 const label=document.createElement('label');label.htmlFor='suiteToken';label.textContent='Approved personal token';
 const token=document.createElement('input');token.id='suiteToken';token.type='password';token.autocomplete='off';token.spellcheck=false;token.required=true;token.minLength=64;token.maxLength=64;
 const signIn=document.createElement('button');signIn.type='submit';signIn.textContent='Sign in';
 form.append(label,token,signIn);
 const signedIn=document.createElement('div');signedIn.id='suiteSignedIn';signedIn.hidden=true;
 const identity=document.createElement('span');identity.id='suiteIdentityUser';
 const logout=document.createElement('button');logout.type='button';logout.textContent='Sign out';
 const admin=document.createElement('a');admin.href='/suite-admin';admin.textContent='Manage tokens';admin.hidden=true;
 signedIn.append(identity,logout,admin);
 const message=document.createElement('p');message.id='suiteIdentityMessage';message.setAttribute('role','status');message.setAttribute('aria-live','polite');
 section.append(heading,description,form,signedIn,message);
 document.querySelector('.topbar').insertAdjacentElement('afterend',section);
 const refresh=async()=>{
  try{
   const response=await fetch('/api/session',{credentials:'same-origin',cache:'no-store'});
   if(!response.ok)throw new Error('no session');
   const data=await response.json();
   status.textContent='Signed in · server account';
   identity.textContent=data.user+' ('+data.role+')';
   admin.hidden=data.role!=='admin';signedIn.hidden=false;form.hidden=true;
  }catch(_){
   status.textContent='Browser-only mode';signedIn.hidden=true;form.hidden=false;
  }
 };
 form.addEventListener('submit',async event=>{
  event.preventDefault();signIn.disabled=true;message.textContent='';
  try{
   const response=await fetch('/api/session',{method:'POST',headers:{Authorization:'Bearer '+token.value.trim()},credentials:'same-origin',cache:'no-store'});
   if(!response.ok)throw new Error('Token not approved (HTTP '+response.status+').');
   token.value='';
   await refresh();
   if(signedIn.hidden)throw new Error('Session not retained. Use HTTPS for the public site.');
   message.textContent='Signed in. Existing local settings have not been changed.';
  }catch(error){message.textContent=error.message;}
  finally{signIn.disabled=false;}
 });
 logout.addEventListener('click',async()=>{
  try{
   const response=await fetch('/api/session',{method:'DELETE',credentials:'same-origin',cache:'no-store'});
   if(!response.ok)throw new Error('Could not sign out.');
   message.textContent='Signed out. Browser settings were not deleted.';
   await refresh();
  }catch(error){message.textContent=error.message;}
 });
 refresh();
}
