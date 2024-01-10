import { request } from '@umijs/max';

const tenantPrefix = '/api/v1/obtenant';

export async function getAllTenants(
  obcluster: string,
): Promise<API.TenantsListResponse> {
  return request(`/api/v1/obtenants?obcluster=${obcluster}`, {
    method: 'GET',
  });
}

export async function getTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.TenantInfoType> {
  return request(`${tenantPrefix}/${ns}/${name}`, {
    method: 'GET',
  });
}

export async function createTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.TenantBody): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}`, {
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
// 创建特定租户的备份策略，密码应采用AES加密 没有body??
export async function createPolicyOfTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'PUT',
  });
}

export async function UpdatePolicyOfTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.TenantPolicy): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'POST',
    data: body,
  });
}

export async function DeletePolicyOfTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/backupPolicy`, {
    method: 'DELETE',
  });
}

// 备租户回放日志
export async function ReplayLogOfTenant({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.ReplayLogType): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/logreplay`, {
    method: 'POST',
    data: body,
  });
}

export async function changeTenantRole({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/role`, {
    method: 'POST',
  });
}
export async function changeTenantPassword({
  ns,
  name,
  ...body
}: API.NamespaceAndName & API.RootPassword): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/rootPassword`, {
    method: 'PUT',
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

// 升级特定租户的租户兼容版本以匹配集群版本
export async function upgradeTenantCompatibilityVersion({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/version`, {
    method: 'POST',
  });
}

export async function modifyUnitConfig({
  ns,
  name,
  zone,
  ...body
}: API.NamespaceAndName & {
  zone: string;
} & API.UnitConfig): Promise<API.CommonResponse> {
  return request(`${tenantPrefix}/${ns}/${name}/${zone}`, {
    method: 'PUT',
    data: body,
  });
}
