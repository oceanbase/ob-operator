const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const test = require('node:test');
const ts = require('typescript');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');

// Execute the real Parameters page and its column callbacks. Only API/router
// hooks and the table shell are replaced; status tags use real Ant Design.
const filename = path.join(
  __dirname,
  '../src/pages/Cluster/Detail/Parameters/index.tsx',
);
let records = [];
let tableProps;
const noop = () => {};
const componentModule = new Module(filename, module);
componentModule.filename = filename;
componentModule.paths = Module._nodeModulePaths(path.dirname(filename));
const realRequire = componentModule.require.bind(componentModule);
componentModule.require = (id) => {
  if (id === '@/api')
    return {
      obcluster: { patchOBCluster: noop, listOBClusterParameters: noop },
    };
  if (id === '@umijs/max')
    return { useParams: () => ({ ns: 'test', name: 'test-cluster' }) };
  if (id === 'ahooks')
    return {
      useRequest: () => ({
        data: { data: records },
        loading: false,
        refresh: noop,
        runAsync: noop,
      }),
    };
  if (id === '@/utils/intl')
    return { intl: { formatMessage: ({ defaultMessage }) => defaultMessage } };
  if (id === '@/utils/component') return { getColumnSearchProps: () => ({}) };
  if (id === '@/components/CustomTooltip' || id === './ParametersModal')
    return { __esModule: true, default: () => null };
  if (id === '@/components/IconTip')
    return { __esModule: true, default: ({ content }) => content };
  if (id === '@ant-design/pro-components')
    return {
      PageContainer: ({ children }) =>
        React.createElement('main', null, children),
    };
  if (id === 'antd')
    return {
      ...realRequire(id),
      Table: (props) => {
        tableProps = props;
        const status = props.columns.find(
          (column) => column.dataIndex === 'status',
        );
        return React.createElement(
          'table',
          null,
          React.createElement(
            'tbody',
            null,
            props.dataSource.map((row) =>
              React.createElement(
                'tr',
                { key: row.name },
                React.createElement('td', null, row.name),
                React.createElement('td', null, status.render(row.status, row)),
              ),
            ),
          ),
        );
      },
    };
  return realRequire(id);
};
componentModule._compile(
  ts.transpileModule(fs.readFileSync(filename, 'utf8'), {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      jsx: ts.JsxEmit.ReactJSX,
      esModuleInterop: true,
    },
  }).outputText,
  filename,
);
const Parameters = componentModule.exports.default;

function render(rows) {
  records = rows;
  tableProps = undefined;
  const html = renderToStaticMarkup(React.createElement(Parameters));
  assert(tableProps, 'real Parameters page must provide table columns');
  return html;
}

function statusColumn() {
  return tableProps.columns.find((column) => column.dataIndex === 'status');
}

test('unmanaged enable_window_compaction has no comparison badge', () => {
  const html = render([
    {
      name: 'enable_window_compaction',
      isManagedByOperator: false,
      status: '',
    },
  ]);
  assert(html.includes('enable_window_compaction'));
  assert(html.includes('<span>—</span>'));
  assert(!html.includes('不匹配'));
  assert(!html.includes('已匹配'));
});

test('unmanaged or missing management flag cannot show even a stale known status', () => {
  for (const managed of [false, undefined, null]) {
    for (const status of [
      '',
      undefined,
      null,
      'matched',
      'not match',
      'pending ob',
    ]) {
      const html = render([
        { name: 'test-param', isManagedByOperator: managed, status },
      ]);
      assert(html.includes('<span>—</span>'), `${managed}/${status}`);
      assert(!html.includes('ant-tag'), `${managed}/${status}`);
      for (const filter of statusColumn().filters) {
        assert.equal(statusColumn().onFilter(filter.value, records[0]), false);
      }
    }
  }
});

test('managed matched and backend not match statuses keep their labels and colors', () => {
  for (const [status, label, color] of [
    ['matched', '已匹配', 'green'],
    ['not match', '不匹配', 'gold'],
  ]) {
    const html = render([
      { name: 'test-param', isManagedByOperator: true, status },
    ]);
    assert(html.includes(label));
    assert(html.includes(`ant-tag-${color}`));
    assert.equal(statusColumn().onFilter(status, records[0]), true);
  }
});

test('empty and transitional managed statuses are not mismatches', () => {
  for (const status of [
    '',
    undefined,
    null,
    'pending ob',
    'Failed',
    'future-status',
  ]) {
    const html = render([
      { name: 'test-param', isManagedByOperator: true, status },
    ]);
    assert(html.includes('<span>—</span>'), String(status));
    assert(!html.includes('ant-tag'), String(status));
    for (const filter of statusColumn().filters) {
      assert.equal(statusColumn().onFilter(filter.value, records[0]), false);
    }
  }
});

test('filter choices follow backend status values and isolate exact managed matches', () => {
  render([
    { name: 'unmanaged', isManagedByOperator: false, status: '' },
    { name: 'stale-unmanaged', isManagedByOperator: false, status: 'matched' },
    { name: 'matched', isManagedByOperator: true, status: 'matched' },
    { name: 'mismatch', isManagedByOperator: true, status: 'not match' },
    { name: 'pending', isManagedByOperator: true, status: 'pending ob' },
    { name: 'uninitialized', isManagedByOperator: true, status: '' },
  ]);
  const column = statusColumn();
  assert.deepEqual(
    column.filters.map((filter) => filter.value),
    ['matched', 'not match'],
  );
  assert.deepEqual(
    records
      .filter((row) => column.onFilter('matched', row))
      .map((row) => row.name),
    ['matched'],
  );
  assert.deepEqual(
    records
      .filter((row) => column.onFilter('not match', row))
      .map((row) => row.name),
    ['mismatch'],
  );
  assert.equal(
    column.onFilter('future-status', {
      isManagedByOperator: true,
      status: 'future-status',
    }),
    false,
  );
});

test('management filter remains independent of comparison status', () => {
  render([]);
  const column = tableProps.columns.find(
    (item) => item.dataIndex === 'isManagedByOperator',
  );
  assert.equal(
    column.onFilter(false, { isManagedByOperator: false, status: '' }),
    true,
  );
  assert.equal(
    column.onFilter(true, { isManagedByOperator: false, status: 'matched' }),
    false,
  );
  assert.equal(
    column.onFilter(true, { isManagedByOperator: true, status: 'not match' }),
    true,
  );
});
