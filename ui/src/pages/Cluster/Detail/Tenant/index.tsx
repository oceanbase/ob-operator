import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { DEFAULT_QUERY_RANGE } from '@/constants';
import TenantsList from '@/pages/Tenant/TenantsList';
import { getClusterDetailReq } from '@/services';
import { getAllTenants } from '@/services/tenant';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useNavigate, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import BasicInfo from '../Overview/BasicInfo';

export default function Tenant() {
  const { ns, name, clusterName } = useParams();
  const access = useAccess();
  const navigate = useNavigate();
  const { data: tenantsListResponse, run: getTenantsList } = useRequest(
    getAllTenants,
    {
      manual: true,
    },
  );
  const { data: clusterDetail } = useRequest(getClusterDetailReq, {
    defaultParams: [{ name: name!, ns: ns! }],
    onSuccess: () => {
      getTenantsList({ obcluster: name });
    },
  });
  const tenantsList = tenantsListResponse?.data;
  const handleAddCluster = () => navigate(`/tenant/new?clusterName=${name}`);
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
          </Col>
        )}
        {tenantsList && (
          <Col span={24}>
            <TenantsList
              tenantsList={tenantsList}
              turnToCreateTenant={handleAddCluster}
            />
          </Col>
        )}
        {access.systemread || access.systemwrite ? (
          <Col span={24}>
            <EventsTable objectType="OBTENANT" />
          </Col>
        ) : null}

        {tenantsList && tenantsList.length > 0 && clusterDetail && (
          <Col span={24}>
            <MonitorComp
              key={`monitor-${clusterName}-${tenantsList.length}`}
              queryRange={DEFAULT_QUERY_RANGE}
              queryScope="OBTENANT"
              groupLabels={['tenant_name']}
              useFor="tenant"
              type="DETAIL"
              filterLabel={[
                { key: 'ob_cluster_name', value: clusterName || '' },
              ]}
              filterQueryMetric={[
                ...tenantsList.map((tenant) => ({
                  key: 'tenant_name' as API.LableKeys,
                  value: tenant.tenantName,
                })),
                { key: 'tenant_name' as API.LableKeys, value: 'sys' },
              ]}
            />
          </Col>
        )}
      </Row>
    </PageContainer>
  );
}
