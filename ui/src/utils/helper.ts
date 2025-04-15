import { InitialStateType } from '@/access';
import type { PoolDetailType } from '@/components/customModal/ModifyUnitDetailModal';
import { STATISTICS_INTERVAL } from '@/constants';
import { getAppInfo, getStatistics } from '@/services';
import { REPORT_PARAMS_MAP, reportData } from '@/services/reportRequest';
import { intl } from '@/utils/intl';
type StatisticStatus = 'running' | 'deleting' | 'operating' | 'failed';

type StatisticDataType = { status: StatisticStatus; count: number }[];

type ObjType = { [key: string]: unknown };

export const getInitialObjOfKeys = <T>(targetObj: T, keys: (keyof T)[]) => {
  return keys.reduce((pre, cur) => {
    pre[cur] = targetObj[cur];
    return pre;
  }, {} as T);
};

export const formatStatisticData = (
  type: 'cluster' | 'tenant' | 'k8sCluster',
  data: StatisticDataType,
) => {
  const r: API.StatisticData = {
    total: 0,
    name:
      type === 'cluster'
        ? intl.formatMessage({
            id: 'Dashboard.src.utils.helper.OceanbaseCluster',
            defaultMessage: 'OceanBase集群',
          })
        : type === 'tenant'
        ? intl.formatMessage({
            id: 'Dashboard.src.utils.helper.OceanbaseTenants',
            defaultMessage: 'OceanBase租户',
          })
        : intl.formatMessage({
            id: 'src.utils.50939F01',
            defaultMessage: 'K8s集群',
          }),

    type,
    deleting: 0,
    operating: 0,
    running: 0,
    failed: 0,
  };
  for (const item of data) {
    r.total += item.count;
    r[item.status] = item.count;
  }
  return r;
};

export const formatPatchPoolData = (
  originUnitData: PoolDetailType,
  type: 'edit' | 'create',
) => {
  const newOriginUnitData: PoolDetailType = {
    unitConfig: {},
  };
  newOriginUnitData.unitConfig = {
    ...originUnitData.unitConfig,
    logDiskSize: originUnitData.unitConfig.logDiskSize + 'Gi',
    memorySize: originUnitData.unitConfig.memorySize + 'Gi',
    cpuCount: String(originUnitData.unitConfig.cpuCount),
  };
  if (type === 'create') {
    newOriginUnitData.zoneName = originUnitData.zoneName;
    newOriginUnitData.priority = originUnitData.priority;
  }
  if (type === 'edit') {
    Object.keys(originUnitData).forEach((key) => {
      if (key !== 'unitConfig') {
        newOriginUnitData.zoneName = originUnitData.zoneName;
      }
      if (key === 'pools') {
        newOriginUnitData.type =
          originUnitData?.pools[originUnitData.zoneName].type;
      }
      if (originUnitData[key]?.priority) {
        newOriginUnitData.priority = originUnitData[key].priority;
        newOriginUnitData.type = originUnitData[key].type;
      }
    });
  }
  return newOriginUnitData;
};

export const strTrim = (obj: ObjType): ObjType => {
  Object.keys(obj).forEach((key: keyof ObjType) => {
    if (typeof obj[key] === 'string') {
      obj[key] = obj[key].trim();
    } else if (typeof obj[key] === 'object' && obj[key] !== null) {
      strTrim(obj[key] as ObjType);
    }
  });
  return obj;
};
export const getAppInfoFromStorage = async (): Promise<API.AppInfo> => {
  try {
    let appInfo: API.AppInfo = JSON.parse(sessionStorage.getItem('appInfo'));
    if (appInfo) {
      return Promise.resolve(appInfo);
    }
    appInfo = (await getAppInfo()).data;
    return appInfo;
  } catch {
    return Promise.reject();
  }
};

export const isReportTimeExpired = (lastTimestamp: number): boolean => {
  return Date.now() - lastTimestamp >= STATISTICS_INTERVAL;
};

export const reportPollData = async () => {
  try {
    const appInfo = await getAppInfoFromStorage();
    if (!appInfo.reportStatistics) return;
    const { data } = await getStatistics();
    await reportData({
      ...REPORT_PARAMS_MAP['polling'],
      version: appInfo.version,
      data,
    });
    localStorage.setItem('lastReportTime', Date.now().toString());
  } catch (err) {}
};

export function floorToTwoDecimalPlaces(num: number) {
  return Math.floor(num * 100) / 100;
}

export function initializeAccess(
  accessObj: { [T: string]: boolean },
  initialState: InitialStateType,
) {
  initialState.policies.forEach((policy) => {
    accessObj[`${policy.domain}${policy.action}`] = false;
  });
  for (const role of initialState.accountInfo.roles) {
    const { policies } = role;
    for (const policy of policies) {
      if (policy.domain) {
        Object.keys(accessObj).forEach((key) => {
          if (
            `${policy.domain}${policy.action}` === key ||
            (key.includes(policy.domain) && policy.action === '*') ||
            (policy.domain === '*' && key.includes(policy.action)) ||
            (policy.domain === '*' && policy.action === '*')
          ) {
            accessObj[key] = true;
          }
        });
      }
    }
  }
}
