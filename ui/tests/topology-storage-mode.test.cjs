const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');

function load(relative, imports) {
  const source = fs.readFileSync(path.join(__dirname, '../src', relative), 'utf8');
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS },
  }).outputText;
  const context = { exports: {}, require: imports };
  vm.runInNewContext(compiled, context);
  return context.exports;
}

const intl = { formatMessage: ({ defaultMessage }) => defaultMessage };
const constants = load('constants/index.ts', (name) => {
  if (name === '@/utils/intl') return { intl };
  if (name.endsWith('.svg')) return { default: name };
  throw new Error(`Unexpected constants import: ${name}`);
});
const { formatTopoData } = load('components/TopoComponent/helper.ts', (name) => {
  if (name === '@/constants') return constants;
  if (name === '@/utils/intl') return { intl };
  if (name === 'lodash') return {};
  throw new Error(`Unexpected topology import: ${name}`);
});
const { formatClusterData } = load('pages/Cluster/Detail/Overview/helper.ts', () => constants);

for (const deploymentMode of ['shared_storage', 'normal', undefined]) {
  test(`topology and overview preserve deployment details: ${deploymentMode}`, () => {
    const input = {
      name: 'ob-demo-ob', namespace: 'ob-demo', status: 'running',
      mode: 'SERVICE', deploymentMode,
      ...(deploymentMode === 'shared_storage' ? {
        sharedStorageInfo: { bucketURL: 's3://acceptance', secretRef: { name: 'storage' } },
        logServiceRef: { name: 'ob-demo-ls' },
      } : {}),
      topology: [{ name: 'zone1', zone: 'zone1', namespace: 'ob-demo', status: 'running',
        observers: [{ name: 'server1', namespace: 'ob-demo', address: '192.0.2.1', status: 'running' }] }],
    };
    const before = JSON.stringify(input);
    for (const replicas of [undefined, [{ zone: 'zone1', minIops: 100, maxIops: 200 }]]) {
      const { basicInfo, topoData } = formatTopoData(input, replicas);
      for (const field of ['deploymentMode', 'sharedStorageInfo', 'logServiceRef', 'mode']) {
        assert.equal(basicInfo[field], input[field], field);
        assert.equal(JSON.stringify(basicInfo[field]), JSON.stringify(formatClusterData(JSON.parse(before)).info[field]), field);
      }
      assert.equal(topoData.children[0].children[0].label, '192.0.2.1');
    }
    assert.equal(JSON.stringify(input), before, 'topology must not mutate API response');
  });
}
