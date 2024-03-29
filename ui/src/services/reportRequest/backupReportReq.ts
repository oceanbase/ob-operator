import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  createBackupPolicyOfTenant,
  deletePolicyOfTenant,
  updateBackupPolicyOfTenant,
} from '../tenant';

export async function createBackupReportWrap({
  ...params
}: API.NamespaceAndName & API.TenantPolicy): Promise<API.CommonResponse> {
  const r = await createBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['createBackup'], data: r.data });
  }
  return r;
}

export async function deleteBackupReportWrap({
  ...params
}: API.NamespaceAndName): Promise<API.CommonResponse> {
  const r = await deletePolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteBackup'], data: r.data });
  }
  return r;
}

export async function editBackupReportWrap({
  ...params
}: API.NamespaceAndName & API.UpdateTenantPolicy): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['editBackup'], data: r.data });
  }
  return r;
}

export async function pauseBackupReportWrap({
  ...params
}: API.NamespaceAndName & API.UpdateTenantPolicy): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['pauseBackup'], data: r.data });
  }
  return r;
}

export async function resumeBackupReportWrap({
  ...params
}: API.NamespaceAndName & API.UpdateTenantPolicy): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['resumeBackup'], data: r.data });
  }
  return r;
}
