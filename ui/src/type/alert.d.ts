import type { AlarmMatcher, OceanbaseOBInstance } from '@/api/generated';

declare namespace Alert {
  type DrawerStatus = 'create' | 'edit' | 'display';
  type ShieldDrawerInitialValues = {
    instance?: OceanbaseOBInstance;
    matchers?: AlarmMatcher;
  };
}
