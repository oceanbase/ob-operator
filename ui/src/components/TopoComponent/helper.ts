import {
BADGE_IMG_MAP,
CLUSTER_IMG_MAP,
SERVER_IMG_MAP,
TOPO_INFO_CONFIG,
ZONE_IMG_MAP
} from '@/constants';
import { intl } from '@/utils/intl';
import { Graph,IG6GraphEvent,INode,IShape } from '@antv/g6';
import _ from 'lodash';
import type { OperateTypeLabel } from './constants';

export type ShapeEventListner = (
  event: IG6GraphEvent,
  node: INode | null,
  shape: IShape,
  graph: Graph,
) => void;

// interface TopoServer extends Server {
//   img:string;
//   badgeImg:string;
// }
type TopoServer = API.Server & {
  img: string;
  badgeImg: string;
};

type TooltipInfo = {
  cpuCount: number;
  memorySize: string;
  maxIops: number;
  minIops: number;
};

export type GraphNodeType = {
  id: string;
  label: string;
  status: string;
  type: string;
  img: string;
  badgeImg: string;
  disable: boolean;
  typeText?: string;
  tooltipInfo?: TooltipInfo;
  children?: GraphNodeType[];
};

type BasicInfoType = {
  name: string;
  namespace: string;
  status: string;
  image: string;
};

export interface EventAttrs {
  onClick?: ShapeEventListner;
  onDBClick?: ShapeEventListner;
  onMouseEnter?: ShapeEventListner;
  onMouseMove?: ShapeEventListner;
  onMouseOut?: ShapeEventListner;
  onMouseOver?: ShapeEventListner;
  onMouseLeave?: ShapeEventListner;
  onMouseDown?: ShapeEventListner;
  onMouseUp?: ShapeEventListner;
  onDragStart?: ShapeEventListner;
  onDrag?: ShapeEventListner;
  onDragEnd?: ShapeEventListner;
  onDragEnter?: ShapeEventListner;
  onDragLeave?: ShapeEventListner;
  onDragOver?: ShapeEventListner;
  onDrop?: ShapeEventListner;
  onContextMenu?: ShapeEventListner;
}

const propsToEventMap = {
  click: 'onClick',
  dbclick: 'onDBClick',
  mouseenter: 'onMouseEnter',
  mousemove: 'onMouseMove',
  mouseout: 'onMouseOut',
  mouseover: 'onMouseOver',
  mouseleave: 'onMouseLeave',
  mousedown: 'onMouseDown',
  mouseup: 'onMouseUp',
  dragstart: 'onDragStart',
  drag: 'onDrag',
  dragend: 'onDragEnd',
  dragenter: 'onDragEnter',
  dragleave: 'onDragLeave',
  dragover: 'onDragOver',
  drop: 'onDrop',
  contextmenu: 'onContextMenu',
};

/**
 * When listening to mouseenter and mouseleave events, evt.shape is null (g6 itself)
 */
export function appenAutoShapeListener(graph: Graph) {
  Object.entries(propsToEventMap).forEach(([eventName, propName]) => {
    graph.on(`node:${eventName}`, (evt) => {
      const shape = evt.shape;
      const item = evt.item as INode;
      const graph = evt.currentTarget as Graph;
      const func =
        (shape?.get(propName) as ShapeEventListner) || evt.target.cfg[propName];
      if (func) {
        func(evt, item, shape, graph);
      }
    });
  });
}

function getZoneTypeText(zone: any, tenantTopoData: API.ReplicaDetailType[]) {
  return tenantTopoData.find((item) => item.zone === zone.zone)?.type;
}

function getTooltipInfo(zone: any, tenantTopoData: API.ReplicaDetailType[]) {
  let targetZone = tenantTopoData.find((item) => item.zone === zone.zone);
  if (targetZone) {
    return {
      maxCPU: targetZone.maxCPU,
      minCPU: targetZone.minCPU,
      memorySize: targetZone.memorySize,
      minIops: targetZone.minIops,
      maxIops: targetZone.maxIops,
    };
  }
  return;
}

function getChildren(zoneList: any, tenantReplicas?: API.ReplicaDetailType[]) {
  let children = [];
  for (let zone of zoneList) {
    let temp: GraphNodeType = {
      id: '',
      label: '',
      status: '',
      type: 'zone',
      img: '',
      badgeImg: '',
      disable: false,
    };
    let typeText = getZoneTypeText(zone, tenantReplicas || []);
    let tooltipInfo = getTooltipInfo(zone, tenantReplicas || []);
    temp.id = zone.name + zone.namespace; //In k8s, resources are queried through name+ns, so ns+name is unique.
    temp.label = zone.zone;
    temp.status = zone.status;
    temp.img = ZONE_IMG_MAP.get(zone.status);
    temp.badgeImg = BADGE_IMG_MAP.get(zone.status);
    if (typeText) {
      temp.typeText = typeText;
    }
    if (tooltipInfo) {
      temp.tooltipInfo = tooltipInfo;
    }
    if (
      tenantReplicas &&
      !tenantReplicas.find((item) => item.zone === zone.zone)
    ) {
      temp.disable = true;
    }
    temp.children = zone.observers.map((server: TopoServer) => {
      return {
        id: server.name + server.namespace,
        label: server.address,
        status: server.status,
        type: 'server',
        img: SERVER_IMG_MAP.get(server.status),
        badgeImg: BADGE_IMG_MAP.get(server.status),
        disable: temp.disable,
      };
    });
    children.push(temp);
  }
  return children;
}
/**
 * format topodata
 */
export const formatTopoData = (
  responseData: any,
  tenantReplicas?: API.ReplicaDetailType[],
): {
  topoData: GraphNodeType;
  basicInfo: BasicInfoType;
} => {
  if (!responseData) return responseData;
  let topoData: GraphNodeType = {
    id: responseData.namespace + responseData.name,
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Topo.helper.Cluster',
      defaultMessage: '集群',
    }),
    status: responseData.status,
    type: 'cluster',
    children: [],
    img: CLUSTER_IMG_MAP.get(responseData.status),
    badgeImg: BADGE_IMG_MAP.get(responseData.status),
    disable: false,
  };
  topoData.children = getChildren(responseData.topology, tenantReplicas);

  // let basicInfo: BasicInfoType = {
  //   name: responseData.name,
  //   namespace: responseData.namespace,
  //   status: responseData.status,
  //   image: responseData.image,
  // };
  let basicInfo:API.ClusterInfo = {};
  for(let key of Object.keys(responseData)){
    if(TOPO_INFO_CONFIG.includes(key)){
      basicInfo[key] = responseData[key];
    }
  }

  return {
    topoData,
    basicInfo,
  };
};

/**
 * Determine whether the old and new topoData attribute values are exactly the same
 */
export const checkTopoDataIsSame = (oldTopoData: any, newTopoData: any): boolean => {
  if (!_.matches(oldTopoData)(newTopoData)) return false;
  if (newTopoData.children.length > oldTopoData.children.length) return false;
  oldTopoData.children.forEach((oldZone: any, idx: number) => {
    let newZone = newTopoData.children[idx];
    if (newZone.children.length > oldZone.children.length) return false;
  });
  return true;
};

export const getServerNumber = (
  topoData: GraphNodeType,
  zoneName: string,
): number => {
  const zones = topoData.children || [];
  for (let zone of zones) {
    if (zone.label === zoneName) {
      return zone.children?.length || 0;
    }
  }
  return 0;
};

export const haveDisabledOperate = (operateList: OperateTypeLabel) => {
  return operateList.find((operate) => operate.disabled);
};