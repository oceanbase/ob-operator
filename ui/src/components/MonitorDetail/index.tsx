import { useAccess } from '@umijs/max';
import { useUpdateEffect } from 'ahooks';
import dayjs from 'dayjs';
import { useEffect, useRef, useState } from 'react';

import MonitorComp from '@/components/MonitorComp';
import { DEFAULT_QUERY_RANGE, REFRESH_FREQUENCY } from '@/constants';
import { DATE_TIME_FORMAT } from '@/constants/datetime';
import DataFilter from './DataFilter';

interface MonitorDetailProps {
  filterData: Monitor.FilterDataType;
  setFilterData: React.Dispatch<React.SetStateAction<Monitor.FilterDataType>>;
  basicInfo: JSX.Element | null | undefined;
  filterLabel: Monitor.LabelType[];
  setFilterLabel: React.Dispatch<React.SetStateAction<Monitor.LabelType[]>>;
  groupLabels: API.LableKeys[];
  queryScope: API.EventObjectType;
}

const getDate = () => {
  return dayjs
    .unix(Math.ceil(new Date().valueOf() / 1000))
    .format(DATE_TIME_FORMAT);
};

//Query is somewhat similar to sql statement, label is equivalent to filter condition，for example: where label1=xxx and label2 = xxx
export default function MonitorDetail({
  filterData,
  setFilterData,
  filterLabel,
  setFilterLabel,
  basicInfo,
  groupLabels,
  queryScope,
}: MonitorDetailProps) {
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [realTime, setRealTime] = useState<string>(getDate());
  const access = useAccess();
  const timerRef = useRef<NodeJS.Timeout>();
  const updateTimer = useRef<NodeJS.Timer>();
  const [queryRange, setQueryRange] =
    useState<Monitor.QueryRangeType>(DEFAULT_QUERY_RANGE);
  const newQueryRangeRef = useRef<Monitor.QueryRangeType>(); //Only used to solve the problem of not getting the latest value in interval

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
      {basicInfo}
      {access.obclusterwrite ? (
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
      ) : null}
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
