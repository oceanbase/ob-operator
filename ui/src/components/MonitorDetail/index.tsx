import { useUpdateEffect } from 'ahooks';
import moment from 'moment';
import { useEffect, useRef, useState } from 'react';

import MonitorComp from '@/components/MonitorComp';
import { REFRESH_FREQUENCY } from '@/constants';
import DataFilter from './DataFilter';

export type Label =
  | 'ob_cluster_name'
  | 'ob_cluster_id'
  | 'tenant_name'
  | 'tenant_id'
  | 'svr_ip'
  | 'obzone';

export type LabelType = {
  key: Label;
  value: string;
};

export type OptionType = {
  label: string;
  value: string | number;
  zone?: string;
};

export type FilterDataType = {
  zoneList?: OptionType[];
  serverList?: OptionType[];
  date?: any;
};

export type QueryRangeType = {
  endTimestamp: number;
  startTimestamp: number;
  step: number;
};

interface MonitorDetailProps {
  filterData:FilterDataType;
  setFilterData:React.Dispatch<React.SetStateAction<FilterDataType>>;
  basicInfo:JSX.Element | null | undefined;
  filterLabel:LabelType[];
  setFilterLabel:React.Dispatch<React.SetStateAction<LabelType[]>>;
  groupLabels:API.LableKeys[];
  queryScope:API.EventObjectType;
}

const getDate = () => {
  return moment
    .unix(Math.ceil(new Date().valueOf() / 1000))
    .format('YYYY-MM-DD HH:mm:ss');
};

const defaultQueryRange = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};

//Query is somewhat similar to sql statement, label is equivalent to filter condition，for example: where label1=xxx and label2 = xxx
export default function MonitorDetail({
  filterData,
  setFilterData,
  filterLabel,
  setFilterLabel,
  basicInfo,
  groupLabels,
  queryScope
}:MonitorDetailProps) {
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [realTime, setRealTime] = useState<string>(getDate());
  const timerRef = useRef<NodeJS.Timeout>();
  const updateTimer = useRef<NodeJS.Timer>();
  const [queryRange, setQueryRange] =
    useState<QueryRangeType>(defaultQueryRange);
  const newQueryRangeRef = useRef<QueryRangeType>(); //Only used to solve the problem of not getting the latest value in interval

  useUpdateEffect(() => {
    if (isRefresh && !timerRef.current) {
      timerRef.current = setInterval(() => {
        let target;
        if (!newQueryRangeRef.current) {
          target = { ...queryRange };
        } else {
          target = { ...newQueryRangeRef.current };
        }
        newQueryRangeRef.current = {
          step: target.step,
          startTimestamp: target.startTimestamp + REFRESH_FREQUENCY,
          endTimestamp: target.endTimestamp + REFRESH_FREQUENCY,
        };
        setQueryRange(newQueryRangeRef.current);
      }, REFRESH_FREQUENCY * 1000);
    }
    if (!isRefresh) {
      clearInterval(timerRef.current);
    }

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [isRefresh]);

  //Real time update time
  useEffect(() => {
    updateTimer.current = setInterval(() => {
      setRealTime(getDate());
    }, REFRESH_FREQUENCY * 1000);
    return () => {
      if (updateTimer.current) clearInterval(updateTimer.current);
    };
  }, [isRefresh]);

  return (
    <div style={{ marginTop: 12 }}>
      { basicInfo }
      <DataFilter
        realTime={realTime}
        isRefresh={isRefresh}
        setIsRefresh={setIsRefresh}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        filterData={filterData}
        setFilterData={setFilterData}
        queryRange={queryRange}
        setQueryRange={setQueryRange}
      />
      <MonitorComp
        isRefresh={isRefresh}
        queryRange={queryRange}
        filterLabel={filterLabel}
        type="DETAIL"
        groupLabels={groupLabels}
        queryScope={queryScope}
      />
    </div>
  );
}
