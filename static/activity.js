// Homepage-only anonymous/personal presence. The identifier is temporary,
// origin-scoped and never sent to sister applications or retained by the server.
(() => {
  const key='homepage-activity-page-v1';
  let id='';
  try {
    id=sessionStorage.getItem(key)||'';
    if (!/^[a-f0-9]{32}$/.test(id)) {
      id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
      sessionStorage.setItem(key,id);
    }
  } catch (_) {
    id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
  }
  let pendingInteraction=false;
  let lastSent=0;
  function send() {
    if (document.visibilityState!=='visible') return;
    const now=Date.now();
    if (now-lastSent<25000)return;
    lastSent=now;
    const body=JSON.stringify({id,visible:true,interacted:pendingInteraction});
    pendingInteraction=false;
    try {
      const headers={'Content-Type':'application/json'};
      // Suite Management retains an ephemeral session ID in memory for
      // loopback Safari. Use it only for same-origin presence authentication.
      if (typeof sessionToken==='string' && /^[a-f0-9]{64}$/.test(sessionToken)) {
        headers.Authorization='Session '+sessionToken;
      }
      fetch('/api/activity/ping',{
        method:'POST',credentials:'same-origin',cache:'no-store',
        headers,body,keepalive:true
      }).catch(()=>{});
    } catch (_) {}
  }
  function interaction() {
    pendingInteraction=true;
    send();
  }
  document.addEventListener('pointerdown',interaction,{passive:true});
  document.addEventListener('keydown',interaction,{passive:true});
  document.addEventListener('visibilitychange',()=>{if(!document.hidden)send()});
  window.addEventListener('pageshow',send);
  window.addEventListener('focus',send);
  send();
  setInterval(send,60000);
})();
