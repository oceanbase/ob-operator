const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const Module = require('node:module');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');
let response;
function load(relative) {
  const filename = path.join(__dirname, '../src', relative);
  const mod = new Module(filename, module);
  mod.filename = filename;
  mod.paths = Module._nodeModulePaths(path.dirname(filename));
  const realRequire = mod.require.bind(mod);
  mod.require = (id) => {
    if (id === '@umijs/max')
      return { request: async () => JSON.parse(JSON.stringify(response)) };
    if (id === '@/utils/intl')
      return {
        intl: { formatMessage: ({ defaultMessage }) => defaultMessage },
      };
    if (id === '@/utils/helper')
      return {
        floorToTwoDecimalPlaces: (value) => Math.floor(value * 100) / 100,
      };
    if (id === '@/utils/package')
      return { isGte4_2: () => true, isGte4_3_3: () => true };
    if (id === '@/components/InputNumber')
      return { __esModule: true, default: realRequire('antd').InputNumber };
    if (id === '@/hook/usePublicKey')
      return {
        encryptText: () => {
          throw Error('Unexpected encryption');
        },
      };
    if (
      id === '@/constants/datetime' ||
      id === '@/components/TopoComponent/helper' ||
      id === '@/pages/Cluster/Detail/Overview/helper' ||
      id === '@/utils/eventTime'
    )
      return {};
    return realRequire(id);
  };
  mod._compile(
    ts.transpileModule(fs.readFileSync(filename, 'utf8'), {
      compilerOptions: {
        module: ts.ModuleKind.CommonJS,
        jsx: ts.JsxEmit.ReactJSX,
        esModuleInterop: true,
      },
    }).outputText,
    filename,
  );
  return mod.exports;
}
const { getEssentialParameters } = load('services/index.ts');
const { findMinParameter, getOriginResourceUsages, formatNewTenantForm } = load(
  'pages/Tenant/helper.ts',
);
const ZoneItem = load('pages/Tenant/ZoneItem/index.tsx').default;
const gib = 2 ** 30;
const resource = (availableLogDisk, logDiskUnlimited = false) => ({
  obZone: 'zone1',
  availableCPU: 4,
  availableMemory: 8 * gib,
  availableDataDisk: 10 * gib,
  availableLogDisk,
  logDiskUnlimited,
});
const essentials = (zones) => ({
  minPoolMemory: 2 * gib,
  obServerResources: Object.values(zones),
  obZoneResourceMap: zones,
});

test('API conversion preserves shared-log capacity semantics and finite GiB values', async () => {
  response = {
    successful: true,
    data: essentials({
      unlimited: resource(0, true),
      finite: resource(13 * gib),
    }),
  };
  const { data } = await getEssentialParameters({
    ns: 'test',
    name: 'cluster',
  });
  assert.equal(data.obZoneResourceMap.unlimited.availableLogDisk, Infinity);
  assert.equal(data.obServerResources[0].availableLogDisk, Infinity);
  assert.equal(data.obZoneResourceMap.finite.availableLogDisk, 13);
  assert.equal(data.obServerResources[1].availableMemory, 8);
  assert.equal(data.minPoolMemory, 2);
});

test('edit adds the existing allocation without turning an unlimited capacity into a number', () => {
  const original = essentials({
    zone1: resource(Infinity, true),
    zone2: resource(13),
  });
  const edited = getOriginResourceUsages(original, {
    zone: 'zone1',
    minCPU: 2,
    memorySize: 4,
    logDiskSize: 4,
  });
  assert.equal(edited.obZoneResourceMap.zone1.availableLogDisk, Infinity);
  assert.equal(original.obZoneResourceMap.zone1.availableCPU, 4);
  const finite = getOriginResourceUsages(original, {
    zone: 'zone2',
    minCPU: 2,
    memorySize: 4,
    logDiskSize: 4,
  });
  assert.equal(finite.obZoneResourceMap.zone2.availableLogDisk, 17);
});

test('multi-zone limit is finite when any selected zone has a finite limit', () => {
  const data = essentials({
    unlimited: resource(Infinity, true),
    finite: resource(13),
    empty: resource(0),
  });
  assert.equal(findMinParameter(['unlimited'], data).maxLogDisk, Infinity);
  assert.equal(findMinParameter(['unlimited', 'finite'], data).maxLogDisk, 13);
  assert.equal(findMinParameter(['finite', 'unlimited'], data).maxLogDisk, 13);
  assert.equal(findMinParameter(['unlimited', 'empty'], data).maxLogDisk, 0);
});

for (const isEdit of [false, true]) {
  test(`zone resource rendering hides the shared-log sentinel (edit=${isEdit})`, () => {
    const html = renderToStaticMarkup(
      React.createElement(ZoneItem, {
        name: 'zone1',
        type: 'new',
        obVersion: '4.4.0',
        checked: true,
        isEdit,
        obZoneResource: resource(Infinity, true),
        checkBoxOnChange: () => {},
      }),
    );
    assert(html.includes('无本地日志盘限制'));
    assert(!html.includes('Infinity'));
    assert(!html.includes('8589934589'));
  });
}

test('finite capacity renders its numeric GB value', () => {
  const html = renderToStaticMarkup(
    React.createElement(ZoneItem, {
      name: 'zone1',
      type: 'new',
      obVersion: '4.4.0',
      checked: true,
      obZoneResource: resource(13),
      checkBoxOnChange: () => {},
    }),
  );
  assert(html.includes('13GB'));
  assert(!html.includes('无本地日志盘限制'));
});

test('tenant payload contains the user allocation, never the unlimited UI capacity', () => {
  const result = formatNewTenantForm(
    { unitConfig: { cpuCount: 2, memorySize: 4, logDiskSize: 8 } },
    'cluster',
    '',
  );
  assert.equal(result.unitConfig.logDiskSize, '8Gi');
  assert(!JSON.stringify(result).includes('Infinity'));
});
