import { MODE_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Link } from '@umijs/max';
import { Button, Card, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

import { COLOR_MAP } from '@/constants';
interface DataType {
  namespace: string;
  name: string;
  status: string;
  statusDetail: string;
  createTime: string;
  image: string;
  cpuPercent: string;
  memoryPercent: string;
  diskPercent: string;
  clusterName: string;
}

interface ClusterListProps {
  handleAddCluster: () => void;
  clusterList: any;
  loading: boolean;
}

const { Text } = Typography;

const columns: ColumnsType<DataType> = [
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.ResourceName',
      defaultMessage: '资源名',
    }),
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Text ellipsis={{ tooltip: value }}>
        <Link to={`${record.namespace}/${record.name}/${record.clusterName}`}>
          {value}
        </Link>
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Namespace',
      defaultMessage: '命名空间',
    }),
    dataIndex: 'namespace',
    key: 'namespace',
    render: (value) => <Text ellipsis={{ tooltip: value }}>{value}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.ClusterName',
      defaultMessage: '集群名',
    }),
    dataIndex: 'clusterName',
    key: 'clusterName',
    render: (value) => <Text ellipsis={{ tooltip: value }}>{value}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.ClusterMode',
      defaultMessage: '集群模式',
    }),
    dataIndex: 'mode',
    key: 'mode',
    render: (value) => (
      <Text ellipsis={{ tooltip: MODE_MAP.get(value)?.text }}>
        {MODE_MAP.get(value)?.text}
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.NumberOfZones',
      defaultMessage: 'Zone 数量',
    }),
    dataIndex: 'zoneCount',
    key: 'zoneCount',
    render: (_, record) => <span>{record?.topology?.length}</span>,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Image',
      defaultMessage: '镜像',
    }),
    dataIndex: 'image',
    width: '20%',
    key: 'image',
    render: (value) => <Text ellipsis={{ tooltip: value }}>{value}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
    render: (value, record) => (
      <Tag color={COLOR_MAP.get(value)}>
        {' '}
        {value === 'operating' ? (
          <Text
            style={{ maxWidth: 110, color: '#d48806', fontSize: 12 }}
            ellipsis={{ tooltip: `${value}/${record.statusDetail}` }}
          >
            {value}/{record.statusDetail}
          </Text>
        ) : (
          value
        )}{' '}
      </Tag>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.CreationTime',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'createTime',
    width: 178,
    key: 'createTime',
  },
];

export default function ClusterList({
  handleAddCluster,
  clusterList,
  loading,
}: ClusterListProps) {
  return (
    <Card
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'dashboard.pages.Cluster.ClusterList.ClusterList',
            defaultMessage: '集群列表',
          })}
        </h2>
      }
      extra={
        <Button onClick={handleAddCluster} type="primary">
          {intl.formatMessage({
            id: 'OBDashboard.pages.Cluster.ClusterList.CreateACluster',
            defaultMessage: '创建集群',
          })}
        </Button>
      }
    >
      <Table
        loading={loading}
        columns={columns}
        dataSource={clusterList}
        scroll={{ x: 1200 }}
        pagination={{ simple: true }}
        rowKey="name"
        bordered
        sticky
      />
    </Card>
  );
}
