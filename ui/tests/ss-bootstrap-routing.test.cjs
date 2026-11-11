const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
function load(file) {
  const context = { exports: {} };
  vm.runInNewContext(ts.transpileModule(fs.readFileSync(path.join(__dirname, '../src', file), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, context);
  return context.exports;
}
const { validBootstrapReplicas } = load('pages/LogService/bootstrap.ts');
const { initialRouteRedirect } = load('pages/Layouts/BasicLayout/initialRoute.ts');
test('LogService bootstrap requires exactly three positive integer replicas', () => {
  for (const topology of [undefined, [], [{ replica: 1 }], [{ replica: 2 }], [{ replica: 4 }], [{ replica: 1.5 }, { replica: 1.5 }], [{ replica: 0 }, { replica: 3 }], [{ replica: NaN }]]) assert.equal(validBootstrapReplicas(topology), false);
  assert.equal(validBootstrapReplicas([{ replica: 3 }]), true);
  assert.equal(validBootstrapReplicas([{ replica: 1 }, { replica: 1 }, { replica: 1 }]), true);
});
test('layout preserves authorized deep links, including queries', () => {
  const links = ['/overview', '/alert', '/cluster', '/logservice'];
  for (const route of ['/alert', '/alert/rules', '/alert/event', '/alert/rules?from=ls', '/cluster/new?deploymentMode=shared_storage', '/logservice/ns/ls']) assert.equal(initialRouteRedirect(route, links), undefined);
});
test('layout rejects substring/query matches and denied menu paths', () => {
  for (const route of ['/x/alert/rules', '/alerting', '/x?target=/alert', '/cluster/new']) assert.equal(initialRouteRedirect(route, ['/overview', '/alert']), '/overview');
  assert.equal(initialRouteRedirect('/alert/rules', []), '/overview');
});
