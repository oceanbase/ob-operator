import { useState } from 'react';

export default () => {
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>();
  const [tenantList, setTenantList] = useState<API.TenantDetail[]>();

  return {
    clusterList,
    setClusterList,
    tenantList,
    setTenantList,
  };
};
