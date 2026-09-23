export type StorageLocation = {
  endpoint: string;
  bucket: string;
  prefix: string;
  region: string;
  scope?: string;
  maxIOPS?: string;
  maxBandwidth?: string;
};
export const emptyLocation = (): StorageLocation => ({
  endpoint: '',
  bucket: '',
  prefix: '',
  region: 'us-east-1',
});

// Supported non-credential OceanBase options must survive structured edits.
const storageOptions = [
  ['scope', 'scope', /^[a-zA-Z0-9_-]{1,64}$/],
  ['max_iops', 'maxIOPS', /^[0-9]{1,20}$/],
  ['max_bandwidth', 'maxBandwidth', /^[0-9]+(?:\.[0-9]+)?(?:[KMGTPE]?B)?$/i],
] as const;

// OceanBase expects the raw host=http://... grammar, not a URL-encoded host.
export function buildBucketURL(v: StorageLocation): string {
  let result = `s3://${v.bucket.trim()}${
    v.prefix ? `/${v.prefix.replace(/^\/+/, '')}` : ''
  }?host=${v.endpoint.trim().replace(/\/$/, '')}&s3_region=${v.region.trim()}`;
  for (const [query, field] of storageOptions) {
    if (v[field] !== undefined)
      result += `&${query}=${encodeURIComponent(v[field]!)}`;
  }
  return result;
}
export function parseBucketURL(raw: string): StorageLocation | undefined {
  try {
    const u = new URL(raw);
    const q = u.searchParams;
    if (
      raw.length > 2048 ||
      u.protocol !== 's3:' ||
      u.username ||
      u.password ||
      u.hash ||
      u.port ||
      !/^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$/.test(u.hostname) ||
      u.hostname.includes('..') ||
      /^\d+\.\d+\.\d+\.\d+$/.test(u.hostname)
    )
      return;
    if (
      q.getAll('host').length !== 1 ||
      q.getAll('s3_region').length !== 1 ||
      !/^[a-zA-Z0-9-]{1,64}$/.test(q.get('s3_region') || '')
    )
      return;
    for (const key of Array.from(q.keys())) {
      if (key === 'host' || key === 's3_region') continue;
      const option = storageOptions.find(([name]) => name === key);
      const values = q.getAll(key);
      if (
        !option ||
        values.length !== 1 ||
        values[0].length > 64 ||
        values[0].trim() !== values[0] ||
        !option[2].test(values[0])
      )
        return;
    }
    const host = q.get('host')!;
    const e = new URL(host);
    if (
      !/^https?:\/\/[^\s/?#]+\/?$/.test(host) ||
      !['http:', 'https:'].includes(e.protocol) ||
      e.username ||
      e.password ||
      e.search ||
      e.hash ||
      e.pathname !== '/' ||
      /[\\%]/.test(host)
    )
      return;
    const prefix = decodeURIComponent(u.pathname).replace(/^\//, '');
    if (
      !/^[a-zA-Z0-9_./-]*$/.test(prefix) ||
      prefix.split('/').some((p) => p === '.' || p === '..')
    )
      return;
    // WHATWG URL normalizes dot segments: reject those in the raw path too.
    if (/\/(?:\.|%2e){1,2}(?:\/|\?)/i.test(raw)) return;
    const result: StorageLocation = {
      endpoint: host.replace(/\/$/, ''),
      bucket: u.hostname,
      prefix,
      region: q.get('s3_region')!,
    };
    for (const [query, field] of storageOptions) {
      if (q.has(query)) result[field] = q.get(query)!;
    }
    return result;
  } catch {
    return;
  }
}

export const bucketURLRules = [
  {
    required: true,
    validator: async (_: unknown, value: string) => {
      if (!value || !parseBucketURL(value))
        throw new Error(
          '请填写有效的 S3 Endpoint、Bucket、Region；仅支持 scope、max_iops、max_bandwidth 附加参数，不支持内嵌密钥 / Invalid S3 configuration',
        );
    },
  },
];

export function sameStorageLocation(a?: string, b?: string): boolean {
  const x = a && parseBucketURL(a),
    y = b && parseBucketURL(b);
  return (
    !!x &&
    !!y &&
    x.endpoint === y.endpoint &&
    x.bucket === y.bucket &&
    x.prefix.replace(/\/$/, '') === y.prefix.replace(/\/$/, '')
  );
}

export function storageLocationsOverlap(a?: string, b?: string): boolean {
  const x = a && parseBucketURL(a),
    y = b && parseBucketURL(b);
  if (
    !x ||
    !y ||
    new URL(x.endpoint).origin !== new URL(y.endpoint).origin ||
    x.bucket !== y.bucket
  )
    return false;
  const p = x.prefix.replace(/\/+$/, ''),
    q = y.prefix.replace(/\/+$/, '');
  return !p || !q || p === q || p.startsWith(q + '/') || q.startsWith(p + '/');
}
