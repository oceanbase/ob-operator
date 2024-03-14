// Functions without UI

/**
 * Get the namespace, name and cluster name or tenant name through the path of the url
 *
 * @returns {Array} [namespace,name]
 * @example /cluster/ns=oceanbase&nm=test/overview => [oceanbase,test]
 */
const getNSName = () => {
  let pathArr = location.hash.split('/'),
    res: string[] = [];
  if (!pathArr.length) return res;
  for (let path of pathArr) {
    
    if (path.split('&').length === 3) {
      const [ns, name,clusterOrTenantName] = path.split('&');
      if (ns.split('=')[0] === 'ns' && name.split('=')[0] === 'nm') {
        res[0] = ns.split('=')[1];
        res[1] = name.split('=')[1];
        res[2] = clusterOrTenantName.split('=')[1];
      }
      return res;
    }else if(path.split('&').length === 2){
      const [ns, name] = path.split('&');
      if (ns.split('=')[0] === 'ns' && name.split('=')[0] === 'nm') {
        res[0] = ns.split('=')[1];
        res[1] = name.split('=')[1];
      }
      return res;
    }
  }
  return res;
};

const clusterInfoConfig = [
  'name',
  'namespace',
  'status',
  'image',
  'resource',
  'storage',
  'backupVolume',
  'monitor',
  'rootPasswordSecret',
  'mode',
  'parameters'
]

// if there is cluster｜zone｜server whose status isn't running,the return status is operating.
const formatClusterData = (responseData: any): API.ClusterDetail => {
  const res: any = {
    info: {},
    metrics: {},
    zones: [],
    servers: [],
  };
  let status: 'running' | 'operating' = 'running';
  for (let key of Object.keys(responseData)) {
    if (key === 'status' && responseData[key] !== 'running') {
      status = 'operating';
    }
    if(clusterInfoConfig.includes(key)){
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
      for (let zone of zones) {
        if (zone.status !== 'running') status = 'operating';
      }
      for (let server of servers) {
        if (server.status !== 'running') status = 'operating';
      }
      res['zones'] = zones;
      res['servers'] = servers;
    }
  }
  res.status = status;
  return res;
};

export { formatClusterData, getNSName };
