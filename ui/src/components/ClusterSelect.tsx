import { getObclusterListReq } from '@/services';
import { history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Select } from 'antd';

const ClusterSelect = ({ value, ...restProps }) => {
  const { data: clusterListRes } = useRequest(getObclusterListReq, {
    defaultParams: [{}],
  });

  const clusterList = clusterListRes?.data || [];

  return (
    <Select
      onChange={(value) => {
        const clusterItem = clusterList.find(
          (item) => item.clusterId === value,
        );
        history.push(
          `/cluster/${clusterItem.namespace}/${clusterItem.name}/${clusterItem.clusterName}/overview`,
        );
        window.location.reload();
      }}
      value={value}
      options={clusterList?.map((item) => ({
        value: item.clusterId,
        label: `${item.clusterName}:${item.clusterId}`,
      }))}
      {...restProps}
    />
  );
};

export default ClusterSelect;
