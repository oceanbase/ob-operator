import type { PoolDetailType } from '@/components/customModal/ModifyUnitDetailModal';
import { getAppInfo, getStatistics } from '@/services';
import { REPORT_PARAMS_MAP, reportData } from '@/services/reportRequest';
import { STATISTICS_INTERVAL } from '@/constants';
import { intl } from '@/utils/intl';
type StatisticStatus = 'running' | 'deleting' | 'operating' | 'failed';

type StatisticDataType = { status: StatisticStatus; count: number }[];

type ObjType = { [key: string]: any };

export const getInitialObjOfKeys = (targetObj: any, keys: string[]) => {
  return keys.reduce((pre, cur) => {
    pre[cur] = targetObj[cur];
    return pre;
  }, {});
};

export const formatStatisticData = (
  type: 'cluster' | 'tenant',
  data: StatisticDataType,
) => {
  let r: API.StatisticData = {
    total: 0,
    name:
      type === 'cluster'
        ? intl.formatMessage({
            id: 'Dashboard.src.utils.helper.OceanbaseCluster',
            defaultMessage: 'OceanBase集群',
          })
        : intl.formatMessage({
            id: 'Dashboard.src.utils.helper.OceanbaseTenants',
            defaultMessage: 'OceanBase租户',
          }),
    type,
    deleting: 0,
    operating: 0,
    running: 0,
    failed: 0,
  };
  for (let item of data) {
    r.total += item.count;
    r[item.status] = item.count;
  }
  return r;
};

export const formatPatchPoolData = (
  originUnitData: PoolDetailType,
  type: 'edit' | 'create',
) => {
  let newOriginUnitData: PoolDetailType = {
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
        newOriginUnitData.zoneName = key;
      }
      if (originUnitData[key]?.priority) {
        newOriginUnitData.priority = originUnitData[key].priority;
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
      strTrim(obj[key]);
    }
  });
  return obj;
};
export const getAppInfoFromStorage = async (): Promise<API.AppInfo> => {
  try {
    let appInfo: API.AppInfo = JSON.parse(sessionStorage.getItem('appInfo'));
    if (!appInfo) {
      appInfo = (await getAppInfo()).data;
    }
    return appInfo;
  } catch (err) {}
};

export const isReportTimeExpired = (lastTimestamp: number): boolean => {
  return Date.now() - lastTimestamp >= STATISTICS_INTERVAL;
};

export const reportPollData = async () => {
  try {
    const appInfo = await getAppInfoFromStorage();
    if (!appInfo.reportStatistics) return;
    let { data } = await getStatistics();
    await reportData({
      ...REPORT_PARAMS_MAP['polling'],
      version: appInfo.version,
      data,
    });
    localStorage.setItem('lastReportTime', Date.now().toString());
  } catch (err) {}
};
