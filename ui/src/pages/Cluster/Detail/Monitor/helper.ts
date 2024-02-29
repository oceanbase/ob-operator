import type { FilterDataType, OptionType } from '.';

export const getFilterData = (clusterDetail: any): FilterDataType => {
  let serverList: OptionType[] = [],
    zoneList: OptionType[] = [];
  if (clusterDetail.servers) {
    for (let server of clusterDetail.servers) {
      if (server.address)
        serverList.push({
          label: server.address,
          value: server.address,
          zone: server.zone,
        });
    }
  }

  if (clusterDetail.zones) {
    for (let zone of clusterDetail.zones) {
      if (zone.zone)
        zoneList.push({
          label: zone.zone,
          value: zone.zone,
        });
    }
  }

  return {
    serverList,
    zoneList,
    date: '',
  };
};

/**
 * step: 每个点位间隔时间s 例如 半小时1800s 间隔step 30s 返回60个点位
 * @param pointNumber 点数，默认15个点位
 * @param startTimeStamp 开始时间戳 精确到s
 * @param endTimeStamp 结束时间戳 精确到s
 * @returns
 */
export const caculateStep = (
  startTimeStamp: number,
  endTimeStamp: number,
  pointNumber: number,
): number => {
  return Math.ceil((endTimeStamp - startTimeStamp) / pointNumber);
};
