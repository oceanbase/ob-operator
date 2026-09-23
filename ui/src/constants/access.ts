import { intl } from '@/utils/intl';

export const ACCESS_ROLES_LIST = [
  {
    label: intl.formatMessage({
      id: 'dashboard.ObjectStorage.Access',
      defaultMessage: '对象存储凭据',
    }),
    value: 'objectstorage',
    descriptions: intl.formatMessage({
      id: 'dashboard.ObjectStorage.AccessDescription',
      defaultMessage:
        '读取凭据名称；创建新凭据并执行只读存储桶检查。不支持读取密钥值、更新或删除凭据。',
    }),
  },
  {
    label: 'LogService',
    value: 'oblogservice',
    descriptions: intl.formatMessage({
      id: 'dashboard.LogService.AccessDescription',
      defaultMessage:
        'LogService：读权限用于查看服务，写权限用于创建服务、调整副本数和删除服务。',
    }),
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.D057DB4A',
      defaultMessage: '集群与租户',
    }),
    value: 'obcluster',
    descriptions: intl.formatMessage({
      id: 'src.components.customModal.783C8ED3',
      defaultMessage:
        '集群与租户: 读权限涉及OB集群、租户的获取，写权限涉及集群和租户的创建、更新和删除',
    }),
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.9E29D52E',
      defaultMessage: '系统信息',
    }),
    value: 'system',
    descriptions: intl.formatMessage({
      id: 'src.components.customModal.EA420EE1',
      defaultMessage:
        '系统信息: 读权限涉及监控指标和监控数据的获取，k8s集群节点、事件、命名空间、存储类的获取，写权限涉及命名空间的创建和系统配置的修改',
    }),
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.4EBEFDA8',
      defaultMessage: '监控告警',
    }),
    value: 'alarm',
    descriptions: intl.formatMessage({
      id: 'src.components.customModal.CC80EAD4',
      defaultMessage:
        '监控告警: 读权限涉及告警、Silencer、Rule、Receiver和Route的读取，写权限则涉及它们的着创建、更新和删除',
    }),
  },
  {
    label: 'OBProxy',
    value: 'obproxy',
    descriptions: intl.formatMessage({
      id: 'src.components.customModal.C97E041E',
      defaultMessage:
        'OBProxy: 读权限涉及OBProxy的获取，写权限涉及OBProxy的创建、更新和删除',
    }),
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.939130EF',
      defaultMessage: '权限控制',
    }),
    value: 'ac',
    descriptions: intl.formatMessage({
      id: 'src.components.customModal.2169D3CC',
      defaultMessage:
        '权限控制: 读权限涉及账号和角色的获取，写权限涉及它们的创建、更新和删除',
    }),
  },
  {
    label: intl.formatMessage({
      id: 'dashboard.K8sCluster.Access',
      defaultMessage: 'K8s 集群',
    }),
    value: 'k8s-cluster',
    descriptions: intl.formatMessage({
      id: 'dashboard.K8sCluster.AccessDescription',
      defaultMessage:
        'K8s 集群：读权限用于查看集群、节点和事件，写权限用于创建、修改、删除集群以及更新节点。',
    }),
  },
];
