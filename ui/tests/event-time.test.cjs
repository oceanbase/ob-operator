const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const dayjs = require('dayjs');
const source = fs.readFileSync(path.join(__dirname, '../src/utils/eventTime.ts'), 'utf8');
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, esModuleInterop: true } }).outputText;
const context = { exports: {}, require };
vm.runInNewContext(compiled, context);
const { formatEventTime } = context.exports;

test('missing, zero, year-0001 and invalid event timestamps render a placeholder', () => {
  for (const timestamp of [undefined, null, 0, -62135596800, NaN, Infinity, 1e20]) {
    assert.equal(formatEventTime(timestamp), '-');
  }
});
test('valid event timestamps retain seconds precision in the browser timezone', () => {
  const timestamp = Date.parse('2026-09-20T07:00:00Z') / 1000;
  assert.equal(formatEventTime(timestamp), dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss'));
});
