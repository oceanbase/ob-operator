import { getAllTenants } from '@/services/tenant';
import { history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Select } from 'antd';

const TenantSelect = ({ value, ...restProps }) => {
  const { data: tenantsListResponse } = useRequest(getAllTenants, {
    defaultParams: [{}],
  });

  const tenantsList = tenantsListResponse?.data || [];

  return (
    <Select
      onChange={(value) => {
        const tenantItem = tenantsList.find(
          (item) => item.tenantName === value,
        );

        history.push(
          `/tenant/${tenantItem.namespace}/${tenantItem.name}/${tenantItem.tenantName}/overview`,
        );
        window.location.reload();
      }}
      value={value}
      options={tenantsList?.map((item) => ({
        value: item.tenantName,
        label: item.tenantName,
      }))}
      popupMatchSelectWidth={120}
      {...restProps}
    />
  );
};

export default TenantSelect;
