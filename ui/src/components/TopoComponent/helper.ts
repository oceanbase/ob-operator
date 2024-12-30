import {
  BADGE_IMG_MAP,
  CLUSTER_IMG_MAP,
  MAX_IOPS,
  SERVER_IMG_MAP,
  TOPO_INFO_CONFIG,
  ZONE_IMG_MAP,
} from '@/constants';
import type { Topo } from '@/type/topo';
import { intl } from '@/utils/intl';
import { Graph, INode } from '@antv/g6';
import _ from 'lodash';

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
        (shape?.get(propName) as Topo.ShapeEventListner) ||
        evt.target.cfg[propName];
      if (func) {
        func(evt, item, shape, graph);
      }
    });
  });
}

function getZoneTypeText(
  zone: Pick<API.ReplicaDetailType, 'zone'>,
  tenantTopoData: API.ReplicaDetailType[],
) {
  return tenantTopoData.find((item) => item.zone === zone.zone)?.type;
}

function getTooltipInfo(
  zone: Pick<API.ReplicaDetailType, 'zone'>,
  tenantTopoData: API.ReplicaDetailType[],
) {
  const targetZone = tenantTopoData.find((item) => item.zone === zone.zone);
  if (targetZone) {
    return {
      maxCPU: targetZone.maxCPU,
      minCPU: targetZone.minCPU,
      memorySize: targetZone.memorySize,
      minIops:
        targetZone.minIops >= MAX_IOPS
          ? intl.formatMessage({
              id: 'src.components.TopoComponent.601BE33A',
              defaultMessage: '无限制',
            })
          : targetZone.minIops,
      maxIops:
        targetZone.maxIops >= MAX_IOPS
          ? intl.formatMessage({
              id: 'src.components.TopoComponent.0E4AAB80',
              defaultMessage: '无限制',
            })
          : targetZone.maxIops,
    };
  }
  return;
}

function getChildren(zoneList: any, tenantReplicas?: API.ReplicaDetailType[]) {
  const children = [];
  for (const zone of zoneList) {
    const temp: Topo.GraphNodeType = {
      id: '',
      label: '',
      status: '',
      type: 'zone',
      img: '',
      badgeImg: '',
      disable: false,
    };
    const typeText = getZoneTypeText(zone, tenantReplicas || []);
    const tooltipInfo = getTooltipInfo(zone, tenantReplicas || []);
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
    temp.children = zone.observers.map((server: Topo.TopoServer) => {
      return {
        id: server.name + server.namespace,
        name: server.name,
        label: server.address,
        status: server.status,
        type: 'server',
        img: SERVER_IMG_MAP.get(server.status),
        badgeImg: BADGE_IMG_MAP.get(server.status),
        disable: temp.disable,
        zone: zone.zone,
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
  topoData: Topo.GraphNodeType;
  basicInfo: Topo.BasicInfoType;
} => {
  if (!responseData) return responseData;
  const topoData: Topo.GraphNodeType = {
    id: responseData.namespace + responseData.name,
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Topo.helper.Cluster',
      defaultMessage: '集群',
    }),
    status: responseData.status,
    supportStaticIP: responseData.supportStaticIP,
    type: 'cluster',
    children: [],
    img: CLUSTER_IMG_MAP.get(responseData.status),
    badgeImg: BADGE_IMG_MAP.get(responseData.status),
    disable: false,
  };
  topoData.children = getChildren(responseData.topology, tenantReplicas);

  const basicInfo: API.ClusterInfo = {};
  for (const key of Object.keys(responseData)) {
    if (TOPO_INFO_CONFIG.includes(key)) {
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
export const checkTopoDataIsSame = (
  oldTopoData: any,
  newTopoData: any,
): boolean => {
  if (!_.matches(oldTopoData)(newTopoData)) return false;
  if (newTopoData.children.length > oldTopoData.children.length) return false;
  oldTopoData.children.forEach((oldZone: any, idx: number) => {
    const newZone = newTopoData.children[idx];
    if (newZone.children.length > oldZone.children.length) return false;
  });
  return true;
};

export const getServerNumber = (
  topoData: Topo.GraphNodeType,
  zoneName: string,
): number => {
  const zones = topoData.children || [];
  for (const zone of zones) {
    if (zone.label === zoneName) {
      return zone.children?.length || 0;
    }
  }
  return 0;
};

export const haveDisabledOperate = (operateList: Topo.OperateTypeLabel) => {
  return operateList.find((operate) => operate.disabled);
};
