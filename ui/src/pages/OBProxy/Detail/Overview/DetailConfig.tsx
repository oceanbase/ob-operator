import type { CommonKVPair, CommonResourceSpec } from '@/api/generated';
import InputLabelComp from '@/components/InputLabelComp';
import { Button, Card, Col, Row } from 'antd';
import { useState } from 'react';
import ConfigDrawer from './ConfigDrawer';

interface DetailConfigProps {
  name?: string;
  namespace?: string;
  image?: string;
  parameters?: CommonKVPair[];
  resource?: CommonResourceSpec;
  serviceType?: string;
  replicas?: number;
  style?: React.CSSProperties;
}

export default function DetailConfig({
  style,
  ...props
}: DetailConfigProps) {
  const {image,serviceType,replicas,resource,parameters} = props;
  const [drawerOpen, setDrawerOpen] = useState(false);
  return (
    <Card
      title={<h2 style={{ marginBottom: 0 }}>详细配置</h2>}
      extra={
        <Button onClick={() => setDrawerOpen(true)} type="primary">
          编辑
        </Button>
      }
      style={style}
    >
      <div style={{ marginBottom: 24 }}>
        <h3>资源设置</h3>
        <Row gutter={[16, 16]}>
          <Col span={24}>部署镜像：{image || '-'}</Col>
          <Col span={24}>服务类型：{serviceType || '-'}</Col>
          <Col span={8}>副本数：{replicas || '-'}</Col>
          <Col span={8}>CPU 核数：{resource?.cpu || '-'}</Col>
          <Col span={8}>内存大小：{resource?.memory || '-'}</Col>
        </Row>
      </div>
      <div>
        <h3>参数设置</h3>
        <InputLabelComp disable={true} value={parameters || []} />
      </div>
      <ConfigDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={580}
        {...props}
      />
    </Card>
  );
}
