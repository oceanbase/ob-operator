const fs = require('fs'),
  path = require('path');
// Run only against an authorized disposable Dashboard. The private config file
// contains url, password (admin), namespace, release and clusterName.
const root = path.resolve(process.argv[2]);
const { chromium } = require('playwright');
const config = JSON.parse(
  fs.readFileSync(root + '/browser-default-config.json', 'utf8')
);
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH,
    headless: true,
    args: ['--no-sandbox'],
  });
  try {
    const context = await browser.newContext({
        viewport: { width: 1440, height: 1000 },
      }),
      page = await context.newPage();
    const errors = [],
      assets = [];
    page.on('pageerror', (e) =>
      errors.push(
        (e.stack || e.message).replaceAll(config.password, '[redacted]')
      )
    );
    page.on('response', (r) => {
      const u = new URL(r.url());
      if (u.origin === config.url && /\.(js|css)$/.test(u.pathname))
        assets.push({ path: u.pathname, status: r.status() });
    });
    // Submit from a fresh browser before the first public-key response arrives.
    let releasePublicKey;
    const submitted = new Promise((resolve) => {
      releasePublicKey = resolve;
    });
    await page.route('**/api/v1/info', async (route) => {
      await submitted;
      await route.continue();
    });
    await page.goto(config.url + '/#/login', {
      waitUntil: 'domcontentloaded',
      timeout: 60000,
    });
    await page.getByPlaceholder('Username', { exact: true }).fill('admin');
    await page
      .getByPlaceholder('Password', { exact: true })
      .fill(config.password);
    const login = page.waitForResponse(
      (r) => new URL(r.url()).pathname === '/api/v1/login'
    );
    try {
      await page.locator('button[type=submit]').click();
    } finally {
      releasePublicKey();
    }
    const response = await login;
    const encryptedPassword = response.request().postDataJSON().password;
    if (
      typeof encryptedPassword !== 'string' ||
      !encryptedPassword ||
      encryptedPassword === config.password
    ) {
      throw Error('Cold login must submit an encrypted password');
    }
    const loginBody = await response.json();
    if (!loginBody.successful) throw Error('Browser login rejected');
    await page.waitForURL((u) => u.hash !== '#/login', { timeout: 30000 });
    await page.waitForLoadState('networkidle');
    if (loginBody.data.needReset) {
      await page.waitForURL((u) => u.hash === '#/reset', { timeout: 30000 });
      await page
        .locator('input[type=password]')
        .nth(1)
        .waitFor({ state: 'visible' });
      const rotated =
        'Qa!9' + require('crypto').randomBytes(16).toString('hex');
      await page.locator('input[type=password]').nth(0).fill(rotated);
      await page.locator('input[type=password]').nth(1).fill(rotated);
      await page.locator('button[type=submit]').click();
      await page.waitForURL((u) => u.hash === '#/overview', { timeout: 30000 });
      config.password = rotated;
      fs.writeFileSync(
        root + '/browser-default-config.json',
        JSON.stringify(config),
        { mode: 0o600 }
      );
      console.log('PASS mandatory initial password reset');
    }
    const screenshots = root + '/browser-default-evidence';
    fs.mkdirSync(screenshots, { recursive: true, mode: 0o700 });
    console.log('PASS browser login with the current Dashboard credential');
    const routes = [
      '/overview',
      '/cluster/' +
        config.namespace +
        '/' +
        config.release +
        '-ob/' +
        config.clusterName +
        '/topo',
      '/cluster/' +
        config.namespace +
        '/' +
        config.release +
        '-ob/' +
        config.clusterName +
        '/overview',
    ];
    const crypto = require('crypto');
    const info = await (
      await context.request.get(config.url + '/api/v1/info')
    ).json();
    for (
      let iteration = 0;
      iteration < Number(process.env.RESET_ITERATIONS || 3);
      iteration++
    ) {
      const username = 'qa-reset-' + crypto.randomBytes(6).toString('hex'),
        password = 'Qa!9' + crypto.randomBytes(16).toString('hex');
      const encrypted = crypto
        .publicEncrypt(
          {
            key: info.data.publicKey.replaceAll('RSA PUBLIC KEY', 'PUBLIC KEY'),
            padding: crypto.constants.RSA_PKCS1_PADDING,
          },
          Buffer.from(password)
        )
        .toString('base64');
      const create = await (
        await context.request.post(config.url + '/api/v1/ac/accounts', {
          data: {
            username,
            nickname: username,
            roles: ['admin'],
            password: encrypted,
          },
        })
      ).json();
      if (!create.successful)
        throw Error('Create account failed: ' + create.message);
      const c = await browser.newContext({
          viewport: { width: 1440, height: 1000 },
        }),
        p = await c.newPage(),
        failures = [];
      p.on('pageerror', (e) => failures.push(e.stack || e.message));
      try {
        const cdp = await c.newCDPSession(p);
        await cdp.send('Emulation.setCPUThrottlingRate', {
          rate: iteration % 2 ? 6 : 1,
        });
        await p.route('**/*', async (route) => {
          const u = new URL(route.request().url());
          if (u.pathname.startsWith('/api/') || u.pathname.endsWith('.js'))
            await new Promise((r) =>
              setTimeout(
                r,
                u.pathname === '/api/v1/ac/info' ||
                  u.pathname === '/api/v1/ac/policies'
                  ? 3000
                  : Math.floor(Math.random() * 300)
              )
            );
          await route.continue();
        });
        await p.goto(config.url + '/#/login', {
          waitUntil: 'networkidle',
          timeout: 60000,
        });
        await p.getByPlaceholder('Username', { exact: true }).fill(username);
        await p.getByPlaceholder('Password', { exact: true }).fill(password);
        const lr = p.waitForResponse(
          (r) => new URL(r.url()).pathname === '/api/v1/login'
        );
        await p.locator('button[type=submit]').click();
        const result = await (await lr).json();
        if (!result.successful || !result.data.needReset)
          throw Error('Expected first-login reset');
        await p.waitForURL((u) => u.hash === '#/reset', { timeout: 60000 });
        await p.locator('input[type=password]').nth(1).waitFor();
        const rotated = 'Qa!9' + crypto.randomBytes(16).toString('hex');
        await p.locator('input[type=password]').nth(0).fill(rotated);
        await p.locator('input[type=password]').nth(1).fill(rotated);
        await p.locator('button[type=submit]').click();
        await p.waitForURL((u) => u.hash === '#/overview', { timeout: 60000 });
        for (const route of [routes[1], routes[2], routes[0]]) {
          await p.goto(config.url + '/#' + route, {
            waitUntil: 'networkidle',
            timeout: 60000,
          });
          await p.locator('main').waitFor({ state: 'visible', timeout: 30000 });
          await p.waitForFunction(
            () => document.body.innerText.length > 100,
            null,
            { timeout: 30000 }
          );
          if (route === routes[1]) {
            await p
              .getByText(config.release + '-ls', { exact: true })
              .first()
              .waitFor({ timeout: 60000 });
          } else if (route === routes[2]) {
            await p
              .locator('.ant-descriptions-item-content')
              .filter({ hasText: config.release + '-ls' })
              .first()
              .waitFor({ timeout: 60000 });
          } else {
            const rows = p.locator('.ant-table-row');
            await (config.nodeName
              ? rows.filter({ hasText: config.nodeName })
              : rows
            )
              .first()
              .waitFor({ timeout: 60000 });
          }
        }
        if (failures.length) throw Error('JS exception ' + failures.join(';'));
        console.log('PASS first-reset iteration ' + iteration);
      } catch (e) {
        fs.writeFileSync(
          root + '/reset-failure-' + iteration + '.json',
          JSON.stringify({ iteration, error: String(e), failures }, null, 2)
        );
        await p.screenshot({
          path: root + '/reset-failure-' + iteration + '.png',
        });
        throw e;
      } finally {
        await c.close();
        const d = await context.request.delete(
          config.url + '/api/v1/ac/accounts/' + username
        );
        if (!d.ok()) throw Error('Account cleanup failed');
      }
    }
    if (assets.length < 2 || assets.some((a) => a.status >= 400))
      throw Error('Static assets failed');
    if (errors.length)
      throw Error('Browser JavaScript exception: ' + errors.join('; '));
    fs.writeFileSync(
      root + '/browser-default-evidence.json',
      JSON.stringify(
        {
          content_ready_checks: true,
          browser: await browser.version(),
          routes,
          assets,
          uncaught_errors: errors,
        },
        null,
        2
      )
    );
    console.log(
      'PASS: actual Chromium login, lists, details and direct JS/CSS loading'
    );
  } finally {
    await browser.close();
  }
})().catch((e) => {
  console.error(String(e.stack || e).replaceAll(config.password, '[redacted]'));
  process.exitCode = 1;
});
