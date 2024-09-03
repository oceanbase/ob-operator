import { obproxy } from '@/api';
import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { DEFAULT_QUERY_RANGE, REFRESH_OBPROXY_TIME } from '@/constants';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import { useEffect, useRef } from 'react';
import ClusterList from './ClusterList';

export default function OBProxy() {
  const navigate = useNavigate();
  const handleAddCluster = () => navigate('new');
  const access = useAccess();
  const timer = useRef<NodeJS.Timeout | null>(null);
  const {
    data: obproxiesRes,
    loading,
    refresh,
  } = useRequest(obproxy.listOBProxies, {
    onSuccess({ data, successful }) {
      if (successful) {
        if (data.some((obcluster) => obcluster.status === 'Pending')) {
          if (!timer.current) {
            timer.current = setInterval(() => {
              refresh();
            }, REFRESH_OBPROXY_TIME);
          }
        } else if (timer.current) {
          clearInterval(timer.current);
          timer.current = null;
        }
      }
    },
  });
  const obproxies = obproxiesRes?.data;

  useEffect(() => {
    return () => {
      if (timer.current) {
        clearInterval(timer.current);
        timer.current = null;
      }
    };
  }, []);
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
        {access.systemread || access.systemwrite ? (
          <Col span={24}>
            <EventsTable objectType="OBPROXY" />
          </Col>
        ) : null}
      </Row>
      {access.systemread || access.systemwrite ? (
        <MonitorComp
          filterLabel={[]}
          queryScope="OBPROXY"
          type="OVERVIEW"
          groupLabels={['cluster']}
          queryRange={DEFAULT_QUERY_RANGE}
        />
      ) : null}
    </PageContainer>
  );
}
