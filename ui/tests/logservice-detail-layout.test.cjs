const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const { matchRoutes } = require('react-router-dom');

const context = { exports: {} };
vm.runInNewContext(ts.transpileModule(
  fs.readFileSync(path.join(__dirname, '../config/routes.ts'), 'utf8'),
  { compilerOptions: { module: ts.ModuleKind.CommonJS } },
).outputText, context);
const routes = context.exports.default;
function routerTree(items) {
  return items.map(item => ({ ...item, children: item.routes && routerTree(item.routes) }));
}
const tree = routerTree(routes);
const matches = url => matchRoutes(tree, url) || [];

test('LogService detail uses a persistent detail layout, like OB cluster detail', () => {
  const matched = matches('/logservice/test/ls');
  assert(matched.some(m => m.route.component === 'LogService/DetailLayout'));
  assert.equal(matched.at(-1).route.redirect, 'overview');
  assert(matches('/cluster/test/ob/ob/overview').some(m => m.route.component === 'Cluster/Detail'));
});

for (const section of ['overview', 'storage', 'monitor', 'parameters', 'events']) {
  test(`LogService ${section} deep link retains the same shell and resource identity`, () => {
    const matched = matches(`/logservice/test/ls/${section}?from=lakehouse`);
    assert(matched.some(m => m.route.component === 'LogService/DetailLayout'));
    assert.equal(matched.at(-1).route.component, 'LogService/Detail');
    assert.equal(matched.at(-1).params.ns, 'test');
    assert.equal(matched.at(-1).params.name, 'ls');
  });
}

test('list and creation routes remain separate from the resource detail shell', () => {
  assert.equal(matches('/logservice').at(-1).route.component, 'LogService');
  assert.equal(matches('/logservice/new').at(-1).route.component, 'LogService/New');
  assert(!matches('/logservice/new').some(m => m.route.component === 'LogService/DetailLayout'));
});

// Render application components with real Ant Design content. Only the shared
// third-party layout shell and network/router hooks are replaced by test doubles.
const Module = require('node:module');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');
let currentPath = '/logservice/test/ls/overview';
let currentAccess = { oblogserviceread: true, obclusterread: true };
let shellProps;
const detail = {
  name: 'ls', namespace: 'test', references: [], protected: false,
  status: { status: 'running' },
  spec: {
    clusterId: 123,
    topology: [{ zone: 'zone1', replica: 3 }],
    logService: { image: 'example/logservice:test', resource: { cpu: 2, memory: '4Gi' }, storage: { storeStorage: { size: '20Gi' }, logStorage: { size: '10Gi' } } },
    objectStoreUrl: { bucketURL: 's3://logs?host=http://s3.example.com&s3_region=test', secretRef: { name: 'test-credentials' } },
    parameters: [{ name: 'test_param', value: 'test_value' }],
  },
  nodes: [], volumes: [], events: [],
};
const noop = () => {};
const locale = { L: zh => zh, errorText: String };
function loadComponent(relative) {
  const filename = path.join(__dirname, '../src', relative);
  const mod = new Module(filename, module);
  mod.filename = filename;
  mod.paths = Module._nodeModulePaths(path.dirname(filename));
  const realRequire = mod.require.bind(mod);
  mod.require = id => {
    if (id === '@umijs/max') return {
      useParams: () => ({ ns: 'test', name: 'ls' }),
      useLocation: () => ({ pathname: currentPath }),
      useAccess: () => currentAccess,
      useModel: () => ({ initialState: { accountInfo: { nickname: 'tester' } }, reportDataInterval: { current: null } }),
      history: { push: noop }, Outlet: () => React.createElement('main', null, 'detail-outlet'),
    };
    if (id === 'ahooks') return { useRequest: () => ({ data: detail, loading: false, refresh: noop, run: noop }) };
    if (id === './common') return locale;
    if (id === '@/pages/Layouts/DetailLayout') return { __esModule: true, default: loadComponent('pages/Layouts/DetailLayout/index.tsx') };
    if (id === '@/assets/logo1.svg') return 'logo.svg';
    if (id.startsWith('@/components/customModal/')) return { __esModule: true, default: () => null };
    if (id === '@/services') return { logoutReq: noop };
    if (id === '@/utils/helper') return { getAppInfoFromStorage: async () => ({ version: 'test' }) };
    if (id === '@/utils/intl') return { intl: { formatMessage: ({ defaultMessage }) => defaultMessage } };
    // The design barrel imports CSS; its Menu wraps the same Ant Design Menu.
    if (id === '@oceanbase/design') return { Menu: realRequire('antd').Menu };
    if (id === '@/services/logservice') return { lsPath: () => '/test-only', lsRequest: () => { throw Error('No network in render tests'); } };
    if (id === '@/pages/SharedStorage/Monitor') return { __esModule: true, default: () => React.createElement('section', null, 'LS-monitor-content') };
    if (id === '@/pages/SharedStorage/ObjectStorage') return {
      StorageSummary: () => React.createElement('section', null, 'log-storage-summary'),
      StoragePreflight: () => React.createElement('section', null, 'log-storage-check'),
    };
    if (id === '@oceanbase/ui') return {
      IconFont: () => null,
      BasicLayout: props => {
        shellProps = props;
        return React.createElement('div', null,
          React.createElement('header', null, props.topHeader.username),
          React.createElement('aside', null, props.sideHeader,
            props.menus.map(item => React.createElement('a', { key: item.link, href: item.link }, item.title))),
          props.children);
      },
    };
    return realRequire(id);
  };
  mod._compile(ts.transpileModule(fs.readFileSync(filename, 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, esModuleInterop: true },
  }).outputText, filename);
  return mod.exports.default;
}

test('LogService reuses the OB detail shell with resource sidebar, top header and global navigation', () => {
  currentAccess = { oblogserviceread: true, obclusterread: true };
  const Layout = loadComponent('pages/LogService/DetailLayout.tsx');
  const html = renderToStaticMarkup(React.createElement(Layout));
  for (const label of ['LogService · ls', '概览与拓扑', '日志对象存储', 'LS 专项监控', '启动参数', '事件', 'tester', 'detail-outlet']) assert(html.includes(label), label);
  assert.deepEqual(shellProps.subSideMenuProps.selectedKeys, ['/logservice']);
  assert.equal(shellProps.subSideMenus.find(m => m.key === 'logservice').accessible, true);
  assert(shellProps.subSideMenus.some(m => m.key === 'cluster'));
  for (const section of ['overview', 'storage', 'monitor', 'parameters', 'events']) assert(html.includes(`/logservice/test/ls/${section}`));
});

test('shared detail shell keeps LogService navigation permission-controlled for every resource', () => {
  const Layout = loadComponent('pages/Layouts/DetailLayout/index.tsx');
  for (const access of [{}, { oblogserviceread: true }, { oblogservicewrite: true }]) {
    currentAccess = access;
    renderToStaticMarkup(React.createElement(Layout, { menus: [], sideHeader: null, subSideSelectKey: 'cluster' }));
    assert.equal(!!shellProps.subSideMenus.find(m => m.key === 'logservice').accessible, !!(access.oblogserviceread || access.oblogservicewrite));
    assert.deepEqual(shellProps.subSideMenuProps.selectedKeys, ['/cluster']);
  }
});

for (const [section, marker] of Object.entries({ overview: 'Zone 拓扑', storage: 'log-storage-summary', monitor: 'LS-monitor-content', parameters: 'test_param', events: '详情' })) {
  test(`${section} URL renders its content without the former standalone tab bar`, () => {
    currentPath = `/logservice/test/ls/${section}`;
    currentAccess = { oblogserviceread: true };
    const Detail = loadComponent('pages/LogService/Detail.tsx');
    const html = renderToStaticMarkup(React.createElement(Detail));
    assert(html.includes(marker), marker);
    assert(!html.includes('role="tablist"'));
    assert(html.includes('扩缩容副本'));
    assert(html.includes('disabled=""'), 'read-only users cannot mutate replicas');
    if (section !== 'monitor') assert(!html.includes('LS-monitor-content'));
    if (section !== 'storage') assert(!html.includes('log-storage-summary'));
  });
}
