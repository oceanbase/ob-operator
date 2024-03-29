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
  version: string;
  data: any;
};

type ReportMapType = {
  [T: string]: { resourceType: ResourceType; eventType: EventType };
};

const REPORT_URL = 'http://openwebapi.test.alipay.net/api/web/oceanbase/report';
// const queryUrl = 'http://openwebapi.test.alipay.net/api/web/oceanbase/query';
const REPORT_COMPONENT = 'oceanbase-dashboard';
export const REPORT_PARAMS_MAP: ReportMapType = {
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
  version,
  data,
}: ReportDataParms) {
  await request(REPORT_URL, {
    method: 'POST',
    data: {
      content: JSON.stringify({
        resourceType,
        eventType,
        version,
        body: data,
      }),
      component: REPORT_COMPONENT,
    },
  });
}