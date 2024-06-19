import type { ObproxyOBProxyOverview } from '@/api/generated';
import { OBPROXY_COLOR_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Link } from '@umijs/max';
import { Button, Card, Col, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

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
    render: (value) => <Text>{value || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.B8BB1499',
      defaultMessage: '对应 OB 集群',
    }),
    dataIndex: 'obCluster',
    key: 'obCluster',
    render: (value) => <Text>{value.name || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.9928287A',
      defaultMessage: '版本',
    }),
    dataIndex: 'image',
    width: '20%',
    key: 'image',
    render: (value) => (
      <Text ellipsis={{ tooltip: value }}>{value || '-'}</Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.A62C101E',
      defaultMessage: '副本数',
    }),
    dataIndex: 'replicas',
    key: 'replicas',
    render: (value) => <Text>{value || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.New.88D0BC94',
      defaultMessage: '服务类型',
    }),
    dataIndex: 'serviceType',
    key: 'serviceType',
    render: (value) => <Text>{value || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.4172D3BD',
      defaultMessage: '服务 IP',
    }),
    dataIndex: 'serviceIp',
    key: 'serviceIp',
    render: (value) => <Text>{value || '-'}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.C2D80635',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'creationTime',
    width: 178,
    key: 'creationTime',
    render: (value) => (
      <span>{dayjs.unix(value).format('YYYY-MM-DD HH:mm:ss') || '-'}</span>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'src.pages.OBProxy.614C7DA9',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
    render: (value) => <Tag color={OBPROXY_COLOR_MAP.get(value)}>{value}</Tag>,
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
          scroll={{ x: 1200 }}
          bordered
          sticky
        />
      </Card>
    </Col>
  );
}
