import { obproxy } from '@/api';
import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { DEFAULT_QUERY_RANGE, REFRESH_OBPROXY_TIME } from '@/constants';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import { useRef } from 'react';
import ClusterList from './ClusterList';

export default function OBProxy() {
  const navigate = useNavigate();
  const handleAddCluster = () => navigate('new');
  const timer = useRef<NodeJS.Timeout>();
  const {
    data: obproxiesRes,
    loading,
    refresh,
  } = useRequest(obproxy.listOBProxies, {
    onSuccess({ data, successful }) {
      if (successful) {
        if (data.some((obcluster) => obcluster.status === 'Pending')) {
          timer.current = setTimeout(() => {
            refresh();
          }, REFRESH_OBPROXY_TIME);
        } else if (timer.current) {
          clearTimeout(timer.current);
        }
      }
    },
  });
  const obproxies = obproxiesRes?.data;
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <ClusterList
            loading={loading}
            obproxies={obproxies}
            handleAddCluster={handleAddCluster}
          />
        </Col>
        <Col span={24}>
          <EventsTable objectType="OBPROXY" />
        </Col>
      </Row>
      <MonitorComp
        filterLabel={[]}
        queryScope="OBPROXY"
        type="OVERVIEW"
        groupLabels={['cluster']}
        queryRange={DEFAULT_QUERY_RANGE}
      />
    </PageContainer>
  );
}
