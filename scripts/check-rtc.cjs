// Real WebRTC transport with explicitly synthetic tracks, not a hardware test.
const path = require('node:path');
const {pathToFileURL} = require('node:url');
const assert = require('node:assert/strict');
const root=path.resolve(__dirname,'../frontend');
const {chromium}=require(path.join(root,'node_modules/playwright'));
(async()=>{
 const {createServer}=await import(pathToFileURL(path.join(root,'node_modules/vite/dist/node/index.js')).href);
 const server=await createServer({root,server:{host:'127.0.0.1',port:0}});
 await server.listen(); let browser;
 try {
  browser=await chromium.launch({executablePath:process.env.CHROME_PATH||'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true,chromiumSandbox:true});
  const page=await browser.newPage();
  await page.goto('http://127.0.0.1:'+server.httpServer.address().port);
  const result=await page.evaluate(async()=>{
   const {LanCallController}=await import('/src/rtc.ts');
   const states=[{},{}],locals=[null,null],remote=[null,null],acquisitions=[0,0],controllers=[],tracks=[],timers=[],contexts=[];
   const ids=['a'.repeat(32),'b'.repeat(32)];let seq=0;
   const check=(condition,message)=>{if(!condition)throw Error(message)};
   const wait=async(predicate)=>{const deadline=Date.now()+15000;while(!predicate()){if(Date.now()>deadline)throw Error('Timed out: '+JSON.stringify(states));await new Promise(r=>setTimeout(r,50));}};
   try {
    for(let i=0;i<2;i++)controllers[i]=new LanCallController({
     send:async(to,id,signal)=>{check(to===ids[1-i],'Wrong recipient');controllers[1-i].receive({seq:++seq,from:ids[i],peer:'fixture '+i,id,text:JSON.stringify(signal)});},
     onState:s=>states[i]=s,onLocalStream:s=>locals[i]=s,onRemoteStream:s=>remote[i]=s,
     acquireMedia:async constraints=>{
      acquisitions[i]++;const stream=new MediaStream();
      if(constraints.audio){const ctx=new AudioContext();contexts.push(ctx);await ctx.resume();const tone=ctx.createOscillator(),out=ctx.createMediaStreamDestination();tone.connect(out);tone.start();stream.addTrack(out.stream.getAudioTracks()[0]);}
      if(constraints.video){const canvas=document.createElement('canvas');canvas.width=160;canvas.height=120;const ctx=canvas.getContext('2d');let frame=0;timers.push(setInterval(()=>{ctx.fillStyle=frame++%2?'red':'blue';ctx.fillRect(0,0,160,120)},60));stream.addTrack(canvas.captureStream(12).getVideoTracks()[0]);}
      tracks.push(...stream.getTracks());return stream;
     }
    });
    await controllers[0].start(ids[1],'fixture 1','video');await wait(()=>states[1].phase==='incoming');check(acquisitions[1]===0,'Incoming call captured media before consent');
    await controllers[1].accept();await wait(()=>states.every(s=>s.phase==='connected'));
    await wait(()=>remote.every(s=>s&&s.getAudioTracks().length&&s.getVideoTracks().length));
    let stats;
    for(let n=0;n<100;n++){stats=await controllers[0].stats();if(stats&&stats.bytesSent>0&&stats.bytesReceived>0)break;await new Promise(r=>setTimeout(r,100));}
    check(stats.bytesSent>0&&stats.bytesReceived>0,'No bidirectional media bytes');
    controllers[0].mute();check(locals[0].getAudioTracks().every(t=>!t.enabled),'Mute did not disable microphone track');
    const old=locals[0].getVideoTracks()[0];await controllers[0].toggleCamera();check(old.readyState==='ended','Camera off did not stop track');await controllers[0].toggleCamera();check(states[0].cameraEnabled,'Camera did not resume');
    controllers[0].hangup();await wait(()=>states.every(s=>s.phase==='ended'));check(tracks.every(t=>t.readyState==='ended'),'Hangup leaked capture tracks');
    const before=acquisitions[1];await controllers[0].start(ids[1],'fixture 1','audio');await wait(()=>states[1].phase==='incoming');controllers[1].reject();await wait(()=>states[0].phase==='ended');check(acquisitions[1]===before,'Reject captured media');
    return {status:'PASS',scope:'Real browser WebRTC with synthetic audio/video; fixture signaling only; no camera, microphone, native bridge or physical device claim',media:stats,consent:true,mute:true,cameraRestart:true,hangupCleanup:true,rejection:true};
   }finally{controllers.forEach(c=>c.destroy());timers.forEach(clearInterval);await Promise.all(contexts.map(c=>c.close()));}
  });
  assert.equal(result.status,'PASS');console.log(JSON.stringify(result,null,2));
 }finally{await browser?.close();await server.close();}
})().catch(error=>{console.error(error);process.exitCode=1});
