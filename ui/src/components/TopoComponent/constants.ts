import { intl } from '@/utils/intl';
type OperateTypeLabel = { value: string; label: string }[];

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
      id: 'OBDashboard.Detail.Topo.constants.Expansion',
      defaultMessage: '扩缩容',
    }),
  },
  {
    value: 'deleteZone',
    label: intl.formatMessage({
      id: 'dashboard.Detail.Topo.constants.DeleteZone',
      defaultMessage: '删除zone',
    }),
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

const zoneOperateOfTenant: OperateTypeLabel = [
  {
    value: 'modifyUnitSpecification',
    label: intl.formatMessage({
      id: 'Dashboard.components.TopoComponent.constants.AdjustUnitSpecifications',
      defaultMessage: '调整 Unit 规格',
    }),
  },
];

export {
  clusterOperate,
  clusterOperateOfTenant,
  serverOperate,
  zoneOperate,
  zoneOperateOfTenant,
};

export type { OperateTypeLabel };
