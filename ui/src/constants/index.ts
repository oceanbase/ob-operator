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

//Unify status constants and colors
const STATUS = ['running', 'deleting', 'operating'];
const COLOR_MAP = new Map([
  ['running', 'geekblue'],
  ['deleting', 'volcano'],
  ['operating', 'gold'],
  ['creating', 'blue'],
]);
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

const REFRESH_CLUSTER_TIME = 10000;
// Monitor automatic refresh interval 15s
const REFRESH_FREQUENCY = 15;
// Number of monitoring points
const POINT_NUMBER = 15;

const SUFFIX_UNIT = 'GB';

const MINIMAL_CONFIG = {
  cpu: 2,
  memory: 10,
  data: 30,
  log: 30,
  redoLog: 30,
};

const RESULT_STATUS = ['running','failed'];

const BACKUP_RESULT_STATUS = ['RUNNING','FAILED','PAUSED']

const RESOURCE_NAME_REG = /^[a-z\-]+$/;
// use for tenant name or zone name
const TZ_NAME_REG =  /^[_a-zA-Z][^-\n]*$/;

export {
  BADGE_IMG_MAP,
  CLUSTER_IMG_MAP,
  COLOR_MAP,
  MINIMAL_CONFIG,
  POINT_NUMBER,
  REFRESH_CLUSTER_TIME,
  REFRESH_FREQUENCY,
  REFRESH_TENANT_TIME,
  SERVER_IMG_MAP,
  STATUS,
  SUFFIX_UNIT,
  ZONE_IMG_MAP,
  RESULT_STATUS,
  BACKUP_RESULT_STATUS,
  RESOURCE_NAME_REG,
  TZ_NAME_REG
};
