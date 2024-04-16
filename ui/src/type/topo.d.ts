import type { Graph, IG6GraphEvent, INode, IShape } from '@antv/g6';

declare namespace Topo {
  type OperateTypeLabel = {
    value: string;
    label: string;
    disabled?: boolean;
  }[];

  type ShapeEventListner = (
    event: IG6GraphEvent,
    node: INode | null,
    shape: IShape,
    graph: Graph,
  ) => void;

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

  type GraphNodeType = {
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

  interface EventAttrs {
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
}
