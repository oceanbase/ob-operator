const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../src/pages/SharedStorage/storageConfig.ts'), 'utf8');
const context = { exports: {}, URL };
vm.runInNewContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, context);
const { buildBucketURL, parseBucketURL, sameStorageLocation, storageLocationsOverlap, bucketURLRules } = context.exports;
const raw = 's3://sharedstorage/cluster-a?host=http://192.0.2.10:9000&s3_region=us-east-1';

test('structured S3 fields round-trip with the OceanBase raw host grammar', () => {
  const fields = parseBucketURL(raw);
  assert.equal(fields.endpoint, 'http://192.0.2.10:9000');
  assert.equal(fields.bucket, 'sharedstorage');
  assert.equal(fields.prefix, 'cluster-a');
  assert.equal(buildBucketURL(fields), raw);
  assert(!buildBucketURL(fields).includes('%3A'));
});
test('credentials, unsupported protocols, duplicate options and unsafe paths are rejected', async () => {
  for (const value of ['', raw + '&access_key=secret', raw + '&host=http://other', raw.replace('s3://', 'oss://'), raw.replace('s3://', 's3://id:secret@'), raw.replace('http://192.0.2.10:9000', 'http://id:secret@host:9000'), raw.replace('/cluster-a?', '/../a?'), raw.replace('/cluster-a?', '/%2e%2e/a?'), raw.replace('9000', '99999'), raw + '&extra=value']) {
    assert.equal(parseBucketURL(value), undefined, value);
    await assert.rejects(bucketURLRules[0].validator(null, value));
  }
});
test('separate data/log locations can share an endpoint without becoming the same location', () => {
  assert.equal(sameStorageLocation(raw, raw), true);
  assert.equal(sameStorageLocation(raw, raw.replace('sharedstorage', 'logservice')), false);
  assert.equal(sameStorageLocation(raw, raw.replace('cluster-a', 'cluster-b')), false);
  assert.equal(sameStorageLocation(raw, undefined), false);
});
test('overlap detection includes root and nested prefixes, but not sibling prefix names', () => {
  assert.equal(storageLocationsOverlap(raw, raw.replace('/cluster-a?', '?')), true);
  assert.equal(storageLocationsOverlap(raw, raw.replace('/cluster-a?', '/cluster-a/logs?')), true);
  assert.equal(storageLocationsOverlap(raw, raw.replace('/cluster-a?', '/cluster-ab?')), false);
  assert.equal(storageLocationsOverlap(raw, raw.replace('sharedstorage', 'logservice')), false);
});
