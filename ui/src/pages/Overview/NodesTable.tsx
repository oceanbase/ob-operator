import CustomTooltip from '@/components/CustomTooltip';
import { NODESTABLE_STATUS_LIST } from '@/constants';
import { getNodeInfoReq } from '@/services';
import { getColumnSearchProps } from '@/utils/component';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Button, Card, Col, Progress, Space, Table, Tag, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import BatchEditNodeDrawer from './BatchEditNodeDrawer';
import EditNodeDrawer from './EditNodeDrawer';

interface DataType {
  key: React.Key;
  name: string;
  status: string;
  roles: string;
  uptime: string;
  version: string;
  internalIp: string;
  externalIp: string;
  os: string;
  kernel: string;
  cri: string;
  cup: string;
  memory: string;
}

const progressContent = (value: number, resource: number) => {
  return resource === 0 ? (
    <Tooltip
      title={intl.formatMessage({
        id: 'src.pages.Overview.E15A1FED',
        defaultMessage:
          'K8s 集群中尚未安装 metrics-server，无法获取节点资源用量',
      })}
    >
      - / -
    </Tooltip>
  ) : (
    <Progress status="normal" strokeLinecap="butt" percent={value} />
  );
};

export default function NodesTable({
  type,
  nodeLoading,
  onSuccess,
  k8sClusterName,
  K8sClustersNodeList,
}) {
  const { data, loading, refresh } = useRequest(getNodeInfoReq, {
    ready: type !== 'k8s',
  });

  const [isDrawerOpen, setIsDrawerOpen] = useState<boolean>(false);
  const [nodeRecord, setNodeRecord] = useState({});
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [batchNodeDrawerOpen, setBatchNodeDrawerOpen] =
    useState<boolean>(false);

  const columns: ColumnsType<DataType> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'name',
      }),
      render: (val) => <CustomTooltip text={val} width={100} />,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.Role',
        defaultMessage: '角色',
      }),
      dataIndex: 'roles',
      key: 'roles',
      width: 120,
      render: (val) => {
        return val?.length !== 0 ? (
          <CustomTooltip text={val} width={100} />
        ) : (
          '-'
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.Status',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (text) => {
        const value = findByValue(NODESTABLE_STATUS_LIST, text);
        return <Tag color={value.badgeStatus}>{value.label}</Tag>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.InternalIpAddress',
        defaultMessage: 'IP',
      }),
      dataIndex: 'internalIP',
      key: 'internalIP',
      width: 120,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'internalIP',
      }),
    },

    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.AllocatedCpu',
        defaultMessage: '已分配CPU',
      }),
      dataIndex: 'cpu',
      key: 'cpu',
      render: (value, record) => progressContent(value, record.cpuTotal),
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.AllocatedMemory',
        defaultMessage: '已分配内存',
      }),
      dataIndex: 'memory',
      key: 'memory',
      render: (value, record) => progressContent(value, record.memoryTotal),
    },
    {
      title: 'labels',
      dataIndex: 'labels',
      ellipsis: true,
      width: 160,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'labels',
        arraySearch: true,
        symbol: '=',
      }),
      render: (text) => {
        const tooltipTitle = text?.map((item) => (
          <div>{`${item.key}=${item.value}`}</div>
        ));

        const content = text?.map((item) => `${item.key}=${item.value}`);
        return content?.length === 0 ? (
          '-'
        ) : (
          <CustomTooltip
            text={content}
            tooltipTitle={tooltipTitle}
            width={150}
          />
        );
      },
    },
    {
      title: 'taints',
      dataIndex: 'taints',
      width: 160,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'taints',
        arraySearch: true,
        symbol: '=',
      }),
      render: (text) => {
        const content = text?.map((item) =>
          item.value
            ? `${item.key}=${item.value}:${item.effect}`
            : `${item.key}:${item.effect}`,
        );
        const tooltipTitle = text?.map((item) =>
          item.value ? (
            <div>{`${item.key}=${item.value}:${item.effect}`}</div>
          ) : (
            <div>{`${item.key}:${item.effect}`}</div>
          ),
        );

        return content?.length === 0 ? (
          '-'
        ) : (
          <CustomTooltip
            text={content}
            tooltipTitle={tooltipTitle}
            width={150}
          />
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.pages.Overview.NodesTable.RunningTime',
        defaultMessage: '启动时间',
      }),
      dataIndex: 'uptime',
      key: 'uptime',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.1B9EA477',
        defaultMessage: '操作',
      }),
      align: 'center',
      render: (text, record) => {
        return (
          <Space size={1}>
            <Button
              type="link"
              onClick={() => {
                setIsDrawerOpen(true);
                setNodeRecord(record);
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.F5A088FB',
                defaultMessage: '编辑',
              })}
            </Button>
          </Space>
        );
      },
    },
  ];

  const rowSelection: TableRowSelection<DataType> = {
    onChange: (_, record) => {
      setSelectedRowKeys(record);
    },
  };

  return (
    <Col span={24}>
      <Card
        loading={loading || nodeLoading}
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'OBDashboard.pages.Overview.NodesTable.Node',
              defaultMessage: '节点',
            })}
          </h2>
        }
        extra={
          selectedRowKeys.length > 0 && (
            <Button onClick={() => setBatchNodeDrawerOpen(true)} type="primary">
              批量编辑
            </Button>
          )
        }
      >
        <Table
          rowSelection={rowSelection}
          columns={columns}
          dataSource={type === 'k8s' ? K8sClustersNodeList : data}
          rowKey="name"
          pagination={{ simple: true }}
          scroll={{ x: 1500 }}
          sticky
        />
      </Card>

      <BatchEditNodeDrawer
        type={type}
        k8sClusterName={k8sClusterName}
        selectedRowKeys={selectedRowKeys}
        visible={batchNodeDrawerOpen}
        onCancel={() => {
          setBatchNodeDrawerOpen(false);
        }}
        onSuccess={() => {
          if (type === 'k8s') {
            onSuccess();
          } else {
            refresh();
          }
          setBatchNodeDrawerOpen(false);
          setSelectedRowKeys([]);
        }}
      />
      <EditNodeDrawer
        type={type}
        k8sClusterName={k8sClusterName}
        visible={isDrawerOpen}
        onCancel={() => {
          setIsDrawerOpen(false);
          setNodeRecord({});
        }}
        onSuccess={() => {
          setIsDrawerOpen(false);
          setNodeRecord({});
          if (type === 'k8s') {
            onSuccess();
          } else {
            refresh();
          }
        }}
        nodeRecord={nodeRecord}
      />
    </Col>
  );
}
