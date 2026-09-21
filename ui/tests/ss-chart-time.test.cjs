const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const { Time } = require('@antv/scale');

const source = fs.readFileSync(path.join(__dirname, '../src/pages/SharedStorage/chartOptions.ts'), 'utf8');
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText;
const context = { exports: {}, Date };
vm.runInNewContext(compiled, context);
const { buildSSChartOptions } = context.exports;

for (const minutes of [15, 60, 360, 1440]) {
  test(`${minutes}-minute chart formats actual G2 time ticks and tooltip without reparsing labels`, () => {
    const start = Date.parse('2026-09-20T07:00:00Z') / 1000;
    const points = Array.from({ length: 121 }, (_, i) => ({ time: start + i * minutes / 2, value: 3 }));
    const options = buildSSChartOptions(points);
    assert.equal(options.data[0].time, start * 1000);
    assert.equal(points[0].time, start, 'does not mutate API samples');
    const scale = new Time({ ...options.meta.time, values: options.data.map(p => p.time) });
    assert.equal(scale.min, start * 1000);
    const ticks = scale.getTicks();
    assert.ok(ticks.length > 1);
    for (const tick of ticks) assert.match(tick.text, /^\d{2}:\d{2}:\d{2}$/);
    const formatted = scale.getText(options.data[0].time);
    const title = options.tooltip.title(formatted, options.data[0]);
    assert.equal(title, new Date(start * 1000).toLocaleString());
    assert.equal(options.xAxis, undefined, 'no second formatter receives the formatted tick text');
    assert.equal(options.data[0].value, 3);
  });
}

test('empty and invalid samples do not produce fake zeroes or invalid dates', () => {
  assert.equal(buildSSChartOptions([]).data.length, 0);
  const points = [
    { time: NaN, value: 3 }, { time: Infinity, value: 3 },
    { time: 0, value: 3 }, { time: 1e20, value: 3 },
    { time: 1790002800, value: NaN }, { time: 1790002800, value: Infinity },
    { time: 1790002800.125, value: 0 },
  ];
  const options = buildSSChartOptions(points);
  assert.equal(options.data.length, 1);
  assert.equal(options.data[0].time, 1790002800125);
  assert.equal(options.data[0].value, 0, 'a real zero is retained');
  assert.equal(options.tooltip.title('Invalid Date', {}), '-');
});
