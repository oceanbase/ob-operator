// Functions without UI
import { CLUSTER_INFO_CONFIG } from '@/constants';

// If there is cluster｜zone｜server whose status isn't running,the return status is operating.
const formatClusterData = (responseData: any): API.ClusterDetail => {
  const res: any = {
    status: responseData.status,
    supportStaticIP: responseData.supportStaticIP,
    info: {},
    metrics: {},
    zones: [],
    servers: [],
  };
  for (const key of Object.keys(responseData)) {
    if (CLUSTER_INFO_CONFIG.includes(key)) {
      res['info'][key] = responseData[key];
    }
    if (key === 'metrics') {
      res[key] = responseData[key];
    }
    if (key === 'topology') {
      const servers: API.Server[] = [];
      const zones = responseData[key].map((zone: any) => {
        const temp: any = {};
        for (const _key in zone) {
          if (_key !== 'observers') {
            temp[_key] = zone[_key];
          } else {
            zone[_key].forEach((server: API.Server) => {
              server.zone = zone.zone;
            });
            temp.servers = zone[_key];
            servers.push(...zone[_key]);
          }
        }
        return temp;
      });
      res['zones'] = zones;
      res['servers'] = servers;
    }
  }
  return res;
};

export { formatClusterData };
