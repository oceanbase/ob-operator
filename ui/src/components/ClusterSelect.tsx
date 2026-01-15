import { getObclusterListReq } from '@/services';
import { history, useLocation } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Select } from 'antd';

interface ClusterSelectProps {
  value?: string;
  [key: string]: any;
}

const ClusterSelect = ({ value, ...restProps }: ClusterSelectProps) => {
  const location = useLocation();
  const { data: clusterListRes } = useRequest(getObclusterListReq, {
    defaultParams: [],
  });

  const clusterList = clusterListRes?.data || [];

  return (
    <Select
      onChange={(selectedValue: string) => {
        const id = selectedValue.split(':')[1];
        const clusterItem = clusterList.find(
          (item: any) => String(item.clusterId) === id,
        );

        if (!clusterItem) {
          return;
        }

        // 获取当前路径
        const currentPath = location.pathname;

        // 解析当前路径，提取子页面路径（如 overview, topo, monitor 等）
        // 路径格式: /cluster/:ns/:name/:clusterName/:subPage
        const pathParts = currentPath.split('/');
        const clusterIndex = pathParts.findIndex((part) => part === 'cluster');

        let targetPath = `/cluster/${clusterItem.namespace}/${clusterItem.name}/${clusterItem.clusterName}`;

        // 如果当前路径有子页面（如 /overview, /topo 等），保持该子页面
        if (clusterIndex !== -1 && pathParts.length > clusterIndex + 4) {
          const subPage = pathParts.slice(clusterIndex + 4).join('/');
          targetPath = `${targetPath}/${subPage}`;
        } else {
          // 如果没有子页面，默认跳转到 overview
          targetPath = `${targetPath}/overview`;
        }

        history.push(targetPath);
        window.location.reload();
      }}
      value={value}
      options={clusterList?.map((item: any) => ({
        value: `${item.clusterName}:${item.clusterId}`,
        label: `${item.clusterName}:${item.clusterId}`,
      }))}
      popupMatchSelectWidth={120}
      {...restProps}
    />
  );
};

export default ClusterSelect;
