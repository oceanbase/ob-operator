import { STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { Button, Card, Col, Modal, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { flatten } from 'lodash';

export default function ServerTable({
  clusterDetail,
}: {
  clusterDetail: API.ClusterDetail[];
}) {
  const servers = clusterDetail?.servers;
  // 是否为单节点集群： zone的个数 ===1 && OBServer的个数 == 1
  const isStandAloneCluster =
    clusterDetail.zones?.length == 1 &&
    flatten((clusterDetail.zones || []).map((item) => item.servers || []))
      ?.length === 1;

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
    {
      title: '操作',
      dataIndex: 'operation',
      render: () => {
        return (
          <>
            <Button
              type="link"
              style={{ paddingLeft: 0 }}
              // TODO: 重启说明，后端会给参数对照
              // disabled={}
              onClick={() => {
                Modal.confirm({
                  title: '确定要重启当前 server 吗?',
                  onOk: () => {},
                });
              }}
            >
              重启
            </Button>
            <Button
              danger
              type="link"
              disabled={isStandAloneCluster}
              onClick={() => {
                Modal.confirm({
                  title: '确定要删除当前 server 吗?',
                  okType: 'danger',
                  onOk: () => {},
                });
              }}
            >
              删除
            </Button>
          </>
        );
      },
    },
  ];

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
          columns={serverColums}
          rowKey="name"
          dataSource={servers}
          pagination={{ simple: true }}
          sticky
        />
      </Card>
    </Col>
  );
}
