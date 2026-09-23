// Run with PLAYWRIGHT_MODULE and CHROMIUM_EXECUTABLE set to local test tools.
// Bundles the real components/design-system CSS; only app data/router hooks are mocked.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const esbuild = require('esbuild');
const less = require('less');
const root = path.resolve(__dirname, '..');
const output =
  process.env.OUTPUT_DIR ||
  path.resolve(root, '../testreports/dima-dashboard-layout/browser');
const mocks = {
  '@umijs/max': `import React from 'react';
    export const history={push:(p)=>{window.history.pushState({},'',p);window.dispatchEvent(new Event('popstate'));}};
    export const Link=({to,children,...p})=><a href={to} {...p}>{children}</a>;
    export const useAccess=()=>({obclusterread:true,obclusterwrite:true,oblogserviceread:true});
    export const useParams=()=>({ns:'test',name:'tenant-resource',tenantName:'sec_test'});
    export const useModel=(name)=>name==='global'?{reportDataInterval:{current:null}}:{initialState:{accountInfo:{nickname:'tester'}}};
    export function useLocation(){const [pathname,set]=React.useState(window.location.pathname);React.useEffect(()=>{const fn=()=>set(window.location.pathname);window.addEventListener('popstate',fn);return()=>window.removeEventListener('popstate',fn)},[]);return {pathname};}
    export const Outlet=()=> <main data-testid="body"><h1>连接租户 / Data filter</h1><div>Zone: all</div></main>;`,
  '@/services': `export function getObclusterListReq(){} export function getClusterDetailReq(){} export function logoutReq(){}`,
  '@/services/tenant': `export function getTenant(){} export function getAllTenants(){}`,
  ahooks: `const info={clusterName:'ob-ai-test-with-a-very-long-cluster-name',clusterId:2026092202,clusterResourceName:'ob-ai-test',tenantName:'sec_test_with_a_long_tenant_name',status:'running'};
    export function useRequest(fn){const data=fn.name==='getTenant'?{data:{info}}:fn.name==='getClusterDetailReq'?{info}:fn.name==='getObclusterListReq'?{data:[{...info,name:'ob-ai-test',namespace:'test'}]}:fn.name==='getAllTenants'?{data:[{...info,name:'tenant-resource',namespace:'test'}]}:undefined;return {data,run:()=>{}};}`,
  '@/constants': `export const STATUS_LIST=[{value:'running',label:'Running',badgeStatus:'success'}];export const MODE_MAP=new Map([['SERVICE','Service']]);`,
  '@/utils/helper': `export const getAppInfoFromStorage=async()=>({version:'test'});`,
  '@/utils/intl': `export const intl={formatMessage:({defaultMessage})=>defaultMessage};`,
  '@/pages/LogService/common': `export const L=(zh,en)=>new URLSearchParams(location.search).get('lang')==='en'?en:zh;`,
  '@oceanbase/ui': `export {default as BasicLayout} from '@oceanbase/ui/es/BasicLayout';export const IconFont=()=>null;`,
};
async function run() {
  fs.mkdirSync(output, { recursive: true });
  await esbuild.build({
    stdin: {
      contents: `import React from 'react';import {createRoot} from 'react-dom/client';import {BrowserRouter} from 'react-router-dom';import Tenant from './src/pages/Tenant/Detail';import ClusterList from './src/pages/Cluster/ClusterList';
      const row={name:'ob-ai-test',namespace:'test',clusterName:'ob-ai-test',deploymentMode:'shared_storage',status:'running',statusDetail:'running',image:'test-image',topology:[],mode:'SERVICE'};
      createRoot(document.getElementById('root')).render(<BrowserRouter>{location.pathname==='/clusters'?<ClusterList clusterList={[row]} loading={false} handleAddCluster={()=>{}}/>:<Tenant/>}</BrowserRouter>);`,
      resolveDir: root,
      loader: 'tsx',
    },
    bundle: true,
    outfile: path.join(output, 'fixture.js'),
    platform: 'browser',
    jsx: 'automatic',
    define: { 'process.env.NODE_ENV': '"test"' },
    loader: {
      '.svg': 'dataurl',
      '.png': 'dataurl',
      '.woff': 'dataurl',
      '.woff2': 'dataurl',
    },
    plugins: [
      {
        name: 'fixture',
        setup(build) {
          build.onResolve({ filter: /.*/ }, (args) => {
            if (
              args.path === 'ahooks' &&
              args.importer.includes('/node_modules/')
            )
              return;
            if (Object.hasOwn(mocks, args.path))
              return { path: args.path, namespace: 'mock' };
            if (args.path.startsWith('@/components/customModal/'))
              return { path: 'empty-modal', namespace: 'mock' };
            if (args.path.startsWith('@/')) {
              const base = path.join(root, 'src', args.path.slice(2));
              const file = ['', '.tsx', '.ts', '.js', '/index.tsx', '/index.ts']
                .map((x) => base + x)
                .find((x) => fs.existsSync(x) && fs.statSync(x).isFile());
              if (file) return { path: file };
            }
          });
          build.onLoad({ filter: /.*/, namespace: 'mock' }, (args) => ({
            contents: mocks[args.path] || 'export default ()=>null;',
            loader: 'tsx',
            resolveDir: root,
          }));
          build.onLoad({ filter: /\.less$/ }, async (args) => ({
            contents: (
              await less.render(fs.readFileSync(args.path, 'utf8'), {
                filename: args.path,
                javascriptEnabled: true,
                plugins: [
                  {
                    install(_, manager) {
                      class Modules extends less.FileManager {
                        supports(name) {
                          return name.startsWith('~');
                        }
                        loadFile(name) {
                          const file = require.resolve(name.slice(1), {
                            paths: [root],
                          });
                          return Promise.resolve({
                            filename: file,
                            contents: fs.readFileSync(file, 'utf8'),
                          });
                        }
                      }
                      manager.addFileManager(new Modules());
                    },
                  },
                ],
              })
            ).css,
            loader: 'css',
          }));
        },
      },
    ],
    logLevel: 'warning',
  });
  const { chromium } = require(process.env.PLAYWRIGHT_MODULE ||
    'playwright-core');
  const browser = await chromium.launch({
    headless: true,
    executablePath: process.env.CHROMIUM_EXECUTABLE,
  });
  const server = http.createServer((req, res) => {
    if (req.url.startsWith('/fixture.')) {
      const f = path.join(output, req.url.slice(1));
      res.setHeader(
        'Content-Type',
        f.endsWith('.css') ? 'text/css' : 'text/javascript',
      );
      res.end(fs.readFileSync(f));
    } else {
      res.setHeader('Content-Type', 'text/html');
      res.end(
        '<!doctype html><html><head><meta charset="utf-8"><link rel="stylesheet" href="/fixture.css"><style>body{margin:0}main{padding:24px}#root{min-width:1280px}</style></head><body><div id="root"></div><script src="/fixture.js"></script></body></html>',
      );
    }
  });
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const failures = [];
  try {
    for (const width of [1280, 1920]) {
      const page = await browser.newPage({ viewport: { width, height: 1000 } });
      page.on('pageerror', (e) =>
        failures.push({ width, pageError: e.message }),
      );
      const base = `http://127.0.0.1:${server.address().port}`;
      await page.goto(base + '/tenant/test/tenant-resource/sec_test/monitor');
      await page.locator('.ob-layout-sider').waitFor();
      await page.waitForTimeout(600);
      for (const label of ['性能监控', '连接租户']) {
        const item = page
          .locator('.ob-layout-sider-content [role="menuitem"]')
          .filter({ hasText: label });
        await item.hover();
        await item.click();
        await page.waitForTimeout(100);
        assert.match(await item.getAttribute('class'), /selected/);
        assert(
          page
            .url()
            .endsWith(label === '性能监控' ? '/monitor' : '/connection'),
        );
        const boxes = await page.evaluate(() => {
          const sider = document
            .querySelector('.ob-layout-sider')
            .getBoundingClientRect();
          return {
            right: sider.right,
            items: [
              ...document.querySelectorAll(
                '.ob-layout-sider-content [role="menuitem"]',
              ),
            ].map((x) => x.getBoundingClientRect().right),
            selects: [
              ...document.querySelectorAll(
                '.ob-layout-sider-header .ant-select',
              ),
            ].map((x) => x.getBoundingClientRect().right),
          };
        });
        if (
          boxes.items.some((x) => x > boxes.right + 1) ||
          boxes.selects.some((x) => x > boxes.right + 1)
        )
          failures.push({ width, label, ...boxes });
      }
      const selection = page
        .locator('.ob-layout-sider-header .ant-select-selection-item')
        .first();
      assert(
        ((await selection.getAttribute('title')) || '').includes('2026092202'),
        'full cluster name remains available on hover',
      );
      assert.equal(
        await selection.evaluate((el) => getComputedStyle(el).textOverflow),
        'ellipsis',
      );
      await page.screenshot({ path: path.join(output, `tenant-${width}.png`) });
      for (const lang of ['en', 'zh']) {
        await page.goto(base + '/clusters?lang=' + lang);
        await page.locator('tbody .ant-tag').first().waitFor();
        const box = await page
          .locator('tbody .ant-tag')
          .first()
          .evaluate((el) => ({
            tag: el.getBoundingClientRect().toJSON(),
            cell: el.closest('td').getBoundingClientRect().toJSON(),
          }));
        if (box.tag.right > box.cell.right - 8)
          failures.push({ width, table: box });
        assert(
          (await page.locator('tbody a').first().getAttribute('href')).endsWith(
            '/lakehouse',
          ),
        );
        await page.screenshot({
          path: path.join(output, `clusters-${lang}-${width}.png`),
        });
      }
      await page.close();
    }
    console.log(JSON.stringify({ failures }, null, 2));
    assert.equal(
      failures.length,
      0,
      'menus, selectors and architecture tag stay inside their containers',
    );
  } finally {
    await browser.close();
    server.close();
  }
}
run().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
