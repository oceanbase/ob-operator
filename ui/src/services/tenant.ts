import { formatStatisticData,getInitialObjOfKeys } from '@/utils/helper';
import { request } from '@umijs/max';

const tenantPrefix = '/api/v1/obtenants';

export async function getTenantStatisticReq(): Promise<API.StatisticDataResponse> {
  const r = await request(`${tenantPrefix}/statistic`, {
    method: 'GET',
  });
  return {
    ...r,
    data: formatStatisticData('tenant', r.data),
  };
}

export async function getAllTenants(
  params:
    | {
        obcluster?: string;
        ns?: string;
      }
    | undefined,
): Promise<API.TenantsListResponse> {
  return request(`${tenantPrefix}`, {
    method: 'GET',
    params: params,
  });
}

export async function getTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.TenantBasicInfoResponse> {
  const r = await request(`${tenantPrefix}/${ns}/${name}`, {
    method: 'GET',
  });
  const infoKeys = [
    'charset',
    'clusterResourceName',
    'tenantName',
    'tenantRole',
    'unitNumber',
    'status',
    'name',
    'namespace',
    'locality',
    'primaryZone'
  ];
  const res: API.TenantBasicInfo = {
    info: {},
    source: {},
    replicas: [],
  };
  if (r.successful) {
    res.info = getInitialObjOfKeys(r.data, infoKeys);
    if (r.data.primaryTenant) res.source!.primaryTenant = r.data.primaryTenant;
    if (r.data.restoreSource?.archiveSource)
      res.source!.archiveSource = r.data.restoreSource.archiveSource;
    if (r.data.restoreSource?.bakDataSource)
      res.source!.bakDataSource = r.data.restoreSource.bakDataSource;
    if (r.data.restoreSource?.until)
      res.source!.until = r.data.restoreSource.until;

    res.replicas = r.data.topology;
    return {
      ...r,
      data: res,
    };
  }
  return r;
}

export async function createTenant({
  ...body
}: API.TenantBody): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}`, {
    method: 'PUT',
    data: body,
  });
}

export async function deleteTenent({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}`, {
    method: 'DELETE',
  });
}

export async function createBackupPolicyOfTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.TenantPolicy): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'PUT',
    data: body,
  });
}

export async function updateBackupPolicyOfTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName &
  API.UpdateTenantPolicy): Promise<API.BackupPolicyResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'PATCH',
    data: body,
  });
}

export async function deletePolicyOfTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'DELETE',
  });
}

export async function getBackupPolicy({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.BackupPolicyResponse> {
  const r = await request(`${tenantPrefix}/${ns}/${name}/backupPolicy`);
  const keys = [
    'destType',
    'archivePath',
    'bakDataPath',
    'scheduleType',
    'scheduleTime',
    'scheduleDates',
    'status',
    'ossAccessSecret',
    'bakEncryptionSecret',
    'jobKeepDays',
    'pieceIntervalDays',
    'recoveryDays',
  ];

  if (r.successful && r.data) {
    return {
      ...r,
      data: getInitialObjOfKeys(r.data, keys),
    };
  }
  return r;
}

export async function getBackupJobs({
  ns,
  name,
  type,
  limit,
}: API.NamespaceAndName & {
  type: API.JobType;
  limit?: number;
}): Promise<API.BackupJobsResponse> {
  const r = await request(`${tenantPrefix}/${ns}/${name}/backup/${type}/jobs`, {
    params: { limit },
  });
  let res: API.BackupJob[] = [];
  if (r.successful) {
    res = r.data?.map((job: API.BackupJob) => ({
      encryptionSecret: job.encryptionSecret,
      endTime: job.endTime.split('.')[0], // Intercept time string, accurate to seconds
      name: job.name,
      path: job.path,
      startTime: job.startTime.split('.')[0],
      status: job.status,
      statusInDatabase: job.statusInDatabase,
      type: job.type,
    }));
    return {
      ...r,
      data: res,
    };
  }
  return r;
}

// tenant replay log
export async function replayLogOfTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.ReplayLogType): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/logreplay`, {
    method: 'POST',
    data: body,
  });
}

// Used to activate tenants or switch between primary and standby tenants
export async function changeTenantRole({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.RoleReqParam ): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/role`, {
    method: 'POST',
    data: body
  });
}

export async function changeTenantPassword({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.UserCredentials): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/userCredentials`, {
    method: 'POST',
    data: body,
  });
}

export async function modifyUnitNumber({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.UnitNumber): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/unitNumber`, {
    method: 'PUT',
    data: body,
  });
}

// Upgrade the tenant-compatible version of a specific tenant to match the cluster version
export async function upgradeTenantCompatibilityVersion({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/version`, {
    method: 'POST',
  });
}

export async function patchTenantConfiguration({
  ns,
  name,
  ...body
}: API.NamespaceAndName &
  API.PatchTenantConfiguration): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}`, {
    method: 'PATCH',
    data: body,
  });
}

// Create obtenant pool
export async function createObtenantPool({
  ns,
  name,
  zoneName,
  ...body
}: API.PoolConfig &
  API.NamespaceAndName & { zoneName: string }): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/pools/${zoneName}`, {
    method: 'PUT',
    data: body,
  });
}

// Delete obtenant pool
export async function deleteObtenantPool({
  ns,
  name,
  zoneName,
}: API.NamespaceAndName & { zoneName: string }): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/pools/${zoneName}`, {
    method: 'DELETE',
  });
}

// Patch obtenant pool
export async function patchObtenantPool({
  ns,
  name,
  zoneName,
  ...body
}: API.PoolConfig &
  API.NamespaceAndName & { zoneName: string }): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/pools/${zoneName}`, {
    method: 'PATCH',
    data: body,
  });
}
