import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Row } from 'antd';
import { useEffect,useRef,useState } from 'react';

import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { REFRESH_TENANT_TIME, RESULT_STATUS } from '@/constants';
import { getAllTenants } from '@/services/tenant';
import TenantsList from './TenantsList';

const defaultQueryRange: Monitor.QueryRangeType = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};
// tenant overview page
export default function TenantPage() {
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([]);
  const navigate = useNavigate();
  const timerRef = useRef<NodeJS.Timeout>();
  const { data: tenantsListResponse, refresh: reGetAllTenants, loading } = useRequest(
    getAllTenants,
    {
      onSuccess: ({ data, successful }) => {
        if (successful) {
          const operatingTenant = data.find(
            (tenant) => !RESULT_STATUS.includes(tenant.status),
          );
          if (operatingTenant) {
            timerRef.current = setTimeout(() => {
              reGetAllTenants();
            }, REFRESH_TENANT_TIME);
          } else if (timerRef.current) {
            clearTimeout(timerRef.current);
          }
          setFilterLabel(
            data.map((item) => ({
              key: 'tenant_name',
              value: item.tenantName,
            })),
          );
        }
      },
    },
  );
  const handleAddCluster = () => navigate('new');
  const tenantsList = tenantsListResponse?.data;

  useEffect(()=>{
    return()=>{
      if(timerRef.current){
        clearTimeout(timerRef.current);
      }
    }
  },[])
  
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <TenantsList
          loading={loading}
          tenantsList={tenantsList}
          turnToCreateTenant={handleAddCluster}
        />
        <EventsTable objectType="OBTENANT" collapsible={false}/>
      </Row>
      <MonitorComp
        filterLabel={filterLabel}
        queryScope="OBTENANT"
        type="OVERVIEW"
        useFor='tenant'
        groupLabels={['tenant_name','ob_cluster_name']}
        queryRange={defaultQueryRange}
        filterData={tenantsList}
      />
    </PageContainer>
  );
}
