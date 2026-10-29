// Decode a real QR bitmap containing only an explicitly synthetic session URL.
const path=require('node:path');
const assert=require('node:assert/strict');
const modules=path.resolve(__dirname,'../frontend/node_modules');
const qr=require(path.join(modules,'qrcode'));
const jsqr=require(path.join(modules,'jsqr'));
const {PNG}=require(path.join(modules,'pngjs'));
(async()=>{
 const text='http://192.0.2.1:42193/#'+'a'.repeat(64);
 const png=PNG.sync.read(await qr.toBuffer(text,{width:512,margin:4,errorCorrectionLevel:'M'}));
 const decoded=jsqr(new Uint8ClampedArray(png.data),png.width,png.height);
 assert.ok(decoded,'QR did not decode');assert.equal(decoded.data,text);
 console.log('PASS: QR encode/decode preserved the synthetic sharing URL exactly; physical scanning not tested');
})().catch(e=>{console.error(e);process.exitCode=1});
