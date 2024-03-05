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
  };

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
    cpuCount: number;
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

  interface CommonResponse {
    data: any;
    message: string;
    successful: boolean;
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

  type UnitConfig = {
    iopsWeight: number;
    logDiskSize: string;
    maxIops: number;
    memorySize: string;
    cpuCount: number;
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
}
