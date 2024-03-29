import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  changeTenantPassword,
  changeTenantRole,
  createObtenantPool,
  createTenant,
  deleteObtenantPool,
  deleteTenent,
  patchObtenantPool,
  patchTenantConfiguration,
  replayLogOfTenant,
} from '../tenant';

export async function createTenantReportWrap({
  ...params
}: API.TenantBody): Promise<API.CommonResponse> {
  const r = await createTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['createTenant'], data: r.data });
  }
  return r;
}

export async function deleteTenantReportWrap({
  ...params
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  const r = await deleteTenent(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteTenant'], data: r.data });
  }
  return r;
}

export async function modifyUnitNumReportWrap({
  ...params
}: API.NamespaceAndName &
  API.PatchTenantConfiguration): Promise<API.CommonResponse> {
  const r = await patchTenantConfiguration(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['modifyUnitNum'],

      data: r.data,
    });
  }
  return r;
}

export async function createObtenantPoolReportWrap({
  ...params
}: API.PoolConfig &
  API.NamespaceAndName & {
    zoneName: string;
  }): Promise<API.CommonResponse> {
  const r = await createObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['addPool'],

      data: r.data,
    });
  }
  return r;
}

export async function patchObtenantPoolReportWrap({
  ...params
}: API.PoolConfig &
  API.NamespaceAndName & {
    zoneName: string;
  }): Promise<API.CommonResponse> {
  const r = await patchObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['editPool'],

      data: r.data,
    });
  }
  return r;
}

export async function deleteObtenantPoolReportWrap({
  ...params
}: API.NamespaceAndName & {
  zoneName: string;
}): Promise<API.CommonResponse> {
  const r = await deleteObtenantPool(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['deletePool'],

      data: r.data,
    });
  }
  return r;
}

export async function changeTenantPasswordReportWrap({
  ...params
}: API.NamespaceAndName & API.UserCredentials): Promise<API.CommonResponse> {
  const r = await changeTenantPassword(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['modifyRootPwd'],

      data: r.data,
    });
  }
  return r;
}

export async function activateTenantReportWrap({
  ...params
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  const r = await changeTenantRole(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['activateStandby'],

      data: r.data,
    });
  }
  return r;
}

export async function switchRoleReportWrap({
  ...params
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  const r = await changeTenantRole(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['Switchover'],

      data: r.data,
    });
  }
  return r;
}

export async function replayLogTenantReportWrap({
  ...params
}: API.NamespaceAndName & API.ReplayLogType): Promise<API.CommonResponse> {
  const r = await replayLogOfTenant(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['Switchover'],

      data: r.data,
    });
  }
  return r;
}
