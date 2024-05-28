import EventsTable from '@/components/EventsTable';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { Row } from 'antd';
import ClusterList from './ClusterList';
import { useRequest } from 'ahooks';
import { obproxy } from '@/api';
export default function OBProxy() {
  const navigate = useNavigate();
  const handleAddCluster = () => navigate('new');
  const { data: obproxiesRes, loading} = useRequest(obproxy.listOBProxies);
  const obproxies = obproxiesRes?.data;
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <ClusterList
          loading={loading}
          obproxies={obproxies}
          handleAddCluster={handleAddCluster}
        />
        <EventsTable objectType="OBPROXY" />
      </Row>
    </PageContainer>
  );
}
