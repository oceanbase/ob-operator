export const getFilterData = (detail: any): Monitor.FilterDataType => {
  const serverList: Monitor.OptionType[] = [],
    zoneList: Monitor.OptionType[] = [],
    res = {};
  if (detail.servers) {
    for (const server of detail.servers) {
      if (server.address)
        serverList.push({
          label: server.address,
          value: server.address,
          zone: server.zone,
        });
    }
    if (serverList.length) {
      res.serverList = serverList;
    }
  }
  if (detail.zones || detail.replicas) {
    for (const zone of detail.zones || detail.replicas) {
      if (zone.zone)
        zoneList.push({
          label: zone.zone,
          value: zone.zone,
        });
    }
    if (zoneList.length) {
      res.zoneList = zoneList;
    }
  }

  return {
    ...res,
    date: '',
  };
};

/**
 * step: the interval between each point unit:s  for example: half an hour 1800s, interval 30s, return sixty points
 * @param pointNumber points，default 15
 * @param startTimeStamp unit:s
 * @param endTimeStamp
 * @returns
 */
export const caculateStep = (
  startTimeStamp: number,
  endTimeStamp: number,
  pointNumber: number,
): number => {
  return Math.ceil((endTimeStamp - startTimeStamp) / pointNumber);
};
