import type { UnitDetailType } from '@/components/customModal/ModifyUnitDetailModal';
import { intl } from '@/utils/intl';
import { clone } from 'lodash';
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

export const formatUnitDetailData = (originUnitData: UnitDetailType) => {
  const _originUnitData: API.UnitConfig = clone(originUnitData);
  _originUnitData.logDiskSize = _originUnitData.logDiskSize + 'GB';
  _originUnitData.memorySize = _originUnitData.memorySize + 'GB';
  _originUnitData.cpuCount = String(_originUnitData.cpuCount); 
  return {
    unitConfig: {
      unitConfig: { ..._originUnitData },
    },
  };
};