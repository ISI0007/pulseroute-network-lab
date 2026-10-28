// Production frontend QA in installed Chrome; no mocked Go bridge or network data.
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const pages = ['Overview', 'Messages', 'Transfers', 'Network Lab', 'History', 'Settings'];
const themes = [['forest', 'Forest'], ['dark', 'Midnight'], ['light', 'Daylight'], ['ocean', 'Ocean']];
const rgb = value => value.match(/[\d.]+/g).slice(0, 3).map(Number);
const luminance = value => rgb(value).map(n => { const c = n / 255; return c <= .04045 ? c / 12.92 : ((c + .055) / 1.055) ** 2.4; }).reduce((sum, c, i) => sum + c * [.2126, .7152, .0722][i], 0);
const contrast = (a, b) => { const x = luminance(a), y = luminance(b); return (Math.max(x, y) + .05) / (Math.min(x, y) + .05); };
(async () => {
  const root = path.resolve(__dirname, '../frontend/dist');
  const out = path.resolve(__dirname, '../docs/evidence');
  fs.mkdirSync(out, { recursive: true });
  const server = http.createServer((req, res) => {
    let file = path.resolve(root, '.' + decodeURIComponent(new URL(req.url, 'http://localhost').pathname));
    if (file !== root && !file.startsWith(root + path.sep)) { res.writeHead(403).end(); return; }
    if (file === root) file = path.join(root, 'index.html');
    fs.readFile(file, (error, data) => {
      if (error) { res.writeHead(404).end(); return; }
      res.setHeader('Content-Type', file.endsWith('.js') ? 'text/javascript' : file.endsWith('.css') ? 'text/css' : 'text/html');
      res.end(data);
    });
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  let browser;
  try {
    browser = await chromium.launch({ executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe', headless: true, chromiumSandbox: true, timeout: 20000 });
    const page = await browser.newPage({ viewport: { width: 1440, height: 1080 }, deviceScaleFactor: 1 });
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto('http://127.0.0.1:' + server.address().port);
    await page.getByText('Your network. In motion.').waitFor();
    const nav = page.getByRole('navigation', { name: 'Workspace navigation' });
    async function navigate(label) {
      await nav.getByRole('button', { name: label, exact: true }).click();
      if (label === 'Overview') await page.getByText('Your network. In motion.').waitFor();
      else await page.getByRole('heading', { name: label, exact: true, level: 1 }).waitFor();
      assert.equal(await nav.locator('[aria-current="page"]').count(), 1);
      assert.equal(await nav.locator('[aria-current="page"]').innerText(), label);
    }
    const ratios = [];
    const layouts = [];
    for (const [theme, name] of themes) {
      await navigate('Settings');
      const choice = page.getByRole('button', { name: 'Use ' + name + ' theme', exact: true });
      await choice.click();
      assert.equal(await choice.getAttribute('aria-pressed'), 'true');
      assert.equal(await page.evaluate(() => document.documentElement.dataset.theme), theme);
      await page.getByText('Appearance preview only.', { exact: false }).waitFor();
      const pairs = await page.evaluate(() => {
        const pairs = [['--text', '--bg'], ['--text', '--panel'], ['--muted', '--bg'], ['--muted', '--panel'], ['--muted', '--panel2'], ['--primary-text', '--primary-bg'], ['--error-text', '--error-bg'], ['--cyan', '--panel']];
        return pairs.map(([fg, bg]) => {
          const probe = document.createElement('span');
          probe.style.cssText = `position:absolute;visibility:hidden;color:var(${fg});background-color:var(${bg})`;
          document.body.appendChild(probe);
          const style = getComputedStyle(probe);
          const result = { foregroundToken: fg, backgroundToken: bg, foreground: style.color, background: style.backgroundColor };
          probe.remove();
          return result;
        });
      });
      for (const pair of pairs) {
        const ratio = contrast(pair.foreground, pair.background);
        ratios.push({ theme, ...pair, contrast: Number(ratio.toFixed(3)) });
        assert.ok(ratio >= 4.5, `${theme} ${pair.foregroundToken}/${pair.backgroundToken} contrast ${ratio}`);
      }
      for (const viewport of [{ width: 1440, height: 1080 }, { width: 1000, height: 720 }, { width: 768, height: 900 }]) {
        await page.setViewportSize(viewport);
        for (const label of pages) {
          await navigate(label);
          const width = await page.evaluate(() => document.documentElement.scrollWidth);
          assert.ok(width <= viewport.width, `${theme}/${label} page overflow ${width} > ${viewport.width}`);
          layouts.push({ theme, screen: label, viewport: viewport.width, documentWidth: width });
          if (theme === 'forest' && viewport.width === 1440) {
            await page.screenshot({ path: path.join(out, 'frontend-' + label.toLowerCase().replaceAll(' ', '-') + '.png'), fullPage: true });
          }
        }
      }
      await page.setViewportSize({ width: 1440, height: 1080 });
      await navigate('Overview');
      await page.screenshot({ path: path.join(out, 'theme-' + theme + '-overview.png'), fullPage: true });
      await navigate('Settings');
      await page.screenshot({ path: path.join(out, 'theme-' + theme + '-settings.png'), fullPage: true });
    }
    await page.emulateMedia({ reducedMotion: 'reduce' });
    const transition = await nav.getByRole('button', { name: 'Overview', exact: true }).evaluate(el => getComputedStyle(el).transitionDuration);
    assert.ok(transition.split(',').every(n => Number.parseFloat(n) === 0), 'Reduced-motion transitions remain active');
    await nav.getByRole('button', { name: 'Overview', exact: true }).focus();
    await page.keyboard.press('Tab');
    const focus = await page.evaluate(() => ({ name: document.activeElement.textContent, outline: getComputedStyle(document.activeElement).outlineStyle, width: getComputedStyle(document.activeElement).outlineWidth }));
    assert.equal(focus.name, 'Messages');
    assert.notEqual(focus.outline, 'none');
    assert.ok(Number.parseFloat(focus.width) > 0);
    await page.keyboard.press('Enter');
    await page.getByRole('heading', { name: 'Messages', exact: true, level: 1 }).waitFor();
    assert.equal(errors.length, 0, errors.join('; '));
    const result = { status: 'PASS', scope: 'Built production frontend in signed Chrome; no mocked Go bridge. Appearance preview tests do not verify native persistence. Representative contrast/focus checks are not a complete WCAG audit.', navigationScreens: 6, themes: 4, layoutChecks: layouts.length, viewports: [1440, 1000, 768], contrastPairs: ratios, minimumContrast: Math.min(...ratios.map(r => r.contrast)), keyboardNavigation: 'PASS', reducedMotion: 'PASS', consoleErrors: errors, nativeDialogAndThemePersistence: 'NOT VERIFIED by this browser-only check; backend/CI evidence separate', recordedAt: new Date().toISOString() };
    fs.writeFileSync(path.join(out, 'ui-checks.json'), JSON.stringify(result, null, 2));
    console.log(`PASS: 6 screens / 4 palettes / ${layouts.length} layouts / ${ratios.length} contrast pairs / keyboard and reduced motion; no fake backend data.`);
  } finally {
    if (browser) await browser.close();
    server.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
