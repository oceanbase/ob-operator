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

  interface ClusterListResponse extends CommonResponse {
    data: ClusterItem[];
  }

  type SimpleClusterList = {
    name: string;
    clusterId: number;
    namespace: string;
    topology: Topology[];
  }[];

  type ClusterList = ClusterItem[];

  type ModalType =
    | 'upgrade'
    | 'addZone'
    | 'scaleServer'
    | 'modifyUnit'
    | 'changePassword'
    | 'logReplay'
    | 'activateTenant'
    | 'switchTenant'
    | 'upgradeTenant'

  type QueryMetricsType = {
    groupLabels: string[];
    labels: { key: string; value: string }[];
    metrics: string[];
    queryRange: { endTimestamp: number; startTimestamp: number; step: number };
  };

  type EventType = 'NORMAL' | 'WARNING';

  type EventObjectType = 'OBCLUSTER' | 'OBTENANT' | 'OBCLUSTER_OVERVIEW';

  type TenantRole = 'Primary' | 'Standby'

  interface TenantDetail {
    charset: string;
    clusterName: string;
    createTime: string;
    locality: string;
    name: string;
    namespace: string;
    status: string;
    tenantName: string;
    tenantRole: TenantRole;
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

  type TenantBody = {
    connectWhiteList?: string;
    name: string;
    obcluster: string;
    pools?: {
      priority: number;
      zone: string;
    }[];
    rootPassword: string;
    source?: {
      restore?: {
        archiveSource: string;
        bakDataSource: string;
        bakEncryptionPassword?: string;
        ossAccessId: string;
        ossAccessKey: string;
        type: string;
        until?: {
          timestamp?: string;
          unlimited?: boolean;
        };
      };
      tenant?: string;
    };
    tenantName: string;
    tenantRole?: TenantRole;
    unitConfig: {
      iopsWeight?: number;
      logDiskSize?: string;
      cupNumber: number;
      maxIops?: number;
      memorySize: string;
      minIops?: number;
    };
    unitNum: number;
  };
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
      charset: string;
      clusterName: string;
      createTime: string;
      locality: string;
      name: string;
      namespace: string;
      primaryTenant: string;
      restoreSource: {
        archiveSource: string;
        bakDataSource: string;
        bakEncryptionSecret: string;
        ossAccessSecret: string;
        type: string;
        until: string;
      };
      rootCredential: string;
      standbyROCredentail: string;
      status: string;
      tenantName: string;
      tenantRole: TenantRole;
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
    };
  }

  type ReplayLogType = {
    timestamp: string;
    unlimited: boolean;
  };

  type UserCredentials = {
    User: string;
    Password: string;
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

  type PatchTenantConfiguration = {
    unitConfig?: {
      pools?: {
        priority: number;
        zone: string;
      }[];
      unitConfig?: UnitConfig;
    };
    unitNum?: number;
  };

  type InfoType = {
    charset: string;
    clusterName: string;
    tenantName: string;
    tenantRole: TenantRole;
    unitNumber: number;
    status: string;
    name: string;
    namespace: string;
    locality: string;
    style?: any;
  };
  interface TenantsListResponse extends CommonResponse {
    data: TenantDetail[];
  }

  interface TenantBasicInfoResponse  extends CommonResponse {
    data:TenantBasicInfo
  }

  type TenantBasicInfo = {
    info:InfoType,
    source?: {
      primaryTenant?: string;
      archiveSource?: string;
      bakDataSource?: string;
      until?: string;
    };
  }
}
