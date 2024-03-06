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
  
  let r:API.StatisticData = {
    total: 0,
    name: type === 'cluster' ? 'OceanBase集群' : 'OceanBase租户',
    type,
    deleting: 0,
    operating: 0,
    running: 0,
    failed: 0
  };
  for (let item of data) {
    r.total += item.count;
    r[item.status] = item.count;
  }
  return r
};
