const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');

// Use the project's installed TypeScript compiler; no additional test dependency.
const source = fs.readFileSync(path.join(__dirname, '../src/pages/Cluster/New/storageMode.ts'), 'utf8');
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText;
const context = { exports: {} };
vm.runInNewContext(compiled, context);
const { normalizeStorageMode } = context.exports;
const create = (deploymentMode) => ({ deploymentMode, mode: 'SERVICE', observer: { image: 'ai:4.6.2.0', storage: { data: { size: 50 }, log: { size: 20 }, redoLog: { size: 50 } } }, sharedStorageInfo: { bucketURL: 's3://test', secretRef: { name: 'minio' } }, logServiceRef: { name: 'ls' } });

test('SS retains references and omits stale redoLog without mutating the form', () => {
  const input = create('shared_storage');
  const output = normalizeStorageMode(input);
  assert.equal(output.observer.storage.redoLog, undefined);
  assert.equal(output.sharedStorageInfo, input.sharedStorageInfo);
  assert.equal(output.logServiceRef, input.logServiceRef);
  assert.equal(output.mode, 'SERVICE');
  assert.equal(output.observer.image, 'ai:4.6.2.0');
  assert.equal(input.observer.storage.redoLog.size, 50);
});

for (const mode of ['normal', undefined]) {
  test(`normal/legacy ${mode} clears SS fields and preserves redoLog`, () => {
    const input = create(mode);
    const output = normalizeStorageMode(input);
    assert.equal(output.sharedStorageInfo, undefined);
    assert.equal(output.logServiceRef, undefined);
    assert.equal(output.observer.storage.redoLog.size, 50);
    assert.equal(input.logServiceRef.name, 'ls');
  });
}
