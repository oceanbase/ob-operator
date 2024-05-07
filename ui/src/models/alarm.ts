import { getSimpleClusterList } from '@/services';
import { getAllTenants } from '@/services/tenant';
import { useEffect, useState } from 'react';

export default () => {
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>();
  const [tenantList, setTenantList] = useState<API.TenantDetail[]>();

  useEffect(() => {
    getSimpleClusterList().then(({ successful, data }) => {
      if (successful) setClusterList(data);
    });
    getAllTenants().then(({ successful, data }) => {
      if (successful) setTenantList(data);
    });
  }, []);

  return {
    clusterList,
    setClusterList,
    tenantList,
    setTenantList,
  };
};
