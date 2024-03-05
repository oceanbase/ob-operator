import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { Link } from '@umijs/max';
import { Button, Col, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { COLOR_MAP } from '@/constants';

import styles from './index.less';

interface TenantsListProps {
  tenantsList: API.TenantDetail[] | undefined;
  turnToCreateTenant: () => void;
}

const columns: ColumnsType<API.TenantDetail> = [
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.ResourceName',
      defaultMessage: '资源名',
    }),
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Link replace to={`/tenant/ns=${record.namespace}&nm=${record.name}&tenantName=${record.tenantName}`}>{value}</Link>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.Namespace',
      defaultMessage: '命名空间',
    }),
    dataIndex: 'namespace',
    key: 'namespace',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.Cluster',
      defaultMessage: '所属集群',
    }),
    dataIndex: 'clusterName',
    key: 'clusterName',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.TenantName',
      defaultMessage: '租户名',
    }),
    dataIndex: 'tenantName',
    key: 'tenantName',
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
    title: 'locality',
    dataIndex: 'locality',
    key: 'locality',
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
    render:(value)=><Tag color={COLOR_MAP.get(value)}>{value} </Tag>,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Tenant.TenantsList.CreationTime',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'createTime',
    key: 'createTime',
  },
];

export default function TenantsList({
  tenantsList,
  turnToCreateTenant,
}: TenantsListProps) {
  return (
    <Col span={24}>
      <ProCard>
        <div className={styles.clusterHeader}>
          <h2>
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
        <Table
          columns={columns}
          dataSource={tenantsList}
          scroll={{ x: 1200 }}
          pagination={{ simple: true }}
          rowKey="name"
          bordered
          sticky
        />
      </ProCard>
    </Col>
  );
}
