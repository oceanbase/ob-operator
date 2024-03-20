import type { PoolDetailType } from '@/components/customModal/ModifyUnitDetailModal';
import { intl } from '@/utils/intl';
type StatisticStatus = 'running' | 'deleting' | 'operating' | 'failed';

type StatisticDataType = { status: StatisticStatus; count: number }[];

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

export const formatPatchPoolData = (originUnitData: PoolDetailType,type:'edit'|'create') => {
  let newOriginUnitData: PoolDetailType = {
    unitConfig: {},
  };
  newOriginUnitData.unitConfig = {
    ...originUnitData.unitConfig,
    logDiskSize: originUnitData.unitConfig.logDiskSize + 'Gi',
    memorySize: originUnitData.unitConfig.memorySize + 'Gi',
    cpuCount: String(originUnitData.unitConfig.cpuCount),
  };
  if(type === 'create'){
    newOriginUnitData.zoneName = originUnitData.zoneName;
    newOriginUnitData.priority = originUnitData.priority;
  }
  if(type === 'edit'){
    Object.keys(originUnitData).forEach((key) => {
      if (originUnitData[key]?.priority) {
        newOriginUnitData.zoneName = key;
        newOriginUnitData.priority = originUnitData[key].priority;
      }
    });
  }
  return newOriginUnitData;
};
