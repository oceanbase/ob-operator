const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const vm = require('node:vm');
const context = { exports: {}, URL };
vm.runInNewContext(
  ts.transpileModule(
    fs.readFileSync(
      path.join(__dirname, '../src/pages/SharedStorage/storageConfig.ts'),
      'utf8',
    ),
    { compilerOptions: { module: ts.ModuleKind.CommonJS } },
  ).outputText,
  context,
);
const {
  parseBucketURL,
  buildBucketURL,
  bucketURLRules,
  storageLocationsOverlap,
} = context.exports;
const base =
  's3://data-bucket/?host=https://s3.example.com&s3_region=cn-wulanchabu';
const options = '&scope=region1&max_iops=10000&max_bandwidth=1GB';

test('OceanBase URL options parse and survive structured field edits', async () => {
  const fields = parseBucketURL(base + options);
  assert(fields);
  assert.equal(fields.endpoint, 'https://s3.example.com');
  assert.equal(fields.bucket, 'data-bucket');
  assert.equal(fields.region, 'cn-wulanchabu');
  assert.equal(fields.prefix, '');
  assert.equal(fields.scope, 'region1');
  assert.equal(fields.maxIOPS, '10000');
  assert.equal(fields.maxBandwidth, '1GB');
  const edited = buildBucketURL({ ...fields, prefix: 'new/path' });
  assert.equal(edited, base.replace('/?', '/new/path?') + options);
  await bucketURLRules[0].validator(null, edited);
  assert.equal(storageLocationsOverlap(edited, base), true);
});

test('options are optional and preserve zero and numeric bandwidth', () => {
  for (const option of [
    'scope=region1',
    'max_iops=0',
    'max_bandwidth=0B',
    'max_bandwidth=1024',
    'max_bandwidth=1.5GB',
  ]) {
    const value = base + '&' + option;
    const fields = parseBucketURL(value);
    assert(fields, option);
    assert.equal(buildBucketURL(fields), value.replace('/?', '?'));
  }
});

test('known options do not permit credentials, unknown keys, duplicates or malformed values', async () => {
  for (const option of [
    'access_key=TOPSECRET',
    'access_id=TOPSECRET',
    'unknown=x',
    'scope=region1&scope=region2',
    'scope=',
    'scope=region1%0A',
    'max_iops=10000%0A',
    'max_bandwidth=1GB%0A',
    'scope=region%26access_key%3DTOPSECRET',
    'max_iops=-1',
    'max_iops=1.2',
    'max_iops=1&max_iops=2',
    'max_bandwidth=',
    'max_bandwidth=fast',
    'max_bandwidth=1GB&max_bandwidth=2GB',
  ]) {
    const raw = base + '&' + option;
    assert.equal(parseBucketURL(raw), undefined, option);
    await assert.rejects(bucketURLRules[0].validator(null, raw));
  }
  assert.equal(
    parseBucketURL(base + options + '&access_key=TOPSECRET'),
    undefined,
  );
});
