import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  createBackupPolicyOfTenant,
  deletePolicyOfTenant,
  updateBackupPolicyOfTenant,
} from '../tenant';

export async function createBackupReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.TenantPolicy & { version: string }): Promise<API.CommonResponse> {
  const r = await createBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['createBackup'], version, data: r.data });
  }
  return r;
}

export async function deleteBackupReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string }): Promise<API.CommonResponse> {
  const r = await deletePolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteBackup'], version, data: r.data });
  }
  return r;
}

export async function editBackupReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.UpdateTenantPolicy & { version: string }): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['editBackup'], version, data: r.data });
  }
  return r;
}

export async function pauseBackupReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.UpdateTenantPolicy & { version: string }): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['pauseBackup'], version, data: r.data });
  }
  return r;
}

export async function resumeBackupReportWrap({
  version,
  ...params
}: API.NamespaceAndName &
  API.UpdateTenantPolicy & { version: string }): Promise<API.CommonResponse> {
  const r = await updateBackupPolicyOfTenant(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['resumeBackup'], version, data: r.data });
  }
  return r;
}
