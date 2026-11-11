const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const test = require('node:test');
const ts = require('typescript');

const filename = path.join(
  __dirname,
  '../src/components/MonitorDetail/index.tsx',
);
const source = fs.readFileSync(filename, 'utf8');

let refreshEnabled = false;
let hookIndex = 0;
let refs = [];
let nextRefreshEffect;

const react = {
  useEffect: () => {},
  useRef: (initialValue) => {
    const index = hookIndex++;
    if (!refs[index]) refs[index] = { current: initialValue };
    return refs[index];
  },
  useState: (initialValue) => {
    const index = hookIndex++;
    if (index === 0) return [refreshEnabled, () => {}];
    return [
      typeof initialValue === 'function' ? initialValue() : initialValue,
      () => {},
    ];
  },
};

const componentModule = new Module(filename, module);
componentModule.filename = filename;
componentModule.paths = Module._nodeModulePaths(path.dirname(filename));
componentModule.require = (id) => {
  if (id === 'react') return react;
  if (id === 'react/jsx-runtime') {
    const element = (type, props) => ({ type, props });
    return { Fragment: Symbol('Fragment'), jsx: element, jsxs: element };
  }
  if (id === '@umijs/max')
    return { useAccess: () => ({ obclusterwrite: true }) };
  if (id === 'ahooks') {
    return {
      useUpdateEffect: (effect) => {
        nextRefreshEffect = effect;
      },
    };
  }
  if (id === 'dayjs') {
    return {
      __esModule: true,
      default: { unix: () => ({ format: () => 'now' }) },
    };
  }
  if (id === '@/components/MonitorComp' || id === './DataFilter') {
    return { __esModule: true, default: () => null };
  }
  if (id === '@/constants') {
    return {
      DEFAULT_QUERY_RANGE: { step: 1, startTimestamp: 1, endTimestamp: 2 },
      REFRESH_FREQUENCY: 30,
    };
  }
  if (id === '@/constants/datetime')
    return { DATE_TIME_FORMAT: 'YYYY-MM-DD HH:mm:ss' };
  throw new Error(`Unexpected module in MonitorDetail test: ${id}`);
};
componentModule._compile(
  ts.transpileModule(source, {
    compilerOptions: {
      esModuleInterop: true,
      jsx: ts.JsxEmit.ReactJSX,
      module: ts.ModuleKind.CommonJS,
    },
  }).outputText,
  filename,
);
const MonitorDetail = componentModule.exports.default;

function render(enabled) {
  refreshEnabled = enabled;
  hookIndex = 0;
  nextRefreshEffect = undefined;
  MonitorDetail({
    filterData: {},
    setFilterData: () => {},
    basicInfo: null,
    filterLabel: [],
    setFilterLabel: () => {},
    groupLabels: [],
    queryScope: 'obtenant',
  });
  assert.equal(typeof nextRefreshEffect, 'function');
  return nextRefreshEffect;
}

test('chart refresh timer is recreated after refresh is disabled and enabled again', () => {
  const originalSetInterval = global.setInterval;
  const originalClearInterval = global.clearInterval;
  const createdTimers = [];
  const clearedTimers = [];
  let cleanup;

  global.setInterval = (callback) => {
    const timer = { id: createdTimers.length + 1, callback };
    createdTimers.push(timer);
    return timer;
  };
  global.clearInterval = (timer) => {
    if (timer) clearedTimers.push(timer);
  };

  try {
    render(false); // useUpdateEffect skips the initial mount.
    cleanup = render(true)();
    cleanup();
    cleanup = render(false)();
    cleanup();
    cleanup = render(true)();

    assert.equal(
      createdTimers.length,
      2,
      're-enabling refresh must create a new chart timer',
    );
    assert.notEqual(createdTimers[0], createdTimers[1]);
    assert(
      clearedTimers.includes(createdTimers[0]),
      'the first timer is cleared when refresh is disabled',
    );
  } finally {
    if (cleanup) cleanup();
    global.setInterval = originalSetInterval;
    global.clearInterval = originalClearInterval;
    refs = [];
  }
});
