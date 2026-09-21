import { request } from '@umijs/max';

export type CredentialReference = { name: string; namespace: string };
export type StorageCheck = {
  ok: boolean;
  code: string;
  checkedAt: string;
  httpStatus?: number;
  scope: string;
};
export async function storageRequest<T>(
  namespace: string,
  action: 'credentials' | 'check',
  options: Record<string, unknown> = {},
): Promise<T> {
  const r = await request<{ successful: boolean; message?: string; data: T }>(
    `/api/v1/objectstorage/${encodeURIComponent(namespace)}/${action}`,
    options,
  );
  if (!r.successful)
    throw new Error(r.message || 'Object storage request failed');
  return r.data;
}
