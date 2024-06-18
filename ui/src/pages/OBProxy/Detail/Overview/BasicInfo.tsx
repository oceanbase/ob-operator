import IconTip from '@/components/IconTip';
import { OBPROXY_COLOR_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Card, Descriptions, Tag } from 'antd';

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
    <Card
      style={style}
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.4AB0B6DC',
            defaultMessage: '基本信息',
          })}
        </h2>
      }
    >
      <Descriptions column={3}>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.25425D3D',
            defaultMessage: '资源名称',
          })}
        >
          {name || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.E31DF8FD',
            defaultMessage: 'OBProxy 集群名',
          })}
        >
          {proxyClusterName || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.00DC8B97',
            defaultMessage: 'OB 集群',
          })}
        >
          {JSON.stringify(obCluster) || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={
            <IconTip
              content={intl.formatMessage({
                id: 'src.pages.OBProxy.Detail.Overview.E0E978D2',
                defaultMessage: 'OBProxy root 密码',
              })}
              tip={intl.formatMessage({
                id: 'src.pages.OBProxy.Detail.Overview.AEC19030',
                defaultMessage: 'root@proxysys 密码',
              })}
            />
          }
        >
          {proxySysSecret || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.B42C8F94',
            defaultMessage: '命名空间',
          })}
        >
          {namespace || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.CE7D4E98',
            defaultMessage: '状态',
          })}
        >
          {status ? (
            <Tag color={OBPROXY_COLOR_MAP.get(status)}>{status}</Tag>
          ) : (
            '-'
          )}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
}
