import type { Topo } from '@/type/topo';
import { intl } from '@/utils/intl';
import { clone } from 'lodash';

const clusterOperate: Topo.OperateTypeLabel = [
{
  value: 'addZone',
  label: intl.formatMessage({
    id: 'dashboard.Detail.Topo.constants.AddZone',
    defaultMessage: '新增zone'
  })
},
{
  value: 'upgradeCluster',
  label: intl.formatMessage({
    id: 'OBDashboard.Detail.Topo.constants.Upgrade',
    defaultMessage: '升级'
  })
},
{
  value: 'deleteCluster',
  label: intl.formatMessage({
    id: 'OBDashboard.Detail.Topo.constants.Delete',
    defaultMessage: '删除'
  })
}];


const zoneOperate: Topo.OperateTypeLabel = [
{
  value: 'scaleServer',
  label: intl.formatMessage({
    id: 'OBDashboard.Detail.Topo.constants.Scale',
    defaultMessage: '扩缩容'
  })
},
{
  value: 'deleteZone',
  label: intl.formatMessage({
    id: 'dashboard.Detail.Topo.constants.DeleteZone',
    defaultMessage: '删除zone'
  }),
  disabled: false
}];


const serverOperate: Topo.OperateTypeLabel = [
{
  value: 'restartServer',
  label: intl.formatMessage({ id: "src.components.TopoComponent.2494A473", defaultMessage: "重启server" }),
  disabled: false
},
{
  value: 'deleteServer',
  label: intl.formatMessage({
    id: 'dashboard.Detail.Topo.constants.DeleteServer',
    defaultMessage: '删除 server'
  }),
  disabled: false
}
// {
//   value: 'add',
//   label: intl.formatMessage({
//     id: 'dashboard.Detail.Topo.constants.AddServer',
//     defaultMessage: '添加server',
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

const clusterOperateOfTenant: Topo.OperateTypeLabel = [
{
  value: 'changeUnitCount',
  label: intl.formatMessage({
    id: 'Dashboard.components.TopoComponent.constants.ModifyTheNumberOfUnits',
    defaultMessage: '修改 Unit 数量'
  })
}];


const getZoneOperateOfTenant = (
haveResourcePool: boolean,
tenantReplicas: API.ReplicaDetailType[],
tenantStatus?: string,
clusterStatus?: string)
: Topo.OperateTypeLabel => {
  return haveResourcePool ?
  [
  {
    value: 'editResourcePools',
    label: intl.formatMessage({
      id: 'Dashboard.components.TopoComponent.constants.EditResourcePool',
      defaultMessage: '编辑资源池'
    }),
    disabled: tenantStatus !== 'running' || clusterStatus !== 'running'
  },
  {
    value: 'deleteResourcePool',
    label: intl.formatMessage({
      id: 'Dashboard.components.TopoComponent.constants.DeleteAResourcePool',
      defaultMessage: '删除资源池'
    }),
    disabled:
    tenantReplicas.length <= 2 ||
    tenantStatus !== 'running' ||
    clusterStatus !== 'running'
  }] :

  [
  {
    value: 'createResourcePools',
    label: intl.formatMessage({
      id: 'Dashboard.components.TopoComponent.constants.AddAResourcePool',
      defaultMessage: '新增资源池'
    }),
    disabled: tenantStatus !== 'running' || clusterStatus !== 'running'
  }];

};

const getZoneOperateOfCluster = (
topoData: Topo.GraphNodeType | undefined,
status: string)
: Topo.OperateTypeLabel => {
  if (!topoData) return [];
  const isDisabled = topoData?.children?.length <= 2 || status !== 'running';
  zoneOperate.forEach((operate) => {
    if (operate.value === 'deleteZone') operate.disabled = isDisabled;
    if (operate.value === 'scaleServer')
    operate.disabled = status !== 'running';
  });
  return zoneOperate;
};
const getServerOperateOfCluster = (
topoData: Topo.GraphNodeType | undefined,
disabled: boolean,
serverZone: string)
: Topo.OperateTypeLabel => {
  if (!topoData) return [];
  // 任何 zone 里面只剩一个 server 就不能删了
  const supportStaticIPisDisabled = topoData?.supportStaticIP;
  const zoneCurrent = topoData?.children?.find(
    (zone) => zone.label === serverZone
  );

  const serverCurrentisDisabled = zoneCurrent?.children?.length === 1;

  serverOperate.forEach((operate) => {
    if (disabled) {
      operate.disabled = disabled;
    }
    if (operate.value === 'deleteServer')
    operate.disabled = serverCurrentisDisabled;
    if (operate.value === 'restartServer')
    operate.disabled = !supportStaticIPisDisabled;
  });
  return serverOperate;
};

const getClusterOperates = (
clusterOperateList: Topo.OperateTypeLabel,
disabled: boolean)
: Topo.OperateTypeLabel => {
  const res = clone(clusterOperateList);
  res.forEach((item) => {
    item.disabled = disabled;
  });
  return res;
};

export {
  clusterOperate,
  clusterOperateOfTenant,
  getClusterOperates,
  getServerOperateOfCluster,
  getZoneOperateOfCluster,
  getZoneOperateOfTenant,
  serverOperate,
  zoneOperate };