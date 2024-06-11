import type { AlarmMatcher } from '@/api/generated';
import type { Dayjs } from 'dayjs';

declare namespace Alert {
  type DrawerStatus = 'create' | 'edit' | 'display';
  type ShieldDrawerInitialValues = {
    instances?: InstancesType;
    matchers?: AlarmMatcher;
    rules?: string[];
  };
  type InstancesKey = 'obcluster' | 'observer' | 'obtenant';
  type AlarmLevel = 'critical' | 'warning' | 'caution' | 'info';
  type InstancesType = {
    type: InstancesKey;
    obcluster: string[];
    observer?: string[];
    obtenant?: string[];
  };
  type ShieldDrawerForm = {
    instances: InstancesType;
    matchers: {
      isRegex?: boolean;
      name?: string;
      value?: string;
    }[];
    endsAt: Dayjs;
    rules:string[];
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

  type LabelsType = {
    value?: string | undefined;
    name?: string | undefined;
    isRegex?: boolean;
    key?: string | undefined;
  };
}
