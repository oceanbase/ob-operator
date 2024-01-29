import { request } from '@umijs/max';

const tenantPrefix = '/api/v1/obtenants';

export async function getAllTenants(
  obcluster?: string,
): Promise<API.TenantsListResponse> {
  let query = '';
  if (obcluster) query = `?obcluster=${obcluster}`;
  return request(`${tenantPrefix}${query}`, {
    method: 'GET',
  });
}

export async function getTenant({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.TenantBasicInfoResponse> {
  let r = await request(`${tenantPrefix}/${ns}/${name}`, {
    method: 'GET',
  })
  let res:API.TenantBasicInfo = {
    info:{
      charset:'',
      clusterName:'',
      tenantName:'',
      tenantRole:'',
      unitNumber:0,
      status:'',
      name:'',
      namespace:'',
      locality:''
    },
    source:{}
  };
  if(r.successful){
    Object.keys(res.info).forEach((key)=>{
      res.info[key] = r.data[key]
    }) 
    if(r.data.primaryTenant)res.source!.primaryTenant = r.data.primaryTenant;
    if(r.data.restoreSource?.archiveSource)res.source!.archiveSource = r.data.restoreSource.archiveSource;
    if(r.data.restoreSource?.bakDataSource)res.source!.bakDataSource = r.data.restoreSource.bakDataSource;
    if(r.data.restoreSource?.until)res.source!.until = r.data.restoreSource.until;
    return {
      ...r,
      data:res
    }
  }
  return r
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
