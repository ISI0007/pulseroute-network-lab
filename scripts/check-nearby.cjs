// A real Go HTTP server supplies the URL through a private, test-only environment.
const path=require('node:path'),fs=require('node:fs'),assert=require('node:assert/strict');
const {chromium}=require(path.resolve(__dirname,'../frontend/node_modules/playwright'));
(async()=>{
 const url=process.env.PULSEROUTE_TEST_URL;
 assert.ok(url,'Run through the opt-in Go integration test');
 const browser=await chromium.launch({executablePath:process.env.CHROME_PATH||'C:/Program Files/Google/Chrome/Application/chrome.exe',headless:true,chromiumSandbox:true});
 try{
  const context=await browser.newContext({acceptDownloads:true,viewport:{width:390,height:844}});
  const page=await context.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
  await page.goto(url);await page.getByText('Connected to the temporary nearby session.',{exact:true}).waitFor();
  assert.equal(new URL(page.url()).hash,'','Pairing token remained in address bar');
  for(const width of [320,390,768,1440]){await page.setViewportSize({width,height:900});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth)<=width,'Portal overflow at '+width);}
  const downloadPromise=page.waitForEvent('download');await page.getByRole('link',{name:'Download offered.txt',exact:true}).click();
  const download=await downloadPromise;assert.equal(await download.failure(),null);assert.equal(fs.readFileSync(await download.path(),'utf8'),'PulseRoute browser download fixture\n');
  await page.locator('#upload-files').setInputFiles({name:'uploaded.txt',mimeType:'text/plain',buffer:Buffer.from('PulseRoute browser upload fixture\n')});
  await page.getByRole('button',{name:'Send selected files',exact:true}).click();
  await page.getByText('Received; SHA-256 matched the sender.',{exact:false}).waitFor();
  await page.getByText('Queue finished. Review each file result above.',{exact:true}).waitFor();
  await page.locator('#upload-files').setInputFiles({name:'corrupt.txt',mimeType:'text/plain',buffer:Buffer.from('bad digest fixture')});
  await page.locator('#expected-hash').fill('0'.repeat(64));
  await page.getByRole('button',{name:'Send selected files',exact:true}).click();
  await page.locator('#results .error').first().waitFor();
  const fresh=await browser.newContext();const unpaired=await fresh.newPage();await unpaired.goto(new URL('/',url).href);
  await unpaired.getByText('Open the full QR pairing link from the sender. This page has no access token.',{exact:true}).waitFor();
  assert.equal(await unpaired.locator('#workspace').isVisible(),false);
  await fresh.close();assert.deepEqual(errors,[]);
  console.log('PASS: real HTTP portal UI, exact downloaded bytes, verified upload, corrupt digest rejection, token removal and unpaired access; four viewport widths. Chromium only; no physical Android/iOS claim.');
 }finally{await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
