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
  const _originUnitData: UnitDetailType = clone(originUnitData);
  _originUnitData.unitConfig.unitConfig.logDiskSize = _originUnitData.unitConfig.unitConfig.logDiskSize + 'Gi';
  _originUnitData.unitConfig.unitConfig.memorySize = _originUnitData.unitConfig.unitConfig.memorySize + 'Gi';
  _originUnitData.unitConfig.unitConfig.cpuCount = String(
    _originUnitData.unitConfig.unitConfig.cpuCount,
  );
  return {
    unitConfig: {
      unitConfig: _originUnitData.unitConfig.unitConfig,
      pools: Object.keys(_originUnitData.unitConfig.pools)
        .map((zone) => ({
          zone,
          priority: _originUnitData.unitConfig.pools?.[zone]?.priority,
          type: 'Full',
        }))
        .filter((item) => item.priority || item.priority === 0),
    },
  };
};