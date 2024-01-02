import { intl } from '@/utils/intl';
type OperateType = { value: string; label: string }[];

const clusterOperate: OperateType = [
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

const zoneOperate: OperateType = [
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

const serverOperate: OperateType = [
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

export { clusterOperate, serverOperate, zoneOperate };
export type { OperateType };
