'use strict';
// Behavioral tests for fixed-origin, fail-closed Homepage recovery.
// Run from Homepage root: node scripts/test-auth-recovery.cjs
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const script=fs.readFileSync('static/launch-suite.js','utf8');
assert.ok(!script.includes('suite_mode=local'),'no anonymous app redirects');
assert.ok(!script.includes('localLink.href'),'no unauthenticated continuation link');
const mockedToken='a'.repeat(64);
async function run({path='/launch/stocks',stored='',statuses=[401],token=mockedToken,expiration=Date.now()+60_000}={}){
 let calls=[],redirects=[],message='';
 const statusElement={set textContent(v){message=v},get textContent(){return message}};
 const ctx={
  location:{pathname:path,replace:url=>redirects.push(url)},
  localStorage:{getItem:key=>key==='themauris-theme'?'dark':key==='mauri-home-session-v1'?(stored?JSON.stringify({token:stored}):null):null},
  document:{documentElement:{dataset:{}},getElementById:id=>id==='status'?statusElement:null},
  fetch:async (url,options)=>{
   calls.push({url,options});
   const status=statuses[Math.min(calls.length-1,statuses.length-1)];
   return {status,ok:status===200,json:async()=>({token,expiresAt:new Date(expiration).toISOString()})};
  },
  Date,Error,JSON,Object,RegExp
 };
 const text=script.replace('void recoverSuite();','globalThis.__recover=recoverSuite;');
 vm.runInNewContext(text,ctx);
 await ctx.__recover();
 return {calls,redirects,message};
}
(async()=>{
 const unknown=await run({path:'/launch/evil'});
 assert.equal(unknown.calls.length,0);
 assert.match(unknown.message,/Unknown suite application/);
 const reject=await run({statuses:[401]});
 assert.deepEqual(reject.redirects,['https://home.themauris.org/#suite_signin=required']);
 assert.equal(reject.calls.length,1);
 const forbidden=await run({statuses:[403]});
 assert.deepEqual(forbidden.redirects,['https://home.themauris.org/#suite_signin=required']);
 const stale=await run({stored:'b'.repeat(64),statuses:[401,200]});
 assert.equal(stale.calls.length,2,'retry stale parent ID using host-only cookie');
 assert.equal(stale.calls[0].options.headers.Authorization,'Session '+'b'.repeat(64));
 assert.equal(Object.keys(stale.calls[1].options.headers).length,0);
 assert.deepEqual(stale.redirects,['https://stocks.themauris.org/#suite_token='+mockedToken]);
 const good=await run({path:'/launch/newsgo',statuses:[200]});
 assert.deepEqual(good.redirects,['https://newsgo.themauris.org/#suite_token='+mockedToken]);
 const outage=await run({statuses:[503]});
 assert.equal(outage.redirects.length,0,'server outage cannot authorize anonymous entry');
 assert.match(outage.message,/Could not verify/);
 const badResponse=await run({statuses:[200],token:'untrusted'});
 assert.equal(badResponse.redirects.length,0,'invalid bearer must not launch app');
 const expired=await run({statuses:[200],expiration:Date.now()-1000});
 assert.equal(expired.redirects.length,0,'expired bearer must not launch app');
 for(const path of ['/launch/newsgo','/launch/stocks','/launch/sailing','/launch/scrapergo']){
  const result=await run({path,statuses:[401]});
  assert.deepEqual(result.redirects,['https://home.themauris.org/#suite_signin=required']);
 }
 const html=fs.readFileSync('static/launch-suite.html','utf8');
 assert.ok(!html.includes('Continue without an account'));
 assert.ok(html.includes('Return to Homepage to sign in or use Guest'));
 console.log('Homepage recovery: personal/guest-only redirects, failures and trusted handoffs passed.');
})().catch(e=>{console.error(e);process.exitCode=1});
