import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Row } from 'antd';
import { useState } from 'react';

import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { DEFAULT_QUERY_RANGE } from '@/constants';
import { getObclusterListReq } from '@/services';
import ClusterList from './ClusterList';

const ClusterPage: React.FC = () => {
  const navigate = useNavigate();
  const [clusterNames, setClusterNames] = useState<Monitor.LabelType[]>([]);

  const { data: clusterListRes, loading } = useRequest(getObclusterListReq, {
    onSuccess: ({ successful, data }) => {
      if (successful) {
        const clusterNames: Monitor.LabelType[] = data.map((item) => ({
          key: 'ob_cluster_name',
          value: item.clusterName,
        }));
        setClusterNames(clusterNames);
      }
    },
  });

  const handleAddCluster = () => navigate('new');
  const clusterList = clusterListRes?.data;

  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <ClusterList
          loading={loading}
          clusterList={clusterList}
          handleAddCluster={handleAddCluster}
        />
        <EventsTable objectType="OBCLUSTER" />
      </Row>
      <MonitorComp
        filterLabel={clusterNames}
        queryScope="OBCLUSTER_OVERVIEW"
        type="OVERVIEW"
        groupLabels={['ob_cluster_name']}
        queryRange={DEFAULT_QUERY_RANGE}
        filterData={clusterList}
      />
    </PageContainer>
  );
};

export default ClusterPage;
