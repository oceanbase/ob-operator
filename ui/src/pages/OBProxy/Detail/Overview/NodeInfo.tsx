import type { ResponseK8sPodInfo } from '@/api/generated';
import { intl } from '@/utils/intl';
import { Card, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface NodeInfoProps {
  pods?: ResponseK8sPodInfo[];
}

const { Text } = Typography;

const columns: ColumnsType<ResponseK8sPodInfo> = [
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.Detail.Overview.3CF533BE',
      defaultMessage: 'Pod 名称',
    }),
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: 'IP',
    dataIndex: 'podIP',
    key: 'podIP',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.Detail.Overview.D21304DA',
      defaultMessage: 'SQL 端口',
    }),
    render: () => <Text>2883</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.Detail.Overview.A124D9CE',
      defaultMessage: '版本',
    }),
    dataIndex: 'containers',
    key: 'containers',
    render: (containers) => <Text>{containers[0]?.image || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.Detail.Overview.2F1969E2',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'startTime',
    key: 'startTime',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.Detail.Overview.80FE51DD',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
  },
];

export default function NodeInfo({ pods }: NodeInfoProps) {
  return (
    <Card
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.DF444F52',
            defaultMessage: '节点信息',
          })}
        </h2>
      }
    >
      <Table columns={columns} rowKey="nodeName" dataSource={pods} />
    </Card>
  );
}
