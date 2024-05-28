import { obproxy } from '@/api';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Col, Row } from 'antd';
import { useEffect } from 'react';
import BasicInfo from './BasicInfo';

export default function Overview() {
  const { ns, name } = useParams();
  const { data: obproxyDetailRes, run: getOBProxy } = useRequest(
    obproxy.getOBProxy,
  );
  const obproxyDetail = obproxyDetailRes?.data;
  useEffect(() => {
    getOBProxy(ns!, name!);
  }, []);

  return (
    <PageContainer title="OBProxy 详情">
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
      </Row>
    </PageContainer>
  );
}
