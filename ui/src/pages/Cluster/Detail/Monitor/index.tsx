import MonitorDetail from '@/components/MonitorDetail';
import { getClusterDetailReq } from '@/services';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';
import { getNSName } from '../Overview/helper';
import type { FilterDataType,LabelType } from '@/components/MonitorDetail';

import { getFilterData } from '@/components/MonitorDetail/helper';


export default function Monitor() {
  const [[ns, name, clusterName]] = useState(getNSName());
  const [filterData, setFilterData] = useState<FilterDataType>({
    zoneList: [],
    serverList: [],
    date: '',
  });
  const [filterLabel, setFilterLabel] = useState<LabelType[]>([
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

  useEffect(() => {
    getClusterDetail({ ns, name });
  }, []);
  return (
    <MonitorDetail
      filterData={filterData}
      setFilterData={setFilterData}
      filterLabel={filterLabel}
      setFilterLabel={setFilterLabel}
      groupLabels={['ob_cluster_name']}
      queryScope='OBCLUSTER'
      basicInfo={
        clusterDetail && (
          <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
        )
      }
    />
  );
}
