import {
  badgeIMgMap,
  clusterImgMap,
  serverImgMap,
  zoneImgMap,
} from '@/constants';
import { intl } from '@/utils/intl';
import { Graph, IG6GraphEvent, INode, IShape } from '@antv/g6';
import _ from 'lodash';

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

type GraphNodeType = {
  id: string;
  label: string;
  status: string;
  type: string;
  img: string;
  badgeImg: string;
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
 * 监听mouseenter和mouseleave事件时，evt.shape为null(g6本身如此)
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

function getChildren(zoneList: any) {
  let children = [];
  for (let zone of zoneList) {
    let temp: GraphNodeType = {
      id: '',
      label: '',
      status: '',
      type: 'zone',
      img: '',
      badgeImg: '',
    };
    temp.id = zone.name + zone.namespace; //k8s里是通过name+ns查询资源 所以ns+name是唯一的
    temp.label = zone.zone;
    temp.status = zone.status;
    temp.img = zoneImgMap.get(zone.status);
    temp.badgeImg = badgeIMgMap.get(zone.status);
    temp.children = zone.observers.map((server: TopoServer) => {
      return {
        id: server.name + server.namespace,
        label: server.address,
        status: server.status,
        type: 'server',
        img: serverImgMap.get(server.status),
        badgeImg: badgeIMgMap.get(server.status),
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
    img: clusterImgMap.get(responseData.status),
    badgeImg: badgeIMgMap.get(responseData.status),
  };
  topoData.children = getChildren(responseData.topology);

  let basicInfo: BasicInfoType = {
    name: responseData.name,
    namespace: responseData.namespace,
    status: responseData.status,
    image: responseData.image,
  };

  return {
    topoData,
    basicInfo,
  };
};

/**
 * 判断新旧topoData属性值是否完全相同
 */
export const checkIsSame = (oldTopoData: any, newTopoData: any): boolean => {
  if (!_.matches(oldTopoData)(newTopoData)) return false;
  //判断children
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
