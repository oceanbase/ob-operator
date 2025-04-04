import badgeDeleting from '@/assets/badge/error.svg';
import badgeRunning from '@/assets/badge/processing.svg';
import badgeOperating from '@/assets/badge/warning.svg';
import clusterDeleting from '@/assets/cluster/deleting.svg';
import clusterOperating from '@/assets/cluster/operating.svg';
import clusterRunning from '@/assets/cluster/running.svg';
import serverDeleting from '@/assets/server/deleting.svg';
import serverRunning from '@/assets/server/running.svg';
import serverOperating from '@/assets/server/warning.svg';
import zoneDeleting from '@/assets/zone/deleting.svg';
import zoneOperating from '@/assets/zone/operating.svg';
import zoneRunning from '@/assets/zone/running.svg';
import { intl } from '@/utils/intl';
import type { DefaultOptionType, SelectProps } from 'antd/es/select';

//Unify status constants and colors
const STATUS = ['running', 'deleting', 'operating'];

const CLUSTER_IMG_MAP = new Map([
  ['running', clusterRunning],
  ['deleting', clusterOperating],
  ['operating', clusterDeleting],
]);
const SERVER_IMG_MAP = new Map([
  ['running', serverRunning],
  ['deleting', serverDeleting],
  ['operating', serverOperating],
]);
const ZONE_IMG_MAP = new Map([
  ['running', zoneRunning],
  ['deleting', zoneDeleting],
  ['operating', zoneOperating],
]);
const BADGE_IMG_MAP = new Map([
  ['running', badgeRunning],
  ['deleting', badgeDeleting],
  ['operating', badgeOperating],
]);

const REFRESH_TENANT_TIME = 5000;

const REFRESH_OBPROXY_TIME = 3000;

const REFRESH_CLUSTER_TIME = 10000;
// Monitor automatic refresh interval 15s
const REFRESH_FREQUENCY = 15;
// Number of monitoring points
const POINT_NUMBER = 15;

// two minutes
const CHECK_STORAGE_INTERVAL = 1000 * 120;

// three hours
const STATISTICS_INTERVAL = 1000 * 60 * 60 * 3;

const SUFFIX_UNIT = 'GB';

const MINIMAL_CONFIG = {
  cpu: 2,
  memory: 10,
  data: 30,
  log: 30,
  redoLog: 30,
};

const RESULT_STATUS = ['running', 'failed'];

const BACKUP_RESULT_STATUS = ['RUNNING', 'FAILED', 'PAUSED'];

const CLUSTER_INFO_CONFIG = [
  'name',
  'clusterName',
  'namespace',
  'status',
  'statusDetail',
  'image',
  'resource',
  'storage',
  'backupVolume',
  'monitor',
  'rootPasswordSecret',
  'mode',
  'parameters',
  'deletionProtection',
  'topology',
  'pvcIndependent',
  'version',
];

const TOPO_INFO_CONFIG = [
  'name',
  'clusterName',
  'namespace',
  'status',
  'statusDetail',
  'image',
  'mode',
  'rootPasswordSecret',
  'deletionProtection',
  'version',
];

// use for tenant name or zone name
const TZ_NAME_REG = /^[_a-zA-Z][^-\n]*$/;

const MIN_RESOURCE_CONFIG = {
  minCPU: 1,
  minLogDisk: 5,
  minMemory: 2,
  minIops: 1024,
  maxIops: 1024,
};

const getMinResource = (defaultValue?: any) => {
  return {
    ...MIN_RESOURCE_CONFIG,
    ...defaultValue,
  };
};

const MODE_MAP = new Map([
  [
    'NORMAL',
    {
      text: intl.formatMessage({
        id: 'Dashboard.src.constants.RegularMode',
        defaultMessage: '常规模式',
      }),
    },
  ],

  [
    'STANDALONE',
    {
      text: intl.formatMessage({
        id: 'Dashboard.src.constants.MonomerMode',
        defaultMessage: '单体模式',
      }),
      limit: intl.formatMessage({
        id: 'Dashboard.src.constants.RequiredKernelVersion',
        defaultMessage: '要求内核版本 >= 4.2.0.0',
      }),
    },
  ],

  [
    'SERVICE',
    {
      text: intl.formatMessage({
        id: 'Dashboard.src.constants.ServiceMode',
        defaultMessage: 'Service模式',
      }),
      limit: intl.formatMessage({
        id: 'src.constants.22EDA0DC',
        defaultMessage: '要求内核版本 >= 4.2.1.4 (4.2.2.x 除外)',
      }),
    },
  ],
]);

const LEVER_OPTIONS_ALARM: SelectProps['options'] = [
  {
    label: intl.formatMessage({
      id: 'src.constants.BA2DCE55',
      defaultMessage: '严重',
    }),
    value: 'critical',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.CA3D7CB8',
      defaultMessage: '警告',
    }),
    value: 'warning',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.2DF58E75',
      defaultMessage: '注意',
    }),
    value: 'caution',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.53A94A0E',
      defaultMessage: '提醒',
    }),
    value: 'info',
  },
];

const SEVERITY_MAP = {
  critical: {
    color: 'red',
    label: intl.formatMessage({
      id: 'src.constants.101F76A0',
      defaultMessage: '严重',
    }),
    weight: 3,
  },
  warning: {
    color: 'gold',
    label: intl.formatMessage({
      id: 'src.constants.9076A376',
      defaultMessage: '警告',
    }),
    weight: 2,
  },
  caution: {
    color: 'blue',
    label: intl.formatMessage({
      id: 'src.constants.0C9F6F8C',
      defaultMessage: '注意',
    }),
    weight: 1,
  },
  info: {
    color: 'green',
    label: intl.formatMessage({
      id: 'src.constants.661BB167',
      defaultMessage: '提醒',
    }),
    weight: 0,
  },
};

const SHILED_STATUS_MAP = {
  expired: {
    text: intl.formatMessage({
      id: 'src.constants.041D1159',
      defaultMessage: '过期',
    }),
    color: 'gold',
    weight: 1,
  },
  pending: {
    text: intl.formatMessage({
      id: 'src.constants.5A1A196E',
      defaultMessage: '未生效',
    }),
    color: 'default',
    weight: 0,
  },
  active: {
    text: intl.formatMessage({
      id: 'src.constants.7BCA3313',
      defaultMessage: '活跃',
    }),
    color: 'green',
    weight: 2,
  },
};

const OBJECT_OPTIONS_ALARM: DefaultOptionType[] = [
  {
    label: intl.formatMessage({
      id: 'src.constants.5D9B444A',
      defaultMessage: '集群',
    }),
    value: 'obcluster',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.FB06E464',
      defaultMessage: '租户',
    }),
    value: 'obtenant',
  },
  {
    label: 'OBServer',
    value: 'observer',
  },
];

const CHANNEL_TYPE_OPTIONS = [
  {
    value: 'dingtalk',
    key: 'dingtalk',
  },
  {
    value: 'wechat',
    key: 'wechat',
  },
];

const ALERT_STATE_MAP = {
  active: {
    text: intl.formatMessage({
      id: 'src.constants.D7C4C6F1',
      defaultMessage: '活跃',
    }),
    color: 'green',
    weight: 0,
  },
  unprocessed: {
    text: intl.formatMessage({
      id: 'src.constants.D953F862',
      defaultMessage: '未处理',
    }),
    color: 'default',
    weight: 1,
  },
  suppressed: {
    text: intl.formatMessage({
      id: 'src.constants.F77CCEC3',
      defaultMessage: '抑制',
    }),
    color: 'red',
    weight: 2,
  },
};

const SERVICE_TYPE = [
  {
    label: 'ClusterIP',
    value: 'ClusterIP',
  },
  {
    label: 'NodePort',
    value: 'NodePort',
  },
  {
    label: 'LoadBalancer',
    value: 'LoadBalancer',
  },
];

const DEFAULT_QUERY_RANGE: Monitor.QueryRangeType = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};

const LABELNAME_REG = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

const VALIDATE_DEBOUNCE = 1000;

// Greater than or equal to 1073741824 means no limit
const MAX_IOPS = 1073741824;

export {
  ALERT_STATE_MAP,
  BACKUP_RESULT_STATUS,
  BADGE_IMG_MAP,
  CHANNEL_TYPE_OPTIONS,
  CHECK_STORAGE_INTERVAL,
  CLUSTER_IMG_MAP,
  CLUSTER_INFO_CONFIG,
  DEFAULT_QUERY_RANGE,
  LABELNAME_REG,
  LEVER_OPTIONS_ALARM,
  MAX_IOPS,
  MINIMAL_CONFIG,
  MIN_RESOURCE_CONFIG,
  MODE_MAP,
  OBJECT_OPTIONS_ALARM,
  POINT_NUMBER,
  REFRESH_CLUSTER_TIME,
  REFRESH_FREQUENCY,
  REFRESH_OBPROXY_TIME,
  REFRESH_TENANT_TIME,
  RESULT_STATUS,
  SERVER_IMG_MAP,
  SERVICE_TYPE,
  SEVERITY_MAP,
  SHILED_STATUS_MAP,
  STATISTICS_INTERVAL,
  STATUS,
  SUFFIX_UNIT,
  TOPO_INFO_CONFIG,
  TZ_NAME_REG,
  VALIDATE_DEBOUNCE,
  ZONE_IMG_MAP,
  getMinResource,
};

export const STATUS_LIST = [
  {
    label: intl.formatMessage({
      id: 'src.constants.NormalOperation',
      defaultMessage: '正常运行',
    }),
    value: 'running',
    badgeStatus: 'geekblue',
  },

  {
    label: intl.formatMessage({
      id: 'src.constants.Creating',
      defaultMessage: '创建中',
    }),

    value: 'creating',
    badgeStatus: 'processing',
  },

  {
    label: intl.formatMessage({
      id: 'src.constants.Deleting',
      defaultMessage: '删除中',
    }),

    value: 'deleting',
    badgeStatus: 'volcano',
  },

  {
    label: intl.formatMessage({
      id: 'src.constants.Unavailable',
      defaultMessage: '不可用',
    }),

    value: 'failed',
    badgeStatus: 'error',
  },

  {
    label: intl.formatMessage({
      id: 'src.constants.Recovering',
      defaultMessage: '运维中',
    }),
    value: 'operating',
    badgeStatus: 'gold',
  },
];

export const OBPROXY_STATUS_LIST = [
  {
    label: intl.formatMessage({
      id: 'src.constants.NormalOperation',
      defaultMessage: '正常运行',
    }),
    value: 'Running',
    badgeStatus: 'geekblue',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.Pending',
      defaultMessage: '等待中',
    }),
    value: 'Pending',
    badgeStatus: 'gold',
  },
];

export const NODESTABLE_STATUS_LIST = [
  {
    label: intl.formatMessage({
      id: 'src.constants.NormalOperation',
      defaultMessage: '正常运行',
    }),
    value: 'ready',
    badgeStatus: 'success',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.Warning',
      defaultMessage: '警告',
    }),
    value: 'not ready',
    badgeStatus: 'warning',
  },
];
export const EVENTSTABLE_STATUS_LIST = [
  {
    label: intl.formatMessage({
      id: 'src.constants.NormalOperation',
      defaultMessage: '正常运行',
    }),
    value: 'Normal',
    badgeStatus: 'default',
  },
  {
    label: intl.formatMessage({
      id: 'src.constants.Warning',
      defaultMessage: '警告',
    }),
    value: 'Warning',
    badgeStatus: 'warning',
  },
];

export const LOADTYPE_LIST = [
  {
    value: 'EXPRESS_OLTP',
    label: 'Express OLTP',
  },
  {
    value: 'EXPRESS_OLTP_PERF',
    label: 'Express OLTP PERF',
  },
  {
    value: 'COMPLEX_OLTP',
    label: 'Complex OLTP',
  },
  {
    value: 'HTAP',
    label: 'HTAP',
  },
  {
    value: 'OLAP',
    label: 'OLAP',
  },
  {
    value: 'KV',
    label: 'KV',
  },
];
