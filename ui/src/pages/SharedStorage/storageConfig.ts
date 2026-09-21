export type StorageLocation = {
  endpoint: string;
  bucket: string;
  prefix: string;
  region: string;
};
export const emptyLocation = (): StorageLocation => ({
  endpoint: '',
  bucket: '',
  prefix: '',
  region: 'us-east-1',
});

// OceanBase expects the raw host=http://... grammar, not a URL-encoded host.
export function buildBucketURL(v: StorageLocation): string {
  return `s3://${v.bucket.trim()}${
    v.prefix ? `/${v.prefix.replace(/^\/+/, '')}` : ''
  }?host=${v.endpoint.trim().replace(/\/$/, '')}&s3_region=${v.region.trim()}`;
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
      Array.from(q.keys()).length !== 2 ||
      q.getAll('host').length !== 1 ||
      q.getAll('s3_region').length !== 1 ||
      !/^[a-zA-Z0-9-]{1,64}$/.test(q.get('s3_region') || '')
    )
      return;
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
    return {
      endpoint: host.replace(/\/$/, ''),
      bucket: u.hostname,
      prefix,
      region: q.get('s3_region')!,
    };
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
          '请填写有效的 S3 Endpoint、Bucket、Region；不支持内嵌密钥或额外 URL 参数 / Invalid S3 configuration',
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
