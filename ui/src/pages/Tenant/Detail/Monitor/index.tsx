import type { FilterDataType,LabelType } from '@/components/MonitorDetail';
import MonitorDetail from '@/components/MonitorDetail';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { getTenant } from '@/services/tenant';
import { useRequest } from 'ahooks';
import { useEffect,useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';

import { getFilterData } from '@/components/MonitorDetail/helper';


export default function Monitor() {
  const [[ns, name, tenantName]] = useState(getNSName());
  const [filterLabel, setFilterLabel] = useState<LabelType[]>([
    {
      key: 'tenant_name',
      value: tenantName,
    },
  ]);
  const [filterData, setFilterData] = useState<FilterDataType>({
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
    getTenantDetail({ ns, name });
  }, []);
  const tenantDetail = tenantDetailResponse?.data;
  return (
    <MonitorDetail
      filterData={filterData}
      setFilterData={setFilterData}
      filterLabel={filterLabel}
      setFilterLabel={setFilterLabel}
      queryScope='OBTENANT'
      groupLabels={['tenant_name']}
      basicInfo={
        tenantDetail && (
          <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
        )
      }
    />
  );
}
