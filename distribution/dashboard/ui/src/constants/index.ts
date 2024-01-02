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
const status = ['running', 'deleting', 'operating'];
const colorMap = new Map();
const clusterImgMap = new Map();
const serverImgMap = new Map();
const zoneImgMap = new Map();
const badgeIMgMap = new Map();


colorMap.set('running', 'geekblue');
colorMap.set('deleting', 'volcano');
colorMap.set('operating', 'gold');
clusterImgMap.set('running', clusterRunning);
clusterImgMap.set('deleting', clusterOperating);
clusterImgMap.set('operating', clusterDeleting);
serverImgMap.set('running', serverRunning);
serverImgMap.set('deleting', serverDeleting);
serverImgMap.set('operating', serverOperating);
zoneImgMap.set('running', zoneRunning);
zoneImgMap.set('deleting', zoneDeleting);
zoneImgMap.set('operating', zoneOperating);
badgeIMgMap.set('running', badgeRunning);
badgeIMgMap.set('deleting', badgeDeleting);
badgeIMgMap.set('operating', badgeOperating);

const REFRESH_CLUSTER_TIME = 10000
// 性能监控自动刷新时间 15s
const REFRESH_FREQUENCY = 15;
// 监控点位数量
const POINT_NUMBER = 15;

export {
  badgeIMgMap,
  clusterImgMap,
  colorMap,
  serverImgMap,
  status,
  zoneImgMap,
  REFRESH_CLUSTER_TIME,
  REFRESH_FREQUENCY,
  POINT_NUMBER
};
