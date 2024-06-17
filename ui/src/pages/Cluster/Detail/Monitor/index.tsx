import MonitorDetail from '@/components/MonitorDetail';
import { getClusterDetailReq } from '@/services';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect,useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';

import { getFilterData } from '@/components/MonitorDetail/helper';


export default function Monitor() {
  const { ns, name, clusterName } = useParams();
  const [filterData, setFilterData] = useState<Monitor.FilterDataType>({
    zoneList: [],
    serverList: [],
    date: '',
  });
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'ob_cluster_name',
      value: clusterName!,
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
    getClusterDetail({ ns:ns!, name:name! });
  }, []);
  return (
    <PageContainer>
      <MonitorDetail
        filterData={filterData}
        setFilterData={setFilterData}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        groupLabels={['ob_cluster_name']}
        queryScope="OBCLUSTER"
        basicInfo={
          clusterDetail && (
            <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
          )
        }
      />
    </PageContainer>
  );
}
