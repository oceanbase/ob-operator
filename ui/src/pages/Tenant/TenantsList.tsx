import { COLOR_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Link } from '@umijs/max';
import { Button, Card, Col, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

import styles from './index.less';

interface TenantsListProps {
  tenantsList: API.TenantDetail[] | undefined;
  turnToCreateTenant: () => void;
  loading?: boolean;
}

const { Text } = Typography;
const columns: ColumnsType<API.TenantDetail> = [
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.ResourceName',
      defaultMessage: '资源名',
    }),
    dataIndex: 'name',
    key: 'name',
    width: '10%',
    render: (value, record) => (
      <Text>
        <Link
          replace
          to={`/tenant/${record.namespace}/${record.name}/${record.tenantName}`}
        >
          {value}
        </Link>
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.Namespace',
      defaultMessage: '命名空间',
    }),
    dataIndex: 'namespace',
    width: '10%',
    key: 'namespace',
    render: (value) => (
      <Text ellipsis={{ tooltip: value }}>
        {value}
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.Cluster',
      defaultMessage: '所属集群',
    }),
    dataIndex: 'clusterResourceName',
    key: 'clusterResourceName',
    width: '10%',
    render: (value) => (
      <Text ellipsis={{ tooltip: `${value}` }}>
        {value}
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.TenantName',
      defaultMessage: '租户名',
    }),
    dataIndex: 'tenantName',
    key: 'tenantName',
    width: '10%',
    render: (value) => <Text ellipsis={{ tooltip: value }}>{value}</Text>,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.TenantRole',
      defaultMessage: '租户角色',
    }),
    dataIndex: 'tenantRole',
    key: 'tenantRole',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.ReplicaDistribution',
      defaultMessage: '副本分布',
    }),
    width: '20%',
    dataIndex: 'locality',
    key: 'locality',
    render: (value) => (
      <Text ellipsis={{ tooltip: value }}>
        {value}
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.NumberOfUnits',
      defaultMessage: 'Unit 数量',
    }),
    dataIndex: 'unitNumber',
    key: 'unitNumber',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
    render: (value) => <Tag color={COLOR_MAP.get(value)}>{value} </Tag>,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.CreationTime',
      defaultMessage: '创建时间',
    }),
    width: 178,
    dataIndex: 'createTime',
    key: 'createTime',
  },
];

export default function TenantsList({
  tenantsList,
  turnToCreateTenant,
  loading,
}: TenantsListProps) {
  return (
    <Col span={24}>
      <Card
        loading={loading}
        title={
          <div className={styles.clusterHeader}>
            <h2 style={{ marginBottom: 0 }}>
              {intl.formatMessage({
                id: 'Dashboard.pages.Tenant.TenantsList.TenantList',
                defaultMessage: '租户列表',
              })}
            </h2>
            <Button onClick={turnToCreateTenant} type="primary">
              {intl.formatMessage({
                id: 'Dashboard.pages.Tenant.TenantsList.CreateATenant',
                defaultMessage: '创建租户',
              })}
            </Button>
          </div>
        }
      >
        <Table
          columns={columns}
          dataSource={tenantsList}
          scroll={{ x: 1200 }}
          pagination={{ simple: true }}
          rowKey="name"
          bordered
          sticky
        />
      </Card>
    </Col>
  );
}
