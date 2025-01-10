import MonitorDetail from '@/components/MonitorDetail';
import { getTenant } from '@/services/tenant';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';

import { getFilterData } from '@/components/MonitorDetail/helper';

export default function Monitor() {
  const { ns, name, tenantName } = useParams();
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'tenant_name',
      value: tenantName!,
    },
  ]);
  const [filterData, setFilterData] = useState<Monitor.FilterDataType>({
    zoneList: [],
    date: '',
  });
  const { data: tenantDetailResponse, run: getTenantDetail } = useRequest(
    getTenant,
    {
      manual: true,
      onSuccess: ({ data, successful }) => {
        if (successful && data) {
          setFilterData(getFilterData(data));
        }
      },
    },
  );

  useEffect(() => {
    getTenantDetail({ ns: ns!, name: name! });
  }, []);
  const tenantDetail = tenantDetailResponse?.data;
  return (
    <PageContainer>
      <MonitorDetail
        filterData={filterData}
        setFilterData={setFilterData}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        queryScope="OBTENANT"
        groupLabels={['tenant_name']}
        basicInfo={
          tenantDetail && (
            <BasicInfo
              info={tenantDetail.info}
              source={tenantDetail.source}
              ns={ns}
              name={name}
            />
          )
        }
      />
    </PageContainer>
  );
}
