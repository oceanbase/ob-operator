import { useRequest, useUpdateEffect } from 'ahooks';
import moment from 'moment';
import { useEffect, useRef, useState } from 'react';

import MonitorComp from '@/components/MonitorComp';
import { REFRESH_FREQUENCY } from '@/constants';
import { getClusterDetailReq } from '@/services';
import BasicInfo from '../Overview/BasicInfo';
import { getNSName } from '../Overview/helper';
import DataFilter from './DataFilter';
import { getFilterData } from './helper';

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
  zoneList: OptionType[];
  serverList: OptionType[];
  date: any;
};

export type QueryRangeType = {
  endTimestamp: number;
  startTimestamp: number;
  step: number; //每个点位间隔时间s 例如 半小时1800s 间隔step 30s 返回60个点位
};

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

//查询和sql语句有些类似，label相当于过滤条件，就是where label1=xxx and label2 = xxx
export default function Monitor() {
  const clusterName = location.hash.split('&')[1].split('=')[1].split('/')[0];
  const [[ns, name]] = useState(getNSName());
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [realTime, setRealTime] = useState<string>(getDate());
  const timerRef = useRef<NodeJS.Timeout>();
  const updateTimer = useRef<NodeJS.Timer>();
  const [queryRange, setQueryRange] =
    useState<QueryRangeType>(defaultQueryRange);
  const newQueryRangeRef = useRef<QueryRangeType>(); //仅用于解决interval中拿不到最新值的问题
  const [filterData, setFilterData] = useState<FilterDataType>({
    zoneList: [],
    serverList: [],
    date: '',
  });
  const [filterLabel, setFilterLable] = useState<LabelType[]>([
    {
      key: 'ob_cluster_name',
      value: clusterName,
    },
  ]);
  const { data: clusterDetail, run: getClusterDetail } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onSuccess: (data) => {
        if (data) {
          setFilterData(getFilterData(data));
        }
      },
    },
  );

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

  //实时更新时间
  useEffect(() => {
    updateTimer.current = setInterval(() => {
      setRealTime(getDate());
    }, REFRESH_FREQUENCY * 1000);
    return () => {
      if (updateTimer.current) clearInterval(updateTimer.current);
    };
  }, [isRefresh]);

  useEffect(() => {
    getClusterDetail({ ns, name });
  }, []);

  return (
    <div style={{ marginTop: 12 }}>
      {clusterDetail && (
        <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
      )}
      <DataFilter
        realTime={realTime}
        isRefresh={isRefresh}
        setIsRefresh={setIsRefresh}
        filterLabel={filterLabel}
        setFilterLable={setFilterLable}
        filterData={filterData}
        setFilterData={setFilterData}
        queryRange={queryRange}
        setQueryRange={setQueryRange}
      />
      <MonitorComp
        isRefresh={isRefresh}
        queryRange={queryRange}
        filterLabel={filterLabel}
        type='detail'
        queryScope='OBCLUSTER'
      />
    </div>
  );
}
