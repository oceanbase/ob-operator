import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Row } from 'antd';
import { useState } from 'react';

import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import { getAllTenants } from '@/services/tenant';
import TenantsList from './TenantsList';


import type { LabelType, QueryRangeType } from '../../components/MonitorDetail';

const defaultQueryRange: QueryRangeType = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};
// 租户概览页
export default function TenantPage() {
  const [filterLabel, setFilterLabel] = useState<LabelType[]>([]);
  const navigate = useNavigate();
  const { data: tenantsListResponse } = useRequest(getAllTenants, {
    onSuccess:({data,successful})=>{
      if(successful){
        setFilterLabel(data.map((item)=>({
          key:'tenant_name',
          value:item.tenantName
        })));
      }
    }
  });
  const handleAddCluster = () => navigate('new');
  const tenantsList = tenantsListResponse?.data;
  
  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        {tenantsList && (
          <TenantsList
            tenantsList={tenantsList}
            turnToCreateTenant={handleAddCluster}
          />
        )}

        <EventsTable objectType="OBTENANT" />
      </Row>
      <MonitorComp
        filterLabel={filterLabel}
        queryScope="OBTENANT"
        type="OVERVIEW"
        useFor='tenant'
        groupLabels={['tenant_name','ob_cluster_name']}
        queryRange={defaultQueryRange}
      />
    </PageContainer>
  );
}
