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
        <Descriptions.Item label="资源名称">{name || '-'}</Descriptions.Item>
        <Descriptions.Item label="OBProxy 集群名">
          {proxyClusterName || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="OB 集群">
          {JSON.stringify(obCluster) || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={
            <IconTip content="OBProxy root 密码" tip={'root@proxysys 密码'} />
          }
        >
          {proxySysSecret || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="命名空间">
          {namespace || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="状态">{status || '-'}</Descriptions.Item>
      </Descriptions>
    </Card>
  );
}
