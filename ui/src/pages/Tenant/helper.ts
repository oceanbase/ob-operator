import { encryptText } from '@/hook/usePublicKey';
import dayjs from 'dayjs';
import { clone,cloneDeep } from 'lodash';
import type { MaxResourceType } from './New/ResourcePools';

const isExist = (val: string | number | undefined): boolean => {
  if (typeof val === 'number') return true;
  return !!val;
};
const formatUnitConfig = (unitConfig: any): API.UnitConfig => {
  let _unitConfig = clone(unitConfig);
  _unitConfig['cpuCount'] = String(_unitConfig['cpuCount']);
  if (isExist(_unitConfig['logDiskSize'])) {
    _unitConfig['logDiskSize'] = _unitConfig['logDiskSize'] + 'Gi';
  }
  if (isExist(_unitConfig['memorySize'])) {
    _unitConfig['memorySize'] = _unitConfig['memorySize'] + 'Gi';
  }
  return _unitConfig;
};

export function formatNewTenantForm(
  originFormData: any,
  clusterName: string,
  publicKey: string,
): API.TenantBody {
  let result: API.TenantBody = {};
  Object.keys(originFormData).forEach((key) => {
    if (key === 'connectWhiteList') {
      result[key] = originFormData[key].join(',');
    } else if (key === 'obcluster') {
      result[key] = clusterName;
    } else if (key === 'pools') {
      result[key] = Object.keys(originFormData[key])
        .filter((zoneKey) => {
          return originFormData[key][zoneKey].checked;
        })
        .map((zoneName) => {
          const { priority } = originFormData[key]?.[zoneName];
          return priority || priority === 0
            ? {
                zone: zoneName,
                priority,
                type: 'Full',
              }
            : {
                zone: zoneName,
                type: 'Full',
              };
        });
    } else if (key === 'source') {
      if (originFormData[key]['tenant'] || originFormData[key]['restore'])
        result[key] = {};
      if (originFormData[key]['tenant']) {
        result[key]['tenant'] = originFormData[key]['tenant'];
      }
      if (originFormData[key]['restore']) {
        let { until } = originFormData[key]['restore'];
        
        result[key]['restore'] = {
          ...originFormData[key]['restore'],
          until:
            (until && until.date && until.time)
              ? {
                  timestamp:
                    dayjs(until.date).format('YYYY-MM-DD') +
                    ' ' +
                    dayjs(until.time).format('HH:mm:ss'),
                }
              : { unlimited: true },
        };
        if(originFormData[key]?.restore?.type === 'OSS'){
          result[key]['restore']['ossAccessId'] = encryptText(
            originFormData[key]['restore'].ossAccessId,
            publicKey,
          );
          result[key]['restore']['ossAccessKey'] = encryptText(
            originFormData[key]['restore'].ossAccessKey,
            publicKey,
          );
        }
        if (originFormData[key]['restore'].bakEncryptionPassword) {
          result[key]['restore']['bakEncryptionPassword'] = encryptText(
            originFormData[key]['restore'].bakEncryptionPassword,
            publicKey,
          );
        } else {
          delete result[key]['restore']['bakEncryptionPassword'];
        }
      }
    } else if (key === 'rootPassword') {
      result[key] = encryptText(originFormData[key], publicKey);
    } else if (key === 'unitConfig') {
      result[key] = formatUnitConfig(originFormData[key]);
    } else {
      result[key] = originFormData[key];
    }
  });
  return result;
}
/**
 * encrypt ossAccessId,ossAccessKey,bakEncryptionPassword
 *
 * format scheduleDates
 */
export function formatBackupForm(originFormData: any, publicKey?: string) {
  let formData = clone(originFormData);
  if (formData.bakEncryptionPassword) {
    formData.bakEncryptionPassword = publicKey
      ? encryptText(originFormData.bakEncryptionPassword, publicKey)
      : originFormData.bakEncryptionPassword;
  }
  if (formData.ossAccessId) {
    formData.ossAccessId = publicKey
      ? encryptText(originFormData.ossAccessId, publicKey)
      : originFormData.ossAccessId;
  }
  if (formData.ossAccessKey)
    formData.ossAccessKey = publicKey
      ? encryptText(originFormData.ossAccessKey, publicKey)
      : originFormData.ossAccessId;
  formData.scheduleTime = dayjs(formData.scheduleTime).format('HH:mm');
  formData.scheduleType = formData.scheduleDates.mode;
  delete formData.scheduleDates.days;
  delete formData.scheduleDates.mode;
  formData.scheduleDates = Object.keys(formData.scheduleDates).map((key) => ({
    day: Number(key),
    backupType: formData.scheduleDates[key],
  }));
  return formData;
}

export function formatBackupPolicyData(backupPolicy: API.BackupPolicy) {
  if (!backupPolicy) return;
  let result: any = {};
  result.days = backupPolicy.scheduleDates.map((item) => item.day);
  result.mode = backupPolicy.scheduleType;
  result.days.forEach((day, index) => {
    result[day] = backupPolicy.scheduleDates[index].backupType;
  });
  return result;
}

function checkDateIsSame(
  preDates: API.ScheduleDatesType,
  curDates: API.ScheduleDatesType,
): boolean {
  if (preDates.length !== curDates.length) return false;
  for (let preDate of preDates) {
    let targetItem = curDates.find((curDate) => curDate.day === preDate.day);
    if (!targetItem) return false;
    if (targetItem.backupType !== preDate.backupType) return false;
  }
  return true;
}

export function checkIsSame(
  preData: API.BackupPolicy,
  curData: API.BackupConfigEditable,
): boolean {
  for (let key of Object.keys(curData)) {
    if (key === 'scheduleDates') {
      if (
        !checkDateIsSame(preData['scheduleDates'], curData['scheduleDates'])
      ) {
        return false;
      }
    } else {
      if (curData[key] !== preData[key]) {
        return false;
      }
    }
  }

  return true;
}

function findMinValue(
  key: 'availableCPU' | 'availableLogDisk' | 'availableMemory',
  resources: API.ServerResource[],
) {
  if (!resources.length) return [];
  return resources.sort((pre, cur) => cur[key] - pre[key])[0][key];
}

export function findMinParameter(
  zones: string[],
  essentialParameter: API.EssentialParametersType,
): MaxResourceType {
  
  const { obServerResources } = essentialParameter;
  const selectResources = obServerResources.filter((resource) =>
    zones.includes(resource.obZone),
  );
  
  return {
    maxCPU: findMinValue('availableCPU', selectResources),
    maxLogDisk: findMinValue('availableLogDisk', selectResources),
    maxMemory: findMinValue('availableMemory', selectResources),
  };
}

/**
 *
 * @Describe Modify the checked status of a zone in a cluster from the cluster list
 */
export const modifyZoneCheckedStatus = (
  clusterList: API.SimpleClusterList,
  zone: string,
  checked: boolean,
  target: {
    id?: number;
    name?: string;
  },
) => {
  const _clusterList = cloneDeep(clusterList);
  for (let cluster of _clusterList) {
    if (cluster.clusterId === target.id || cluster.name === target.name) {
      cluster.topology.forEach((zoneItem) => {
        if (zoneItem.zone === zone) {
          zoneItem.checked = checked;
        }
      });
    }
  }
  return _clusterList;
};

export const checkScheduleDatesHaveFull = (scheduleDates): boolean => {
  for (let key of Object.keys(scheduleDates)) {
    if (!isNaN(key)) {
      if (scheduleDates[key] === 'Full') {
        return true;
      }
    }
  }
  return false;
};

export const getClusterFromTenant = (
  clusterList: API.SimpleClusterList,
  clusterResourceName: string,
): API.SimpleCluster | undefined => {
  return clusterList.find((cluster) => cluster.name === clusterResourceName);
};

const formatReplicToOption = (
  replicaList: API.ReplicaDetailType[] | API.Topology[],
): API.OptionsType => {
  return replicaList.map((replica) => ({
    label: replica.zone,
    value: replica.zone,
  }));
};

const filterAbnormalZone = (zones: API.Topology[]): API.Topology[] => {
  return zones.filter(
    (zone) =>
      !!zone.observers.find((server) => server.statusDetail === 'running'),
  );
};

export const getZonesOptions = (
  cluster: API.SimpleCluster | undefined,
  replicaList: API.ReplicaDetailType[] | undefined,
): API.OptionsType => {
  if (!replicaList) return [];
  if (!cluster) return formatReplicToOption(replicaList);
  const newReplicas = filterAbnormalZone(cluster.topology).filter((zone) => {
    if (replicaList.find((replica) => replica.zone === zone.zone)) {
      return false;
    } else {
      return true;
    }
  });
  return formatReplicToOption(newReplicas);
};

export const getOriginResourceUsages = (
  resourceUsages: API.EssentialParametersType | undefined,
  current: API.ReplicaDetailType | undefined,
) => {
  if (!resourceUsages) return;
  if(!current) return resourceUsages;
  const originResourceUsages = cloneDeep(resourceUsages);
  originResourceUsages.obServerResources.forEach((resource) => {
    if (resource.obZone === current.zone) {
      resource.availableCPU += Number(current.minCPU);
      resource.availableLogDisk += Number(current.logDiskSize);
      resource.availableMemory += Number(current.memorySize);
      originResourceUsages.obZoneResourceMap[current.zone].availableCPU +=
        Number(current.minCPU);
      originResourceUsages.obZoneResourceMap[current.zone].availableLogDisk +=
        Number(current.logDiskSize.split('Gi')[0]);
      originResourceUsages.obZoneResourceMap[current.zone].availableMemory +=
        Number(current.memorySize.split('Gi')[0]);
    }
  });
  return originResourceUsages;
};