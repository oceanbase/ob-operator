import { obproxy } from '@/api';
import MonitorComp from '@/components/MonitorComp';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect } from 'react';
import BasicInfo from '../Overview/BasicInfo';
import { DEFAULT_QUERY_RANGE } from '@/constants';

export default function Monitor() {
  const { ns, name } = useParams();
  const { data: obproxyDetailRes, run: getOBProxy } = useRequest(
    obproxy.getOBProxy,
    {
      manual: true,
    },
  );
  const obproxyDetail = obproxyDetailRes?.data;

  useEffect(() => {
    getOBProxy(ns!, name!);
  }, []);
  return (
    <PageContainer>
      <BasicInfo
        name={obproxyDetail?.name}
        namespace={obproxyDetail?.namespace}
        status={obproxyDetail?.status}
        obCluster={obproxyDetail?.obCluster}
        proxySysSecret={obproxyDetail?.proxySysSecret}
        proxyClusterName={obproxyDetail?.proxyClusterName}
      />
      <MonitorComp
        filterLabel={[]}
        queryScope="OBPROXY"
        type="DETAIL"
        groupLabels={['cluster']}
        queryRange={DEFAULT_QUERY_RANGE}
      />
    </PageContainer>
  );
}
