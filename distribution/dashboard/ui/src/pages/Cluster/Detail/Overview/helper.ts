// 与UI无关函数

/**
 * 通过url的path获取namespace和name
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
      const [ns, name,clusterName] = path.split('&');
      if (ns.split('=')[0] === 'ns' && name.split('=')[0] === 'nm') {
        res[0] = ns.split('=')[1];
        res[1] = name.split('=')[1];
        res[2] = clusterName.split('=')[1];
      }
      return res;
    }
  }
  return res;
};

// 存在集群｜zone｜server状态不为running 则返回status为operating
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
    if (key !== 'topology' && key !== 'metrics') {
      res['info'][key] = responseData[key];
    } else if (key === 'metrics') {
      res[key] = responseData[key];
    } else {
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
