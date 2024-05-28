import type { ObproxyOBProxyOverview } from '@/api/generated';
import { Link } from '@umijs/max';
import { Button, Card, Col, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface ClusterListProps {
  handleAddCluster: () => void;
  obproxies: ObproxyOBProxyOverview[] | undefined;
  loading: boolean;
}

const { Text } = Typography;

const columns: ColumnsType<ObproxyOBProxyOverview> = [
  {
    title: 'OBProxy 资源名',
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Text ellipsis={{ tooltip: value }}>
        <Link to={`${record.namespace}/${record.name}`}>{value}</Link>
      </Text>
    ),
  },
  {
    title: 'OBProxy 集群名称',
    dataIndex: 'proxyClusterName',
    key: 'proxyClusterName',
  },
  {
    title: '对应 OB 集群',
    dataIndex: 'obCluster',
    key: 'obCluster',
    render: (value) => <Text>{value.name}</Text>,
  },
  {
    title: '版本',
    dataIndex: 'image',
    key: 'image',
  },
  {
    title: 'Replicas',
    dataIndex: 'Replicas',
    key: 'Replicas',
  },
  {
    title: 'serviceIp',
    dataIndex: 'serviceIp',
    key: 'serviceIp',
  },
  {
    title: 'creationTime',
    dataIndex: 'creationTime',
    key: 'creationTime',
  },
  {
    title: 'status',
    dataIndex: 'status',
    key: 'status',
  },
];

export default function ClusterList({
  handleAddCluster,
  loading,
  obproxies,
}: ClusterListProps) {
  return (
    <Col span={24}>
      <Card
        title={<h2 style={{ marginBottom: 0 }}>集群列表</h2>}
        extra={
          <Button onClick={handleAddCluster} type="primary">
            创建 OBProxy 集群
          </Button>
        }
      >
        <Table
          loading={loading}
          columns={columns}
          dataSource={obproxies}
          pagination={{ simple: true }}
          rowKey="name"
          bordered
          sticky
        />
      </Card>
    </Col>
  );
}
