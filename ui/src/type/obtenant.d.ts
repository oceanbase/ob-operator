declare namespace OBTenant {
  type MinResourceConfig = {
    minCPU: number;
    minMemory: number;
    minLogDisk: number;
    minIops: number;
    maxIops: number;
  };
  type MaxResourceType = {
    maxCPU?: number;
    maxLogDisk?: number;
    maxMemory?: number;
  };
  type ScheduleDates = {
    [T: number]: API.BackupType;
    days: number[];
    mode: API.ScheduleType;
  };
  type OperateType = 'edit' | 'create';
}
