import type {
  AlarmMatcher,
  OceanbaseOBInstance,
  OceanbaseOBInstanceType,
} from '@/api/generated';
import type { Dayjs } from 'dayjs';

declare namespace Alert {
  type DrawerStatus = 'create' | 'edit' | 'display';
  type ShieldDrawerInitialValues = {
    instances?: OceanbaseOBInstance;
    matchers?: AlarmMatcher;
    rules?: string[];
  };
  type InstancesKey = 'obcluster' | 'observer' | 'obtenant';
  type InstancesType = {
    type: OceanbaseOBInstanceType;
    obcluster: string[];
    observer?: string[];
    obtenant?: string[];
  };
  type ShieldDrawerForm = {
    instances: InstancesType;
    matchers?: {
      isRegex: boolean;
      name: string;
      value: string;
    }[];
    endsAt: Dayjs;
    id?: string;
    comment: string;
  };
  type SelectList = string[] | TenantsList[] | ServersList[];

  type TenantsList = {
    clusterName: string;
    tenants?: string[];
  };

  type ServersList = {
    clusterName: string;
    servers?: string[];
  };

  type InstanceParamType = {
    type: InstancesKey;
    observer?: string;
    obtenant?: string;
    obcluster?: string;
  };
}
