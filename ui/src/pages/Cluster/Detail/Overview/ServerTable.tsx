import { obcluster } from '@/api';
import { STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Button, Card, Col, message, Modal, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';

export default function ServerTable({
  clusterDetail,
  clusterDetailRefresh,
}: {
  clusterDetail: API.ClusterDetail[];
  clusterDetailRefresh: () => void;
}) {
  const { info, servers, supportStaticIP, status } = clusterDetail || {};

  const { namespace, name } = info;
  const { runAsync: deleteOBServers } = useRequest(obcluster.deleteOBServers, {
    manual: true,
    onSuccess: (res) => {
      if (res.successful) {
        message.success(
          intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.D01DC5A6',
            defaultMessage: '删除 Server 已成功',
          }),
        );
        clusterDetailRefresh();
      }
    },
  });
  const { runAsync: restartOBServers, loading: restartOBServersLoading } =
    useRequest(obcluster.restartOBServers, {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.2A1DB242',
              defaultMessage: '重启 Server 已成功',
            }),
          );
          clusterDetailRefresh();
        }
      },
    });

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
        id: 'src.pages.Cluster.Detail.Overview.AD557262',
        defaultMessage: 'observer 地址',
      }),
      dataIndex: 'address',
      key: 'address',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.13866D9B',
        defaultMessage: 'K8s 集群',
      }),
      dataIndex: 'k8sCluster',
      render: (text) => {
        return <span>{text || '-'}</span>;
      },
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
      title: 'K8s Node',
      dataIndex: 'nodeName',
    },
    {
      title: 'K8s Node IP',
      dataIndex: 'nodeIp',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.64ED384A',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (_, record) => {
        // 任何 zone 里面只剩一个 server 就不能删了
        const sameZone = servers?.filter(
          (server) => server?.zone === record?.zone,
        );
        return (
          <>
            <Button
              type="link"
              style={{ paddingLeft: 0 }}
              disabled={
                restartOBServersLoading ||
                !supportStaticIP ||
                status !== 'running'
              }
              onClick={() => {
                Modal.confirm({
                  title: intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.EFF8446F',
                    defaultMessage: '确定要重启当前 server 吗?',
                  }),
                  onOk: () => {
                    restartOBServers(namespace, name, {
                      observers: [record?.name],
                    });
                  },
                });
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.E0965DEE',
                defaultMessage: '重启',
              })}
            </Button>
            <Button
              danger
              type="link"
              disabled={
                servers.length === 1 ||
                sameZone.length === 1 ||
                status !== 'running'
              }
              onClick={() => {
                Modal.confirm({
                  title: intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.1FE7273F',
                    defaultMessage: '确定要删除当前 server 吗?',
                  }),
                  okType: 'danger',
                  onOk: () => {
                    deleteOBServers(namespace, name, {
                      observers: [record?.name],
                    });
                  },
                });
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.9DCE1679',
                defaultMessage: '删除',
              })}
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
