import IconTip from '@/components/IconTip';
import { Card, Descriptions } from 'antd';

interface BasicInfoProps {
  name?: string;
  proxyClusterName?: string;
  obCluster?: {
    name: string;
    namespace: string;
  };
  proxySysSecret?: string;
  namespace?: string;
  status?: string;
  style?: React.CSSProperties;
}

export default function BasicInfo({
  name,
  proxyClusterName,
  obCluster,
  proxySysSecret,
  namespace,
  status,
  style,
}: BasicInfoProps) {
  return (
    <Card style={style} title={<h2 style={{ marginBottom: 0 }}>基本设置</h2>}>
      <Descriptions column={3}>
        <Descriptions.Item label="Name">{name || '-'}</Descriptions.Item>
        <Descriptions.Item label="OBProxy Cluster Name">
          {proxyClusterName || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="OBCluster">
          {JSON.stringify(obCluster) || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={<IconTip content="OBProxy root secret" tip={''} />}
        >
          {proxySysSecret || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Namespace">
          {namespace || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Status">{status || '-'}</Descriptions.Item>
      </Descriptions>
    </Card>
  );
}
