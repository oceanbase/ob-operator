import { intl } from '@/utils/intl';
import type { GraphNodeType } from './helper';
type OperateTypeLabel = { value: string; label: string; disabled?: boolean }[];

const clusterOperate: OperateTypeLabel = [
  {
    value: 'addZone',
    label: intl.formatMessage({
      id: 'dashboard.Detail.Topo.constants.AddZone',
      defaultMessage: '新增zone',
    }),
  },
  {
    value: 'upgradeCluster',
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Topo.constants.Upgrade',
      defaultMessage: '升级',
    }),
  },
  {
    value: 'deleteCluster',
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Topo.constants.Delete',
      defaultMessage: '删除',
    }),
  },
];

const zoneOperate: OperateTypeLabel = [
  {
    value: 'scaleServer',
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Topo.constants.Scale',
      defaultMessage: '扩缩容',
    }),
  },
  {
    value: 'deleteZone',
    label: intl.formatMessage({
      id: 'dashboard.Detail.Topo.constants.DeleteZone',
      defaultMessage: '删除zone',
    }),
    disabled: false,
  },
];

const serverOperate: OperateTypeLabel = [
  // {
  //   value: 'add',
  //   label: intl.formatMessage({
  //     id: 'dashboard.Detail.Topo.constants.AddServer',
  //     defaultMessage: '添加server',
  //   }),
  // },
  // {
  //   value: 'delete',
  //   label: intl.formatMessage({
  //     id: 'dashboard.Detail.Topo.constants.DeleteServer',
  //     defaultMessage: '删除server',
  //   }),
  // },
  // {
  //   value: 'upgrade',
  //   label: intl.formatMessage({
  //     id: 'dashboard.Detail.Topo.constants.UpgradeServer',
  //     defaultMessage: '升级server',
  //   }),
  // },
];

const clusterOperateOfTenant: OperateTypeLabel = [
  {
    value: 'changeUnitCount',
    label: intl.formatMessage({
      id: 'Dashboard.components.TopoComponent.constants.ModifyTheNumberOfUnits',
      defaultMessage: '修改 Unit 数量',
    }),
  },
];

const getZoneOperateOfTenant = (
  haveResourcePool: boolean,
  tenantReplicas: API.ReplicaDetailType[],
): OperateTypeLabel => {
  return haveResourcePool
    ? [
        {
          value: 'editResourcePools',
          label: intl.formatMessage({
            id: 'Dashboard.components.TopoComponent.constants.EditResourcePool',
            defaultMessage: '编辑资源池',
          }),
          disabled: false,
        },
        {
          value: 'deleteResourcePool',
          label: intl.formatMessage({
            id: 'Dashboard.components.TopoComponent.constants.DeleteAResourcePool',
            defaultMessage: '删除资源池',
          }),
          disabled: tenantReplicas.length <= 2,
        },
      ]
    : [
        {
          value: 'createResourcePools',
          label: intl.formatMessage({
            id: 'Dashboard.components.TopoComponent.constants.AddAResourcePool',
            defaultMessage: '新增资源池',
          }),
          disabled: false,
        },
      ];
};

const getZoneOperateOfCluster = (
  topoData: GraphNodeType | undefined,
): OperateTypeLabel => {
  if (!topoData) return [];
  const isDisabled = topoData?.children?.length <= 2;
  zoneOperate.forEach((operate) => {
    if (operate.value === 'deleteZone') operate.disabled = isDisabled;
  });
  return zoneOperate;
};

export {
  clusterOperate,
  clusterOperateOfTenant,
  getZoneOperateOfCluster,
  getZoneOperateOfTenant,
  serverOperate,
  zoneOperate,
};

export type { OperateTypeLabel };
