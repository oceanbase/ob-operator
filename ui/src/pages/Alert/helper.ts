import type { OceanbaseOBInstanceType } from '@/api/generated';
import { flatten } from 'lodash';


type SelectList = string[] | TenantsList[] | ServersList[];

export type TenantsList = {
  clusterName: string;
  tenants?: string[];
};

export type ServersList = {
  clusterName: string;
  servers: string[];
};

/**
 * 
 * @description Generate data that conforms to the cluster structure
 * @example
 *
 * getSelectList(clusterList,observer)
 *  => [{
 *     clusterName:'test',
 *      servers:[1.1.1.1]
 *     }]
 */
export const getSelectList = (
  clusterList: API.SimpleClusterList,
  type: OceanbaseOBInstanceType,
  tenantList?: API.TenantDetail[],
): SelectList => {
  if (type === 'obcluster') {
    return clusterList.map((cluster) => cluster.clusterName);
  }
  if (type === 'obtenant') {
    return clusterList.map((cluster) => ({
      clusterName: cluster.clusterName,
      tenants: tenantList
        ?.filter(
          (tenant) =>
            tenant.namespace === cluster.namespace &&
            tenant.clusterResourceName === cluster.name,
        )
        .map((tenant) => tenant.tenantName),
    }));
  }
  if (type === 'observer') {
    return clusterList.map((cluster) => ({
      clusterName: cluster.clusterName,
      servers: flatten(
        cluster.topology.map((zone) =>
          zone.observers.map((server) => server.address),
        ),
      ),
    }));
  }
  return [];
};
