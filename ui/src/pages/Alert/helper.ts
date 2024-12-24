import type {
  AlarmMatcher,
  AlertAlert,
  CommonKVPair,
  OceanbaseOBInstance,
  OceanbaseOBInstanceType,
  SilenceSilencerParam,
  SilenceSilencerResponse,
} from '@/api/generated';
import { ALERT_STATE_MAP, SHILED_STATUS_MAP } from '@/constants';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { clone, difference, flatten, uniq } from 'lodash';

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
  selectedTenants?: string[],
  selectedServers?: string[],
): Alert.SelectList => {
  if (type === 'obcluster') {
    if (selectedTenants?.length) {
      // ['sys&clusterName']
      if (selectedTenants.length === 1 && selectedTenants[0].includes('sys')) {
        return [selectedTenants[0].split('&')[1]];
      }
      return clusterList
        .filter((cluster) => {
          const temp = tenantList?.filter((tenant) =>
            selectedTenants.includes(tenant.tenantName),
          );
          return temp?.some(
            (tenant) => tenant.clusterResourceName === cluster.name,
          );
        })
        .map((cluster) => cluster.clusterName);
    }
    if (selectedServers?.length) {
      return clusterList
        .filter((cluster) => {
          for (const zone of cluster.topology) {
            for (const observer of zone.observers) {
              if (selectedServers.includes(observer.address)) return true;
            }
          }
          return false;
        })
        .map((cluster) => cluster.clusterName);
    }
    return clusterList.map((cluster) => cluster.clusterName);
  }
  if (type === 'obtenant') {
    return clusterList.map((cluster) => ({
      clusterName: cluster.clusterName,
      tenants: [
        ...(tenantList
          ?.filter(
            (tenant) =>
              tenant.namespace === cluster.namespace &&
              tenant.clusterResourceName === cluster.name,
          )
          .map((tenant) => tenant.tenantName) || []),
        `sys&${cluster.clusterName}`,
      ],
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

const removeUselessAttr = (instances: Alert.InstancesType) => {
  Object.keys(instances).forEach((key: keyof Alert.InstancesType) => {
    if (!instances[key]) {
      delete instances[key];
    }
  });
};

const handleSystemTenant = (instances: string[] | undefined) => {
  if (!instances) return;
  return instances?.map((tenant) => {
    if (tenant.includes('sys')) return 'sys';
    else return tenant;
  });
};

/**
 * Format form data
 *
 * @example
 *
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
 * {
 *  obcluster:['clustera'],
 *  obtenant:['tenanta','tenantb']
 *  type:'obtenant'
 * } =>
 * [
 *  { type:'obtenant', obcluster:'clustera',obtenant: 'tenanta'},
 *  { type:'obtenant', obcluster:'clustera',obtenant: 'tenantb'}
 * ]
 */
export const formatShieldSubmitData = (
  formData: Alert.ShieldDrawerForm,
  selectList: Alert.ServersList[] & Alert.TenantsList[] & string[],
): SilenceSilencerParam => {
  const cloneFormData = clone(formData);
  removeUselessAttr(cloneFormData.instances);

  let selectInstance = cloneFormData.instances[cloneFormData.instances.type];
  if (
    selectInstance?.includes('allServers') ||
    selectInstance?.includes('allTenants')
  ) {
    const temp = selectList.find(
      (item) => item?.clusterName === cloneFormData.instances.obcluster[0],
    ) as Alert.ServersList & Alert.TenantsList;

    cloneFormData.instances[cloneFormData.instances.type] =
      temp?.tenants || temp?.servers || [];
    selectInstance = temp?.tenants || temp?.servers || [];
  }
  if (selectInstance?.includes('allClusters')) {
    cloneFormData.instances['obcluster'] = selectList;
    selectInstance = selectList;
  }
  if (cloneFormData.instances.type === 'obtenant') {
    selectInstance = handleSystemTenant(selectInstance);
  }
  const tempInstances =
    selectInstance?.map((item) => ({
      type: cloneFormData.instances.type,
      [cloneFormData.instances.type]: item,
    })) || [];

  tempInstances.forEach((item) => {
    const diffKey = difference(
      Object.keys(cloneFormData.instances),
      Object.keys(item),
    );
    for (const key of diffKey) {
      item[key] = cloneFormData.instances[key as Alert.InstancesKey]![0];
    }
  });

  return {
    ...cloneFormData,
    matchers: cloneFormData.matchers || [],
    instances: tempInstances,
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
    res.obcluster = [resInstances[0].obcluster!];
  } else if (types.includes('obtenant')) {
    res.type = 'obtenant';
    res.obtenant = getInstanceValues('obtenant');
    res.obcluster = [resInstances[0].obcluster!];
    for (let i = 0; i < res.obtenant.length; i++) {
      if (res.obtenant[i] === 'sys')
        res.obtenant[i] = `sys&${res.obcluster[0]}`;
    }
  }
  return res;
};

/**
 * @description
 * Sort alarms by status and time
 */
const sortAlarm = (
  alarms: AlertAlert[] | SilenceSilencerResponse[],
  map: { [T: string]: { text: string; color: string; weight: number } },
) => {
  if (!alarms || !alarms.length) return [];
  const types = uniq(alarms.map((item) => item.status.state));
  types.sort((pre, cur) => map[cur].weight - map[pre].weight);
  let res: AlertAlert[] | SilenceSilencerResponse[] = [];
  for (const type of types) {
    const events = alarms.filter((event) => event.status.state === type);
    events.sort((preEvent, curEvent) => curEvent.startsAt - preEvent.startsAt);
    res = [...res, ...events];
  }
  return res;
};

export const sortEvents = (alertEvents: AlertAlert[]) => {
  return sortAlarm(alertEvents, ALERT_STATE_MAP);
};

export const sortAlarmShielding = (
  listSilencers: SilenceSilencerResponse[],
) => {
  return sortAlarm(listSilencers, SHILED_STATUS_MAP);
};

/**
 * @description Each item in LabelInput can be empty, but cannot be incomplete.
 */
export const validateLabelValues = (
  labelValues: AlarmMatcher[] | CommonKVPair[],
) => {
  for (const labelValue of labelValues) {
    const tempArr = Object.keys(labelValue).map((key) =>
      Boolean(labelValue[key as keyof (AlarmMatcher | CommonKVPair)]),
    );
    const stautsArr = tempArr.filter((item) => item === true);
    if (stautsArr.length === 1) return false;
    if (
      stautsArr.length === 2 &&
      tempArr.length > 2 &&
      (labelValue as AlarmMatcher).isRegex
    ) {
      return false;
    }
  }
  return true;
};

/**
 *
 * @param duration The unit is seconds
 */
export const formatDuration = (duration: number) => {
  if (Math.floor(duration / 60) > 180) {
    return intl.formatMessage(
      {
        id: 'src.pages.Alert.05CE0380',
        defaultMessage: '${Math.floor(duration / 3600)}小时',
      },
      { CallExpression0: Math.floor(duration / 3600) },
    );
  }
  if (duration > 600) {
    return intl.formatMessage(
      {
        id: 'src.pages.Alert.232CC9E7',
        defaultMessage: '${Math.floor(duration / 60)}分钟',
      },
      { CallExpression0: Math.floor(duration / 60) },
    );
  }
  return intl.formatMessage(
    { id: 'src.pages.Alert.5ECB5E4A', defaultMessage: '${duration}秒' },
    { duration: duration },
  );
};

export const isListSame = (pre: string[], cur: string[]) => {
  if (pre.length !== cur.length) return false;
  for (const tenant of pre) {
    if (!cur.includes(tenant)) return false;
  }
  return true;
};
