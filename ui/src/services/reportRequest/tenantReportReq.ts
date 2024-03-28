import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  changeTenantPassword,
  createObtenantPool,
  createTenant,
  deleteObtenantPool,
  deleteTenent,
  patchObtenantPool,
  patchTenantConfiguration,
  changeTenantRole,
  replayLogOfTenant
} from '../tenant';

export async function createTenantReportWrap({
  version,
  ...params
}: API.TenantBody & { version: string }): Promise<API.CommonResponse> {
  const r = await createTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['createTenant'], version, data: r.data });
  }
  return r;
}

export async function deleteTenantReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string }): Promise<API.CommonResponse> {
  const r = await deleteTenent(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteTenant'], version, data: r.data });
  }
  return r;
}

export async function modifyUnitNumReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.PatchTenantConfiguration & {
    version: string;
  }): Promise<API.CommonResponse> {
  const r = await patchTenantConfiguration(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['modifyUnitNum'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function createObtenantPoolReportWrap({
  version,
  ...params
}: API.PoolConfig &
  API.NamespaceAndName & {
    zoneName: string;
    version: string;
  }): Promise<API.CommonResponse> {
  const r = await createObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['addPool'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function patchObtenantPoolReportWrap({
  version,
  ...params
}: API.PoolConfig &
  API.NamespaceAndName & {
    zoneName: string;
    version: string;
  }): Promise<API.CommonResponse> {
  const r = await patchObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['editPool'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function deleteObtenantPoolReportWrap({
  version,
  ...params
}: API.NamespaceAndName & {
  zoneName: string;
  version: string;
}): Promise<API.CommonResponse> {
  const r = await deleteObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['deletePool'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function changeTenantPasswordReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.UserCredentials & { version: string }): Promise<API.CommonResponse> {
  const r = await changeTenantPassword(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['modifyRootPwd'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function activateTenantReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string }): Promise<API.CommonResponse> {
  const r = await changeTenantRole(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['activateStandby'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function switchRoleReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string }): Promise<API.CommonResponse> {
  const r = await changeTenantRole(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['Switchover'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function replayLogTenantReportWrap({
  version,
  ...params
}: API.NamespaceAndName & API.ReplayLogType & { version: string }): Promise<API.CommonResponse> {
  const r = await replayLogOfTenant(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['Switchover'],
      version,
      data: r.data,
    });
  }
  return r;
}
