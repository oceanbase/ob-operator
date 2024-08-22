import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import { useEffect, useRef, useState } from 'react';

import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import {
  DEFAULT_QUERY_RANGE,
  REFRESH_TENANT_TIME,
  RESULT_STATUS,
} from '@/constants';
import { getAllTenants } from '@/services/tenant';
import TenantsList from './TenantsList';

// tenant overview page
export default function TenantPage() {
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([]);
  const access = useAccess();
  const navigate = useNavigate();
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const {
    data: tenantsListResponse,
    refresh: reGetAllTenants,
    loading,
  } = useRequest(getAllTenants, {
    onSuccess: ({ data, successful }) => {
      if (successful) {
        const operatingTenant = data.find(
          (tenant) => !RESULT_STATUS.includes(tenant.status),
        );
        if (operatingTenant) {
          if (!timerRef.current)
            timerRef.current = setInterval(() => {
              reGetAllTenants();
            }, REFRESH_TENANT_TIME);
        } else if (timerRef.current) {
          clearInterval(timerRef.current);
          timerRef.current = null;
        }
        setFilterLabel(
          data.map((item) => ({
            key: 'tenant_name',
            value: item.tenantName,
          })),
        );
      }
    },
  });
  const handleAddCluster = () => navigate('new');
  const tenantsList = tenantsListResponse?.data;

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, []);

  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <TenantsList
            loading={loading}
            tenantsList={tenantsList}
            turnToCreateTenant={handleAddCluster}
          />
        </Col>
        {access.systemread || access.systemwrite ? (
          <Col span={24}>
            <EventsTable objectType="OBTENANT" collapsible={false} />
          </Col>
        ) : null}
      </Row>
      {access.systemread || access.systemwrite ? (
        <MonitorComp
          filterLabel={filterLabel}
          queryScope="OBTENANT"
          type="OVERVIEW"
          useFor="tenant"
          groupLabels={['tenant_name', 'ob_cluster_name']}
          queryRange={DEFAULT_QUERY_RANGE}
          filterData={tenantsList}
        />
      ) : null}
    </PageContainer>
  );
}
