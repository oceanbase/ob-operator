const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const ctx = {exports:{}};
vm.runInNewContext(ts.transpileModule(fs.readFileSync(path.join(__dirname, '../src/pages/Cluster/New/sharedStorageSizing.ts'),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText,ctx);
test('SS minimum cache avoids the insufficient 6 GiB initialization from the normal preset',()=>{
  assert.equal(ctx.exports.minimumDataGiB(true,30),50);
  assert.equal(ctx.exports.minimumDataGiB(false,30),30);
});
