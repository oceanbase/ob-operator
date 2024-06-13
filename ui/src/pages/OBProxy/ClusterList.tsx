import type { ObproxyOBProxyOverview } from '@/api/generated';
import { intl } from '@/utils/intl';
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
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.F92A2ED2',
      defaultMessage: 'OBProxy 资源名',
    }),
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Text ellipsis={{ tooltip: value }}>
        <Link to={`${record.namespace}/${record.name}`}>{value}</Link>
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.807D961E',
      defaultMessage: 'OBProxy 集群名称',
    }),
    dataIndex: 'proxyClusterName',
    key: 'proxyClusterName',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.B8BB1499',
      defaultMessage: '对应 OB 集群',
    }),
    dataIndex: 'obCluster',
    key: 'obCluster',
    render: (value) => <Text>{value.name}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.9928287A',
      defaultMessage: '版本',
    }),
    dataIndex: 'image',
    key: 'image',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.A62C101E',
      defaultMessage: '副本数',
    }),
    dataIndex: 'Replicas',
    key: 'Replicas',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.4172D3BD',
      defaultMessage: '服务 IP',
    }),
    dataIndex: 'serviceIp',
    key: 'serviceIp',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.C2D80635',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'creationTime',
    key: 'creationTime',
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.614C7DA9',
      defaultMessage: '状态',
    }),
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
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'src.pages.OBProxy.0608BC29',
              defaultMessage: '集群列表',
            })}
          </h2>
        }
        extra={
          <Button onClick={handleAddCluster} type="primary">
            {intl.formatMessage({
              id: 'src.pages.OBProxy.EFFF5E84',
              defaultMessage: '创建 OBProxy 集群',
            })}
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
