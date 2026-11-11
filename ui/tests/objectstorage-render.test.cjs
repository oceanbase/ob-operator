// Component render tests, not a substitute for authenticated browser E2E.
const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const Module = require('node:module');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');
let access = {};
function load(relative) {
  const filename = path.join(__dirname, '../src', relative);
  const mod = new Module(filename, module);
  mod.filename = filename; mod.paths = Module._nodeModulePaths(path.dirname(filename));
  const realRequire = mod.require.bind(mod);
  mod.require = id => {
    if (id === '@umijs/max') return { useAccess: () => access };
    if (id === '@/pages/LogService/common') return { L: zh => zh, errorText: e => String(e) };
    if (id === '@/hook/usePublicKey') return { usePublicKey: () => '', encryptText: () => false };
    if (id === '@/services/objectstorage') return { storageRequest: () => { throw Error('Render tests must not make network calls'); } };
    if (id === './storageConfig') return load('pages/SharedStorage/storageConfig.ts');
    return realRequire(id);
  };
  const code = ts.transpileModule(fs.readFileSync(filename, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, esModuleInterop: true } }).outputText;
  mod._compile(code, filename); return mod.exports;
}
const { BucketInput, CredentialPicker, StoragePreflight, StorageSummary } = load('pages/SharedStorage/ObjectStorage.tsx');
const raw = 's3://sharedstorage/data?host=http://192.0.2.10:9000&s3_region=us-east-1';
const render = (component, props) => renderToStaticMarkup(React.createElement(component, props));

test('structured fields show endpoint, bucket, region and prefix independently', () => {
  const html = render(BucketInput, { value: raw });
  for (const value of ['192.0.2.10:9000', 'sharedstorage', 'us-east-1', '路径前缀', '结构化配置']) assert(html.includes(value));
});
test('a reused connection leaves log bucket/prefix blank for explicit selection', () => {
  const html = render(BucketInput, { value: '', defaultLocation: { endpoint: 'http://192.0.2.10:9000', region: 'us-east-1', bucket: '', prefix: '' } });
  assert(html.includes('value="http://192.0.2.10:9000"'));
  assert(!html.includes('value="sharedstorage"'));
});
test('storage summary does not echo inline credentials or unknown URL options', () => {
  const html = render(StorageSummary, { bucketURL: raw + '&access_key=DO_NOT_ECHO', secretName: 'minio-credentials' });
  assert(!html.includes('DO_NOT_ECHO')); assert(html.includes('minio-credentials'));
  assert(html.includes('无法按当前 S3 表单解析'));
});
test('reader sees credential selector but no create action; manual reference fallback remains', () => {
  access = { objectstorageread: true };
  const read = render(CredentialPicker, { namespace: 'oceanbase-ai' });
  assert(!read.includes('新建凭据')); assert(read.includes('选择同命名空间 Secret'));
  access = {};
  assert(render(CredentialPicker, { namespace: 'oceanbase-ai' }).includes('输入同命名空间 Secret 名称'));
});
test('preflight requires storage write permission and documents its limited scope', () => {
  access = {};
  const html = render(StoragePreflight, { namespace: 'oceanbase-ai', bucketURL: raw, secretName: 'minio-credentials' });
  assert(html.includes('disabled=""')); assert(html.includes('需要对象存储写权限'));
  assert(html.includes('不验证对象读写/删除和路径权限'));
});

test('storage summary and structured input accept supported OceanBase URL options', () => {
  const value = 's3://data-bucket/?host=https://s3.example.com&s3_region=cn-wulanchabu&scope=region1&max_iops=10000&max_bandwidth=1GB';
  const html = render(StorageSummary, { bucketURL: value, secretName: 'data-credentials', maxIOPS: '', maxBandwidth: '' });
  for (const text of ['https://s3.example.com', 'data-bucket', 'cn-wulanchabu', 'region1', '10000', '1GB']) assert(html.includes(text), text);
  assert(!html.includes('无法按当前 S3 表单解析'));
  assert(!html.includes('未配置'));
  const fields = render(BucketInput, { value });
  assert(fields.includes('value="https://s3.example.com"'));
  assert(fields.includes('value="data-bucket"'));
  access = { objectstoragewrite: true };
  const preflight = render(StoragePreflight, { namespace: 'test', bucketURL: value, secretName: 'data-credentials' });
  assert(!preflight.includes('disabled=""'));
  access = {};
});

test('explicit resource limits override URL fallback values in the summary', () => {
  const value = raw + '&max_iops=10000&max_bandwidth=1GB';
  const html = render(StorageSummary, { bucketURL: value, maxIOPS: '5000', maxBandwidth: '2GB' });
  assert(html.includes('5000')); assert(html.includes('2GB'));
  assert(!html.includes('10000')); assert(!html.includes('1GB'));
});
