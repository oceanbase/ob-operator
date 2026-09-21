// Regression found by authenticated Playwright on 2026-09-21: initially both
// namespace and useRequest.data are undefined. This must not dereference items.
const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const Module = require('node:module');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');
const { Form } = require('antd');
const filename = path.join(__dirname, '../src/pages/Cluster/New/SharedStorage.tsx');
const mod = new Module(filename, module);
mod.filename = filename;
mod.paths = Module._nodeModulePaths(path.dirname(filename));
const realRequire = mod.require.bind(mod);
mod.require = id => {
  if (id === '@umijs/max') return { useAccess: () => ({ oblogserviceread: true }) };
  if (id === 'ahooks') return { useRequest: () => ({ data: undefined, loading: false, refresh() {} }) };
  if (id === '@/pages/LogService/New') return { __esModule: true, default: () => null };
  if (id === '@/pages/LogService/common') return { L: zh => zh, errorText: String };
  if (id === '@/pages/SharedStorage/ObjectStorage') return { BucketInput: () => null, CredentialPicker: () => null, StoragePreflight: () => null, StorageSummary: () => null };
  if (id === '@/pages/SharedStorage/storageConfig') return { bucketURLRules: [], parseBucketURL: () => undefined, storageLocationsOverlap: () => false };
  if (id === '@/services/logservice') return { lsRequest: () => { throw Error('No network in render regression'); } };
  return realRequire(id);
};
mod._compile(ts.transpileModule(fs.readFileSync(filename, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, esModuleInterop: true } }).outputText, filename);
const SharedStorage = mod.exports.default;
for (const section of ['storage', 'logservice']) {
  test(`SS wizard ${section} renders before namespace selection and request completion`, () => {
    function InitialState() {
      const [form] = Form.useForm();
      return React.createElement(Form, { form }, React.createElement(SharedStorage, { form, section }));
    }
    assert.doesNotThrow(() => renderToStaticMarkup(React.createElement(InitialState)));
  });
}
