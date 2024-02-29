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

//统一状态常量和颜色
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

const REFRESH_CLUSTER_TIME = 10000;
// 性能监控自动刷新时间 15s
const REFRESH_FREQUENCY = 15;
// 监控点位数量
const POINT_NUMBER = 15;

const SUFFIX_UNIT = 'GB';

export {
  POINT_NUMBER,
  REFRESH_CLUSTER_TIME,
  REFRESH_FREQUENCY,
  SUFFIX_UNIT,
  BADGE_IMG_MAP,
  CLUSTER_IMG_MAP,
  COLOR_MAP,
  SERVER_IMG_MAP,
  STATUS,
  ZONE_IMG_MAP,
};
