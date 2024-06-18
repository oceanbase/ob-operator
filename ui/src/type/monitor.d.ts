declare namespace Monitor {
  type Label =
    | 'ob_cluster_name'
    | 'ob_cluster_id'
    | 'tenant_name'
    | 'tenant_id'
    | 'svr_ip'
    | 'obzone'
    | 'cluster';

  type LabelType = {
    key: Label;
    value: string;
  };

  type OptionType = {
    label: string;
    value: string | number;
    zone?: string;
  };

  type FilterDataType = {
    zoneList?: OptionType[];
    serverList?: OptionType[];
    date?: any;
  };

  type QueryRangeType = {
    endTimestamp: number;
    startTimestamp: number;
    step: number;
  };
}
