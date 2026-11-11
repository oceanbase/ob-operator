import { request } from '@umijs/max';

export type Zone = { zone: string; replica: number; rpcPort: number; httpPort: number };
export type LSSpec = {
  clusterId: number;
  logService: { image: string; resource: { cpu: string; memory: string }; storage: {
    storeStorage: { size: string; storageClass?: string }; logStorage: { size: string; storageClass?: string };
  } };
  topology: Zone[];
  objectStoreUrl: { bucketURL: string; secretRef: { name: string } };
  parameters?: { name: string; value: string }[];
};
export type LSItem = {
  name: string; namespace: string; resourceVersion: string; createdAt: string;
  deleting: boolean; protected: boolean; references: string[]; spec: LSSpec;
  status: { status: string; operationContext?: { task?: string; taskStatus?: string } };
};
export type LSDetail = LSItem & {
  nodes: { name: string; zone: string; deleting: boolean; status: {
    status: string; ready: boolean; podName?: string; podIP?: string; serviceIP?: string; podPhase?: string;
  } }[];
  volumes: { name: string; phase: string; size: string }[];
  events: { time: string; type: string; reason: string; message: string; object: string }[];
};
export async function lsRequest<T>(path = '', options: Record<string, unknown> = {}): Promise<T> {
  const res = await request<{ data: T; successful: boolean; message?: string }>(`/api/v1/logservices${path}`, options);
  if (!res.successful) throw new Error(res.message || 'LogService request failed');
  return res.data;
}
export const lsPath = (ns: string, name: string) => `/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`;
