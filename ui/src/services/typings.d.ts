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

  type Storage = {
    size:number;
    storageClass:string;
  }

  type ClusterInfo = {
    name: string;
    namespace: string;
    status: string;
    image: string;
    rootPasswordSecret: string;
    mode: ClusterMode;
    resource?:{
      cpu: number;
      memory: number;
    };
    storage?: {
      dataStorage: Storage;
      redoLogStorage: Storage;
      sysLogStorage: Storage;
    };
    backupVolume?: {
      address: string;
      path: string;
    };
    monitor?: {
      image: string;
      resource: {
        cpu: number;
        memory: number;
      };
    };
    parameters?: {
      key: string;
      value: string;
    }[];
    // statusDetail: string;
    // createTime: string;
    // metrics: Metrics;
    // clusterId: number;
    // clusterName: string;
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
    checked?: boolean;
  };

  type ClusterItem = {
    topology: Topology[];
  } & ClusterInfo;

  type TooltipData = {
    label:string | Element;
    value:string;
    toolTipData:any[]
  }

  type OptionsType = {
    label:string;
    value:string;
  }[]

  interface ClusterListResponse extends CommonResponse {
    data: ClusterItem[];
  }

  interface StorageClassesResponse extends CommonResponse {
    data: TooltipData[];
  }

  type SimpleCluster = {
    name: string;
    clusterName: string;
    clusterId: number;
    namespace: string;
    topology: Topology[];
    status: string;
  }

  type SimpleClusterList = SimpleCluster[];

  interface SimpleClusterListResponse extends CommonResponse {
    data: SimpleClusterList;
  }

  type ClusterList = ClusterItem[];

  type ModalType =
    | 'upgradeCluster'
    | 'addZone'
    | 'scaleServer'
    | 'changePassword'
    | 'logReplay'
    | 'activateTenant'
    | 'switchTenant'
    | 'upgradeTenant'
    | 'changeUnitCount'
    | 'modifyUnitSpecification'
    | 'deleteCluster'
    | 'deleteZone'

  type LableKeys =
    | 'ob_cluster_name'
    | 'ob_cluster_id'
    | 'tenant_name'
    | 'tenant_id'
    | 'svr_ip'
    | 'obzone';

  type MetricsLabels = { key: LableKeys; value: string }[];
  type QueryMetricsType = {
    groupLabels: string[];
    labels: MetricsLabels;
    metrics: string[];
    queryRange: { endTimestamp: number; startTimestamp: number; step: number };
    type: MonitorUseTarget;
    useFor: MonitorUseFor;
    filterData?: API.ClusterItem[] | API.TenantDetail[];
  };

  type ClusterMode = 'NORMAL' | 'STANDALONE' | 'SERVICE'

  type MonitorUseFor = 'cluster' | 'tenant';

  type EventType = 'NORMAL' | 'WARNING';

  type MonitorUseTarget = 'OVERVIEW' | 'DETAIL';

  type EventObjectType = 'OBCLUSTER' | 'OBTENANT' | 'OBCLUSTER_OVERVIEW';

  type TenantRole = 'PRIMARY' | 'STANDBY';

  type JobType = 'FULL' | 'INCR' | 'CLEAN' | 'ARCHIVE';

  type DestType = 'NFS' | 'OSS';

  type ReplicaDetailType = {
    iopsWeight: number;
    logDiskSize: string;
    maxIops: number;
    memorySize: string;
    maxCPU: string;
    minCPU: string;
    minIops: number;
    priority: number;
    type: string;
    zone: string;
  };

  type ScheduleDatesType = {
    backupType: 'Full' | 'Incremental';
    day: number;
  }[];

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
    topology: ReplicaDetailType[];
    unitNumber: number;
  }

  interface BackupPolicy {
    destType: DestType;
    archivePath: string;
    bakDataPath: string;
    scheduleType: string;
    scheduleTime: string;
    scheduleDates: ScheduleDatesType;
  }

  interface BackupJob {
    encryptionSecret: string;
    endTime: string;
    name: string;
    path: string;
    startTime: string;
    status: string;
    statusInDatabase: string;
    type: string;
  }
  [];

  type NamespaceAndName = {
    ns: string;
    name: string;
  };

  type UnitConfig = {
    iopsWeight?: number;
    logDiskSize?: string;
    cpuCount: string;
    maxIops?: number;
    memorySize: string;
    minIops?: number;
  }

  type PoolConfig = {
    priority: number;
    unitConfig: {
      iopsWeight: number;
      logDiskSize: string;
      cpuCount: string;
      maxIops: number;
      memorySize: string;
      minIops: number;
    };
  };

  type TenantBody = {
    connectWhiteList?: string;
    name: string;
    namespace?:string;
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
    unitConfig: UnitConfig;
    unitNum: number;
  };
  type TenantPolicy = {
    archivePath: string;
    bakDataPath: string;
    bakEncryptionPassword?: string;
    destType: DestType;
    jobKeepDays?: number;
    ossAccessId: string;
    ossAccessKey: string;
    pieceIntervalDays?: number;
    recoveryDays?: number;
    scheduleDates: ScheduleDatesType;
    scheduleTime: string;
    scheduleType: 'Weekly' | 'Monthly';
  };

  type UpdateTenantPolicy = {
    jobKeepDays?: number;
    pieceIntervalDays?: number;
    recoveryDays?: number;
    scheduleDates?: ScheduleDatesType;
    scheduleType?: 'Weekly' | 'Monthly';
    status?: string;
  };

  type StatisticData = {
    total: number;
    name: string;
    type: 'cluster' | 'tenant';
    deleting: number;
    operating: number;
    running: number;
    failed: number;
  };

  type ServerResource = {
    availableCPU: number;
    availableDataDisk: number;
    availableLogDisk: number;
    availableMemory: number;
    obServerIP: string;
    obZone: string;
  };

  type ZoneResource = {
    availableCPU: number;
    availableDataDisk: number;
    availableLogDisk: number;
    availableMemory: number;
    obZone: string;
  };

  type EssentialParametersType = {
    minPoolMemory: number;
    obServerResources: ServerResource[];
    obZoneResourceMap: {
      [T]: ZoneResource;
    };
  };

  interface CommonResponse {
    data: any;
    message: string;
    successful: boolean;
  }

  interface StatisticDataResponse extends CommonResponse {
    data: StatisticData;
  }

  interface TenantsListResponse extends CommonResponse {
    data: TenantDetail[];
  }

  interface BackupJobsResponse extends CommonResponse {
    data: BackupJob[];
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
      topology: ReplicaDetailType[];
      unitNumber: number;
    };
  }

  interface BackupPolicy {
    archivePath: string;
    bakDataPath: string;
    bakEncryptionSecret: string;
    destType: DestType;
    jobKeepDays: string;
    name: string;
    namespace: string;
    ossAccessSecret: string;
    pieceIntervalDays: string;
    recoveryDays: string;
    scheduleDates: ScheduleDatesType;
    scheduleTime: string;
    scheduleType: string;
    status: string;
    tenantName: string;
  }

  interface BackupPolicyResponse extends CommonResponse {
    data: BackupPolicy;
  }

  type BackupConfigEditable = {
    destType: DestType;
    jobKeepDays: number;
    pieceIntervalDays: number;
    recoveryDays: number;
    scheduleDates: ScheduleDatesType;
    scheduleTime: string;
    scheduleType: string;
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
    clusterResourceName: string;
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

  interface TenantBasicInfoResponse extends CommonResponse {
    data: TenantBasicInfo;
  }

  type TenantBasicInfo = {
    info: InfoType;
    source?: {
      primaryTenant?: string;
      archiveSource?: string;
      bakDataSource?: string;
      until?: string;
    };
    replicas?: ReplicaDetailType[];
  };

  interface EssentialParametersTypeResponse extends CommonResponse {
    data: EssentialParametersType;
  }
}
