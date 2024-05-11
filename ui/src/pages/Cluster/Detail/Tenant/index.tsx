import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import TenantsList from '@/pages/Tenant/TenantsList';
import { getClusterDetailReq } from '@/services';
import { getAllTenants } from '@/services/tenant';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import BasicInfo from '../Overview/BasicInfo';

const defaultQueryRange = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};

export default function Tenant() {
  const { ns, name, clusterName } = useParams();
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
  const handleAddCluster = () => navigate('/tenant/new');
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
        )}
        {tenantsList && (
          <TenantsList
            tenantsList={tenantsList}
            turnToCreateTenant={handleAddCluster}
          />
        )}
        <EventsTable objectType="OBTENANT" />
        {tenantsList && (
          <Col span={24}>
            <MonitorComp
              queryRange={defaultQueryRange}
              type="OVERVIEW"
              queryScope="OBTENANT"
              groupLabels={['tenant_name']}
              useFor="tenant"
              filterLabel={[{ key: 'ob_cluster_name', value: clusterName }]}
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
