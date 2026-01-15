import { getAllTenants } from '@/services/tenant';
import { history, useLocation } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Select } from 'antd';

interface TenantSelectProps {
  value?: string;
  [key: string]: any;
}

const TenantSelect = ({ value, ...restProps }: TenantSelectProps) => {
  const location = useLocation();
  const { data: tenantsListResponse } = useRequest(getAllTenants, {
    defaultParams: [],
  });

  const tenantsList = tenantsListResponse?.data || [];

  return (
    <Select
      onChange={(selectedValue: string) => {
        const tenantItem = tenantsList.find(
          (item: any) => item.tenantName === selectedValue,
        );

        if (!tenantItem) {
          return;
        }

        // 获取当前路径
        const currentPath = location.pathname;

        // 解析当前路径，提取子页面路径（如 overview, topo, backup, sql 等）
        // 路径格式: /tenant/:ns/:name/:tenantName/:subPage
        const pathParts = currentPath.split('/');
        const tenantIndex = pathParts.findIndex((part) => part === 'tenant');

        let targetPath = `/tenant/${tenantItem.namespace}/${tenantItem.name}/${tenantItem.tenantName}`;

        // 如果当前路径有子页面（如 /overview, /topo, /sql/:sqlId 等），保持该子页面
        if (tenantIndex !== -1 && pathParts.length > tenantIndex + 4) {
          const subPage = pathParts.slice(tenantIndex + 4).join('/');
          targetPath = `${targetPath}/${subPage}`;
        } else {
          // 如果没有子页面，默认跳转到 overview
          targetPath = `${targetPath}/overview`;
        }

        history.push(targetPath);
        window.location.reload();
      }}
      value={value}
      options={tenantsList?.map((item: any) => ({
        value: item.tenantName,
        label: item.tenantName,
      }))}
      popupMatchSelectWidth={120}
      {...restProps}
    />
  );
};

export default TenantSelect;
