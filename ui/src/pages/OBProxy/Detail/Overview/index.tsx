import { obproxy } from '@/api';
import EventsTable from '@/components/EventsTable';
import { REFRESH_OBPROXY_TIME } from '@/constants';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import { useEffect, useRef } from 'react';
import BasicInfo from './BasicInfo';
import DetailConfig from './DetailConfig';
import NodeInfo from './NodeInfo';

export default function Overview() {
  const { ns, name } = useParams();
  const timer = useRef<NodeJS.Timeout>();
  const {
    data: obproxyDetailRes,
    run: getOBProxy,
    refresh,
  } = useRequest(obproxy.getOBProxy, {
    manual: true,
    onSuccess: ({ successful, data }) => {
      if (successful) {
        if (data.status === 'Pending') {
          timer.current = setTimeout(() => {
            refresh();
          }, REFRESH_OBPROXY_TIME);
        } else {
          clearTimeout(timer.current);
        }
      }
    },
  });
  const obproxyDetail = obproxyDetailRes?.data;
  useEffect(() => {
    getOBProxy(ns!, name!);
  }, []);

  return (
    <PageContainer
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.Detail.Overview.1CA5DF47',
        defaultMessage: 'OBProxy 详情',
      })}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <BasicInfo
            name={obproxyDetail?.name}
            namespace={obproxyDetail?.namespace}
            status={obproxyDetail?.status}
            obCluster={obproxyDetail?.obCluster}
            proxySysSecret={obproxyDetail?.proxySysSecret}
            proxyClusterName={obproxyDetail?.proxyClusterName}
          />
        </Col>
        <Col span={24}>
          <DetailConfig
            name={obproxyDetail?.name}
            namespace={obproxyDetail?.namespace}
            image={obproxyDetail?.image}
            parameters={obproxyDetail?.parameters}
            resource={obproxyDetail?.resource}
            replicas={obproxyDetail?.replicas}
            serviceType={obproxyDetail?.service.type}
            submitCallback={refresh}
          />
        </Col>
        <Col span={24}>
          <NodeInfo pods={obproxyDetail?.pods} />
        </Col>
        <Col span={24}>
          {obproxyDetail?.name && (
            <EventsTable objectType={'OBPROXY'} name={obproxyDetail?.name} />
          )}
        </Col>
      </Row>
    </PageContainer>
  );
}
