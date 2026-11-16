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
  for (const name of [
    'result.json',
    'job-proof.json',
    'failure.png',
    'failure-text-private.txt',
  ]) {
    fs.rmSync(path.join(root, name), { force: true });
  }
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH,
    headless: true,
    args: ['--no-sandbox'],
  });
  let page;
  try {
    const context = await browser.newContext({
      viewport: { width: 1440, height: 1000 },
    });
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
    await page.goto(config.url + '/#/login', {
      waitUntil: 'networkidle',
      timeout: 60000,
    });
    await page.getByPlaceholder('Username', { exact: true }).fill('admin');
    await page
      .getByPlaceholder('Password', { exact: true })
      .fill(config.password);
    const login = page.waitForResponse(
      (r) => new URL(r.url()).pathname === '/api/v1/login'
    );
    await page.locator('button[type=submit]').click();
    const response = await login;
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
    const assert = require('node:assert/strict');
    const { execFileSync } = require('node:child_process');
    const k = (...args) =>
      JSON.parse(
        execFileSync('kubectl', args.concat(['-o', 'json']), {
          encoding: 'utf8',
        })
      );
    const deployment = k(
      '-n',
      config.namespace,
      'get',
      'deployment',
      'oceanbase-dashboard-' + config.release
    );
    const expectedSA =
      config.expectedServiceAccount ||
      deployment.spec.template.spec.serviceAccountName;
    await page.goto(
      config.url +
        '/#/cluster/' +
        config.namespace +
        '/' +
        config.release +
        '-ob/' +
        encodeURIComponent(config.clusterName) +
        '/overview',
      { waitUntil: 'networkidle' }
    );
    await page
      .locator('.ant-descriptions-item-content')
      .filter({ hasText: config.clusterName })
      .first()
      .waitFor({ timeout: 60000 });
    await page
      .getByRole('button', { name: /Cluster Management|集群管理/ })
      .hover();
    await page
      .getByText(/^(Log Download|日志下载)$/)
      .first()
      .click();
    const dialog = page.locator('.ant-modal-content:visible');
    await dialog.locator('.ant-select-selector').click();
    await page
      .getByText(/^(Last 30 minutes|近 ?30 ?分钟|最近 ?30 ?分钟)$/)
      .last()
      .click();
    const creation = page.waitForResponse((r) =>
      /\/api\/v1\/obclusters\/namespace\/.+\/log$/.test(
        new URL(r.url()).pathname
      )
    );
    await dialog
      .getByRole('button', { name: /^(OK|Confirm|确\s*认|确\s*定)$/ })
      .click();
    const body = await (await creation).json();
    assert.equal(body.successful, true, 'log job request must succeed');
    const job = body.data;
    const actual = k('-n', job.namespace, 'get', 'job', job.name);
    const proof = {
      testedAt: new Date().toISOString(),
      namespace: job.namespace,
      job: job.name,
      expectedServiceAccount: expectedSA,
      actualServiceAccount: actual.spec.template.spec.serviceAccountName,
      dashboardImage: deployment.spec.template.spec.containers.find(
        (c) => c.name === 'dashboard'
      ).image,
    };
    fs.writeFileSync(root + '/job-proof.json', JSON.stringify(proof, null, 2));
    assert.equal(
      proof.actualServiceAccount,
      expectedSA,
      'log Job must use the configured Dashboard ServiceAccount'
    );
    const sharedLevel = deployment.spec.template.spec.securityContext?.seLinuxOptions?.level;
    if (sharedLevel) {
      assert.equal(
        actual.spec.template.spec.securityContext?.seLinuxOptions?.level,
        sharedLevel,
        'log Job must share the Dashboard SELinux level for its report PVC'
      );
      proof.sharedVolumeSELinuxLevel = sharedLevel;
      fs.writeFileSync(root + '/job-proof.json', JSON.stringify(proof, null, 2));
    }
    if (process.env.EXPECT_MISSING_ACCOUNT === '1') {
      await page
        .getByRole('button', { name: /^(View Details|查看详情)$/ })
        .waitFor({ timeout: 120000 });
      await page
        .getByRole('button', { name: /^(View Details|查看详情)$/ })
        .click();
      await page
        .getByText(/FailedCreate:.*serviceaccount.*not found/)
        .first()
        .waitFor({ timeout: 30000 });
      proof.status = 'PASS';
      proof.scenario = 'missing-account-error-visible';
    } else {
      const outcome = await Promise.race([
        page
          .getByRole('button', { name: /^(Download Link|下载链接)$/ })
          .waitFor({ timeout: 600000 })
          .then(() => 'download'),
        page
          .getByRole('button', { name: /^(View Details|查看详情)$/ })
          .waitFor({ timeout: 600000 })
          .then(() => 'failure'),
      ]);
      assert.equal(
        outcome,
        'download',
        'collection must succeed, not show a failure'
      );
      const downloadEvent = page.waitForEvent('download');
      await page
        .getByRole('button', { name: /^(Download Link|下载链接)$/ })
        .click();
      const download = await downloadEvent;
      const filename = root + '/observer-logs-private.tar.gz';
      await download.saveAs(filename);
      assert.equal(await download.failure(), null);
      assert.ok(fs.statSync(filename).size > 0, 'download must contain data');
      const entries = execFileSync('tar', ['-tzf', filename], {
        encoding: 'utf8',
      })
        .split('\n')
        .filter(Boolean);
      assert.ok(
        entries.some((name) => /(?:^|\/)observer_log_.+\.tar\.gz$/.test(name)),
        'download must include an Observer log archive'
      );
      assert.ok(
        !entries.some((n) => /(^|\/)(config\.yaml|k8s|token)(\/|$)/.test(n)),
        'archive must not contain generated credentials'
      );
      const completed = k('-n', job.namespace, 'get', 'job', job.name);
      assert.ok(
        completed.status.succeeded > 0,
        'Kubernetes collection Job must finish'
      );
      proof.status = 'PASS';
      proof.scenario = 'custom-release-log-download';
      proof.archiveBytes = fs.statSync(filename).size;
      proof.archiveEntries = entries;
    }
    assert.deepEqual(errors, [], 'no browser JavaScript errors');
    assert.ok(assets.length > 0, 'Dashboard assets were loaded');
    assert.ok(
      assets.every((asset) => asset.status < 400),
      'Dashboard assets must load successfully'
    );
    await page.screenshot({ path: root + '/log-download.png', fullPage: true });
    fs.writeFileSync(root + '/result.json', JSON.stringify(proof, null, 2));
    console.log('PASS ' + proof.scenario);
  } catch (error) {
    if (page) {
      await page
        .screenshot({ path: root + '/failure.png', fullPage: true })
        .catch(() => {});
      fs.writeFileSync(
        root + '/failure-text-private.txt',
        await page
          .locator('body')
          .innerText()
          .catch(() => '')
      );
    }
    throw error;
  } finally {
    await browser.close();
  }
})().catch((e) => {
  console.error(String(e.stack || e).replaceAll(config.password, '[redacted]'));
  process.exitCode = 1;
});
