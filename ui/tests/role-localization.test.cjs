const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const Module = require('node:module');
const ts = require('typescript');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');
const { createIntl } = require('react-intl');
const domains = [
  'objectstorage',
  'oblogservice',
  'obcluster',
  'system',
  'alarm',
  'obproxy',
  'ac',
  'k8s-cluster',
];

function fixture(locale) {
  const errors = [];
  const intl = createIntl({
    locale,
    messages: require(`../src/i18n/strings/${locale}.json`),
    onError: (e) => errors.push(e),
  });
  const cache = new Map();
  function load(relative) {
    if (cache.has(relative)) return cache.get(relative);
    const file = path.resolve(__dirname, '../src', relative);
    const mod = new Module(file, module);
    mod.filename = file;
    mod.paths = Module._nodeModulePaths(path.dirname(file));
    const original = mod.require.bind(mod);
    mod.require = (id) => {
      if (id === '@/utils/intl') return { intl };
      if (id === '@/constants/access') return load('constants/access.ts');
      if (id === '@/pages/Access/type') return load('pages/Access/type.ts');
      if (id === '@/api') return { access: {} };
      if (id === '@umijs/max')
        return {
          useModel: () => ({
            initialState: {
              policies: domains.flatMap((domain) =>
                ['read', 'write'].map((action) => ({
                  domain,
                  action,
                  object: '*',
                })),
              ),
            },
          }),
        };
      if (id === '.')
        return {
          __esModule: true,
          default: ({ title, children }) =>
            React.createElement(
              'section',
              null,
              React.createElement('h1', null, title),
              children,
            ),
        };
      if (id === '../IconTip')
        return {
          __esModule: true,
          default: ({ tip, content }) =>
            React.createElement(
              'span',
              { title: tip, 'data-permission-label': true },
              content,
            ),
        };
      return original(id);
    };
    mod._compile(
      ts.transpileModule(fs.readFileSync(file, 'utf8'), {
        compilerOptions: {
          module: ts.ModuleKind.CommonJS,
          jsx: ts.JsxEmit.ReactJSX,
          esModuleInterop: true,
        },
      }).outputText,
      file,
    );
    cache.set(relative, mod.exports);
    return mod.exports;
  }
  return { load, errors };
}

for (const locale of ['en-US', 'zh-CN']) {
  test(`${locale}: create and edit titles use valid translated messages`, () => {
    const { load, errors } = fixture(locale);
    const Modal = load('components/customModal/HandleRoleModal.tsx').default;
    for (const type of ['create', 'edit']) {
      const html = renderToStaticMarkup(
        React.createElement(Modal, {
          visible: true,
          type,
          setVisible() {},
          editValue: { name: 'role', policies: [] },
        }),
      );
      const expected =
        locale === 'en-US'
          ? `${type === 'create' ? 'Create' : 'Edit'} Role`
          : `${type === 'create' ? '创建' : '编辑'}角色`;
      assert(html.includes(`<h1>${expected}</h1>`), expected);
      assert.equal(
        (html.match(/data-permission-label="true"/g) || []).length,
        domains.length,
      );
    }
    assert.deepEqual(
      errors.map((e) => e.message),
      [],
      'all rendered messages must exist and parse',
    );
  });
  test(`${locale}: every server policy has a localized label and description`, () => {
    const { load, errors } = fixture(locale);
    const list = load('constants/access.ts').ACCESS_ROLES_LIST;
    assert.deepEqual(list.map((x) => x.value).sort(), [...domains].sort());
    for (const domain of domains) {
      const row = list.find((x) => x.value === domain);
      assert(row.label && row.descriptions, domain);
      if (locale === 'en-US')
        assert(!/[\u3400-\u9fff]/u.test(row.label + row.descriptions), domain);
      else assert(/[\u3400-\u9fff]/u.test(row.descriptions), domain);
    }
    assert.deepEqual(
      errors.map((e) => e.message),
      [],
    );
  });
}
