import { obproxy } from '@/api';
import EventsTable from '@/components/EventsTable';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { REFRESH_OBPROXY_TIME } from '@/constants';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { history, useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Row, message } from 'antd';
import { useEffect, useRef } from 'react';
import BasicInfo from './BasicInfo';
import DetailConfig from './DetailConfig';
import NodeInfo from './NodeInfo';

export default function Overview() {
  const { ns, name } = useParams();
  const access = useAccess();
  const timer = useRef<NodeJS.Timeout | null>(null);
  const {
    data: obproxyDetailRes,
    run: getOBProxy,
    refresh,
  } = useRequest(obproxy.getOBProxy, {
    manual: true,
    onSuccess: ({ successful, data }) => {
      if (successful) {
        if (data.status === 'Pending' && !timer.current) {
          timer.current = setInterval(() => {
            refresh();
          }, REFRESH_OBPROXY_TIME);
        } else if (data.status !== 'Pending' && timer.current) {
          clearInterval(timer.current);
          timer.current = null;
        }
      }
    },
  });
  const obproxyDetail = obproxyDetailRes?.data;
  const deleteCluster = async () => {
    const res = await obproxy.deleteOBProxy(ns!, name!);
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'src.pages.OBProxy.Detail.Overview.5015890A',
          defaultMessage: '删除成功',
        }),
      );
      history.replace('/obproxy');
    }
  };

  useEffect(() => {
    getOBProxy(ns!, name!);

    return () => {
      if (timer.current) {
        clearInterval(timer.current);
        timer.current = null;
      }
    };
  }, []);

  return (
    <PageContainer
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.Detail.Overview.1CA5DF47',
        defaultMessage: 'OBProxy 详情',
      })}
      extra={
        access.obproxywrite ? (
          <Button
            onClick={() =>
              showDeleteConfirm({
                onOk: deleteCluster,
                title: intl.formatMessage({
                  id: 'src.pages.OBProxy.Detail.Overview.A9E634FB',
                  defaultMessage: '确认删除该 OBProxy 吗？',
                }),
              })
            }
            type="primary"
            danger
          >
            {intl.formatMessage({
              id: 'OBDashboard.Detail.Overview.Delete',
              defaultMessage: '删除',
            })}
          </Button>
        ) : null
      }
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
            service={obproxyDetail?.service}
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
