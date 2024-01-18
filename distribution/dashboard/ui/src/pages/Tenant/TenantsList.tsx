import { ProCard } from '@ant-design/pro-components';
import { Link } from '@umijs/max';
import { Button, Col, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';

import styles from './index.less';

interface TenantsListProps {
  tenantsList: API.TenantsListResponse | undefined;
  turnToCreateTenant: () => void;
}

const columns: ColumnsType<API.TenantDetail> = [
  {
    title: '资源名',
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Link to={`ns=${record.namespace}&nm=${record.name}`}>{value}</Link>
    ),
  },
  {
    title: '命名空间',
    dataIndex: 'namespace',
    key: 'namespace',
  },
  {
    title: '所属集群',
    dataIndex: 'clusterName',
    key: 'clusterName',
  },
  {
    title: '租户名',
    dataIndex: 'tenantName',
    key: 'tenantName',
  },
  {
    title: '租户角色',
    dataIndex: 'tenantRole',
    key: 'tenantRole',
  },
  {
    title: 'locality',
    dataIndex: 'locality',
    key: 'locality',
  },
  {
    title: 'Unit 数量',
    dataIndex: 'unitNumber',
    key: 'unitNumber',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: '创建时间',
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
          <h2>租户列表</h2>
          <Button onClick={turnToCreateTenant} type="primary">
            创建租户
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
