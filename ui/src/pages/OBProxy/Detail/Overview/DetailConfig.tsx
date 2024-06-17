import InputLabelComp from '@/components/InputLabelComp';
import { OBProxy } from '@/type/obproxy';
import { intl } from '@/utils/intl';
import { Button, Card, Col, Row } from 'antd';
import { useState } from 'react';
import ConfigDrawer from './ConfigDrawer';
interface DetailConfigProps extends OBProxy.CommonProxyDetail {
  style?: React.CSSProperties;
}

export default function DetailConfig({ style, ...props }: DetailConfigProps) {
  const { image, serviceType, replicas, resource, parameters } = props;

  const [drawerOpen, setDrawerOpen] = useState(false);
  return (
    <Card
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.DFE99A80',
            defaultMessage: '详细配置',
          })}
        </h2>
      }
      extra={
        <Button onClick={() => setDrawerOpen(true)} type="primary">
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.6258C614',
            defaultMessage: '编辑',
          })}
        </Button>
      }
      style={style}
    >
      <div style={{ marginBottom: 24 }}>
        <h3>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.6AE22B46',
            defaultMessage: '资源设置',
          })}
        </h3>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.9D704A15',
              defaultMessage: '部署镜像：',
            })}
            {image || '-'}
          </Col>
          <Col span={24}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.23ED4374',
              defaultMessage: '服务类型：',
            })}
            {serviceType || '-'}
          </Col>
          <Col span={8}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.FEFC3429',
              defaultMessage: '副本数：',
            })}
            {replicas || '-'}
          </Col>
          <Col span={8}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.6137B053',
              defaultMessage: 'CPU 核数：',
            })}
            {resource?.cpu || '-'}
          </Col>
          <Col span={8}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.5DDD1A0A',
              defaultMessage: '内存大小：',
            })}
            {resource?.memory ? `${resource.memory}GB` : '-'}
          </Col>
        </Row>
      </div>
      <div>
        <h3>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.D9C95C8E',
            defaultMessage: '参数设置',
          })}
        </h3>
        <InputLabelComp
          allowDelete={false}
          disable={true}
          value={parameters || []}
        />
      </div>
      {props.name && props.namespace ? (
        <ConfigDrawer
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          width={880}
          {...props}
        />
      ) : null}
    </Card>
  );
}
