import { getAppInfoFromStorage } from '@/utils/helper';
import { request } from '@umijs/max';

type ResourceType =
  | 'Statistics'
  | 'OBCluster'
  | 'OBTenant'
  | 'OBTenantBackupPolicy';

type EventType =
  | 'Normal'
  | 'Create'
  | 'Delete'
  | 'Upgrade'
  | 'AddZone'
  | 'DeleteZone'
  | 'ScaleZone'
  | 'ModifyUnitNum'
  | 'AddPool'
  | 'EditPool'
  | 'DeletePool'
  | 'ModifyRootPwd'
  | 'ActivateStandby'
  | 'Switchover'
  | 'ReplayLog'
  | 'Edit'
  | 'Pause'
  | 'Resume';

type ReportDataParms = {
  resourceType: ResourceType;
  eventType: EventType;
  version?: string;
  data: unknown;
};

type ReportMapType = {
  [T: string]: { resourceType: ResourceType; eventType: EventType };
};

export const REPORT_PARAMS_MAP: ReportMapType = {
  // polling
  polling: {
    resourceType: 'Statistics',
    eventType: 'Normal',
  },
  // cluster
  createCluster: {
    resourceType: 'OBCluster',
    eventType: 'Create',
  },
  deleteCluster: {
    resourceType: 'OBCluster',
    eventType: 'Delete',
  },
  upgradeCluster: {
    resourceType: 'OBCluster',
    eventType: 'Upgrade',
  },
  addZone: {
    resourceType: 'OBCluster',
    eventType: 'AddZone',
  },
  deleteZone: {
    resourceType: 'OBCluster',
    eventType: 'DeleteZone',
  },
  scaleZone: {
    resourceType: 'OBCluster',
    eventType: 'ScaleZone',
  },
  // tenant
  createTenant: {
    resourceType: 'OBTenant',
    eventType: 'Create',
  },
  deleteTenant: {
    resourceType: 'OBTenant',
    eventType: 'Delete',
  },
  modifyUnitNum: {
    resourceType: 'OBTenant',
    eventType: 'ModifyUnitNum',
  },
  addPool: {
    resourceType: 'OBTenant',
    eventType: 'AddPool',
  },
  editPool: {
    resourceType: 'OBTenant',
    eventType: 'EditPool',
  },
  deletePool: {
    resourceType: 'OBTenant',
    eventType: 'DeletePool',
  },
  modifyRootPwd: {
    resourceType: 'OBTenant',
    eventType: 'ModifyRootPwd',
  },
  activateStandby: {
    resourceType: 'OBTenant',
    eventType: 'ActivateStandby',
  },
  switchover: {
    resourceType: 'OBTenant',
    eventType: 'Switchover',
  },
  replayLog: {
    resourceType: 'OBTenant',
    eventType: 'ReplayLog',
  },
  // backup
  createBackup: {
    resourceType: 'OBTenantBackupPolicy',
    eventType: 'Create',
  },
  deleteBackup: {
    resourceType: 'OBTenantBackupPolicy',
    eventType: 'Delete',
  },
  editBackup: {
    resourceType: 'OBTenantBackupPolicy',
    eventType: 'Edit',
  },
  pauseBackup: {
    resourceType: 'OBTenantBackupPolicy',
    eventType: 'Pause',
  },
  resumeBackup: {
    resourceType: 'OBTenantBackupPolicy',
    eventType: 'Resume',
  },
};

export async function reportData({
  resourceType,
  eventType,
  data,
}: ReportDataParms): Promise<unknown> {
  const appInfo = await getAppInfoFromStorage();

  return await request(`${appInfo.reportHost}/api/web/oceanbase/report`, {
    method: 'POST',
    data: {
      content: JSON.stringify({
        resourceType,
        eventType,
        version: appInfo.version,
        body: data,
      }),
      component: appInfo.appName,
    },
  });
}
