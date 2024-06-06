import type { ResponseK8sPodInfo } from '@/api/generated';
import { Card, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface NodeInfoProps {
  pods?: ResponseK8sPodInfo[];
}

const { Text } = Typography;

const columns: ColumnsType<ResponseK8sPodInfo> = [
  {
    title: 'Pod 名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: 'IP',
    dataIndex: 'podIP',
    key: 'podIP',
  },
  {
    title: 'SQL 端口',
    render: () => <Text>2883</Text>,
  },
  {
    title: '版本',
    dataIndex: 'containers',
    key: 'containers',
    render: (containers) => <Text>{containers[0]?.image || '-'}</Text>,
  },
  {
    title: '创建时间',
    dataIndex: 'startTime',
    key: 'startTime',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
  },
];

export default function NodeInfo({ pods }: NodeInfoProps) {
  return (
    <Card title={<h2 style={{ marginBottom: 0 }}>节点信息</h2>}>
      <Table columns={columns} rowKey="nodeName" dataSource={pods} />
    </Card>
  );
}
