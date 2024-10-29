import { STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { Card, Col, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';

const getServerColums = () => {
  const serverColums: ColumnsType<API.Server> = [
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.ServerName',
        defaultMessage: 'Server名',
      }),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Namespace',
        defaultMessage: '命名空间',
      }),
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Address',
        defaultMessage: '地址',
      }),
      dataIndex: 'address',
      key: 'address',
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Status',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      key: 'status',
      render: (text) => {
        const value = findByValue(STATUS_LIST, text);
        return <Tag color={value.badgeStatus}>{value.label}</Tag>;
      },
    },
  ];
  return serverColums;
};

export default function ServerTable({ servers }: { servers: API.Server[] }) {
  return (
    <Col span={24}>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.ServerTable.ServerList',
              defaultMessage: 'Server 列表',
            })}
          </h2>
        }
      >
        <Table
          columns={getServerColums()}
          rowKey="name"
          dataSource={servers}
          pagination={{ simple: true }}
          sticky
        />
      </Card>
    </Col>
  );
}
