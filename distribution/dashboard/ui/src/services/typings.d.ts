/* eslint-disable */
// 该文件由 OneAPI 自动生成，请勿手动修改！

declare namespace API {
  interface User {
    username: string;
    password: string;
  }

  type Metrics = {
    cpuPercent: number;
    diskPercent: number;
    memoryPercent: number;
  };

  type NodeSelector = {
    key: string;
    value: string;
  };

  type ClusterInfo = {
    name: string;
    namespace: string;
    status: string;
    statusDetail: string;
    image: string;
    createTime: string;
    metrics: Metrics;
    clusterId: number;
    clusterName: string;
  };

  type Zone = {
    name: string;
    namespace: string;
    zone: string;
    replicas: string;
    status: string;
    rootService: string;
    statusDetail: string;
    nodeSelector: NodeSelector[];
    servers: Server[];
  };

  type Server = {
    name: string;
    namespace: string;
    status: string;
    statusDetail: string;
    address: string;
    metrics: Metrics;
    zone?: string; //所属zone
  };

  type ClusterDetail = {
    info: ClusterInfo;
    zones: Zone[];
    servers: Server[];
    metrics: Metrics;
    status?: 'running' | 'operating';
  };

  type Topology = {
    name: string;
    namespace: string;
    zone: string;
    replicas: string;
    status: string;
    statusDetail: string;
    rootService: string;
    observers: Server[];
  };

  type ClusterItem = {
    topology: Topology[];
  } & ClusterInfo;

  type ClusterList = ClusterItem[];

  type modalType = 'upgrade' | 'addZone' | 'scaleServer';

  type QueryMetricsType = {
    groupLabels: string[];
    labels: { key: string; value: string }[];
    metrics: string[];
    queryRange: { endTimestamp: number; startTimestamp: number; step: number };
  };

  type EventType = 'NORMAL' | 'WARNING';

  type EventObjectType = 'OBCLUSTER' | 'OBTENANT' | 'OBCLUSTER_OVERVIEW';

  interface TenantDetail {
    clusterName: string;
    createTime: string;
    name: string;
    namespace: string;
    status: string;
    tenantName: string;
    tenantRole: string;
    topology: [
      {
        iopsWeight: 0;
        logDiskSize: string;
        maxCPU: string;
        maxIops: 0;
        memorySize: string;
        minCPU: string;
        minIops: 0;
        priority: 0;
        type: string;
        zone: string;
      },
    ];
    unitNumber: 0;
  }

  type NamespaceAndName = {
    ns: string;
    name: string;
  };

  type TenantBody = {};
  type TenantPolicy = {};

  interface CommonResponse {
    data: any;
    message: string;
    successful: boolean;
  }

  interface TenantsListResponse extends CommonResponse {
    data: TenantDetail[];
  }

  interface TenantInfoType extends CommonResponse {
    data: {
      clusterName: string;
    };
  }

  type ReplayLogType = {
    timestamp: number;
    unlimited: boolean;
  };

  type RootPassword = {
    rootPassword: string;
  };

  type UnitNumber = {
    unitNum: number;
  };

  type UnitConfig = {
    iopsWeight: number;
    logDiskSize: string;
    maxCPU: string;
    maxIops: number;
    memorySize: string;
    minCPU: string;
    minIops: number;
  };
}
