// Functions without UI
import { CLUSTER_INFO_CONFIG, RESULT_STATUS } from "@/constants";

// if there is cluster｜zone｜server whose status isn't running,the return status is operating.
const formatClusterData = (responseData: any): API.ClusterDetail => {
  const res: any = {
    status:responseData.status,
    info: {},
    metrics: {},
    zones: [],
    servers: [],
  };
  // let status: 'running' | 'operating' = 'running';
  
  for (let key of Object.keys(responseData)) {
    // if (key === 'status' && !RESULT_STATUS.includes(responseData[key])) {
    //   status = 'operating';
    // }
    if(CLUSTER_INFO_CONFIG.includes(key)){
      res['info'][key] = responseData[key];
    }
    if (key === 'metrics') {
      res[key] = responseData[key];
    }
    if(key === 'topology'){
      const servers: API.Server[] = [];
      const zones = responseData[key].map((zone: any) => {
        let temp: any = {};
        for (let _key in zone) {
          if (_key !== 'observers') {
            temp[_key] = zone[_key];
          } else {
            zone[_key].forEach((server:API.Server)=>{
              server.zone = zone.zone
            })
            temp.servers = zone[_key]
            servers.push(...zone[_key]);
          }
        }
        return temp;
      });
      // for (let zone of zones) {
      //   if (!RESULT_STATUS.includes(zone.status)) status = 'operating';
      // }
      // for (let server of servers) {
      //   if (!RESULT_STATUS.includes(server.status)) status = 'operating';
      // }
      res['zones'] = zones;
      res['servers'] = servers;
    }
  }
  return res;
};

export { formatClusterData };
