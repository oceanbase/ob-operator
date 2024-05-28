import type {
  OceanbaseOBInstance,
  OceanbaseOBInstanceType,
  SilenceSilencerParam,
} from '@/api/generated';
import { Alert } from '@/type/alert';
import { clone, flatten } from 'lodash';

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
): Alert.SelectList => {
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

/**
 * Format form data
 *
 * @example
 * 1.handle allServers allTenants allClusters
 * {
 *  observers:['allServers']
 *  type:observers
 * } =>
 * {
 *   observers:['1.1.1.1','2.2.2.2']
 *   type:observers
 * }
 *
 * 2.format instances
 *
 * @example
 * {
 *    type:'obtenant',
 *    obtenant:['tenanta','tenantb']
 * } =>
 * [
 *    {
 *      type:'obtenant',
 *      obtenant:'tenanta'
 *    },
 *    {
 *      type:'obtenant',
 *      obtenant:'tenantb'
 *    }
 * ]
 */
export const formatShieldSubmitData = (
  formData: Alert.ShieldDrawerForm,
  selectList: Alert.ServersList[] & Alert.TenantsList[] & string[],
): SilenceSilencerParam => {
  const cloneFormData = clone(formData);
  const selectInstance = cloneFormData.instances[cloneFormData.instances.type];

  if (
    selectInstance?.includes('allServers') ||
    selectInstance?.includes('allTenants')
  ) {
    const temp = selectList.find(
      (item) => item?.clusterName === cloneFormData.instances.obcluster[0],
    ) as Alert.ServersList & Alert.TenantsList;
    cloneFormData.instances[cloneFormData.instances.type] =
      temp?.tenants || temp?.servers || [];
  }
  if (selectInstance?.includes('allClusters')) {
    cloneFormData.instances['obcluster'] = selectList;
  }

  return {
    ...cloneFormData,
    matchers: cloneFormData.matchers || [],
    instances: flatten(
      Object.keys(cloneFormData.instances)
        .filter((key) => key !== 'type')
        .map((key) => {
          return (
            cloneFormData.instances[key as Alert.InstancesKey]?.map(
              (value: string) =>
                ({
                  type: key,
                  [key]: value,
                } as OceanbaseOBInstance),
            ) || []
          );
        }),
    ),
    endsAt: Math.floor(cloneFormData.endsAt.valueOf() / 1000),
    startsAt: Math.floor(Date.now() / 1000),
    createdBy: localStorage.getItem('user') || '',
  };
};

export const getInstancesFromRes = (
  resInstances: OceanbaseOBInstance[],
): Alert.InstancesType => {
  const getInstanceValues = (type: Alert.InstancesKey) => {
    return flatten(
      resInstances
        .filter((instance) => instance.type === type)
        .map((item) => item[type]!),
    );
  };
  const res: Alert.InstancesType = {
    obcluster: getInstanceValues('obcluster'),
    type: 'obcluster',
  };
  const types = resInstances.map((item) => item.type);
  if (types.includes('observer')) {
    res.type = 'observer';
    res.observer = getInstanceValues('observer');
  } else if (types.includes('obtenant')) {
    res.type = 'obtenant';
    res.obtenant = getInstanceValues('obtenant');
  }
  return res;
};
