import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Card, Col, Progress, Table, Tag } from 'antd';
import type { ColumnsType, TableProps } from 'antd/es/table';

import CustomTooltip from '@/components/CustomTooltip';
import { getNodeInfoReq } from '@/services';

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

const columns: ColumnsType<DataType> = [
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.NodeName',
      defaultMessage: '节点名',
    }),
    dataIndex: 'name',
    key: 'name',
    width: 120,
    render: (val) => <CustomTooltip text={val} width={100} />,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (val) => {
      let color = 'default';
      if (val === 'ready') {
        color = 'success';
      } else if (val === 'not ready') {
        color = 'warning';
      }
      return <Tag color={color}>{val}</Tag>;
    },
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.Role',
      defaultMessage: '角色',
    }),
    dataIndex: 'roles',
    key: 'roles',
    width: 120,
    render: (val) => <CustomTooltip text={val} width={100} />,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.RunningTime',
      defaultMessage: '已运行时间',
    }),
    dataIndex: 'uptime',
    key: 'uptime',
    width: 120,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.Version',
      defaultMessage: '版本',
    }),
    dataIndex: 'version',
    key: 'version',
    width: 120,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.InternalIpAddress',
      defaultMessage: '内部IP',
    }),
    dataIndex: 'internalIP',
    key: 'internalIP',
    width: 120,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.ExternalIpAddress',
      defaultMessage: '外部IP',
    }),
    dataIndex: 'externalIP',
    key: 'externalIP',
    width: 120,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.OperatingSystem',
      defaultMessage: '操作系统',
    }),
    dataIndex: 'os',
    key: 'os',
    width: 140,
    render: (val) => <CustomTooltip text={val} width={100} />,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.KernelVersion',
      defaultMessage: '内核版本',
    }),
    dataIndex: 'kernel',
    key: 'kernel',
    width: 140,
    render: (val) => <CustomTooltip text={val} width={100} />,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.ContainerRuntime',
      defaultMessage: '容器运行时',
    }),
    dataIndex: 'cri',
    key: 'cri',
    width: 140,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.AllocatedCpu',
      defaultMessage: '已分配CPU',
    }),
    dataIndex: 'cpu',
    key: 'cpu',
    render: (value) => (
      <Progress status="normal" strokeLinecap="butt" percent={value} />
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Overview.NodesTable.AllocatedMemory',
      defaultMessage: '已分配内存',
    }),
    dataIndex: 'memory',
    key: 'memory',
    render: (value) => (
      <Progress status="normal" strokeLinecap="butt" percent={value} />
    ),
  },
];

export default function NodesTable() {
  const { data,loading } = useRequest(getNodeInfoReq);
  return (
    <Col span={24}>
      <Card loading={loading}>
        <h2>
          {intl.formatMessage({
            id: 'OBDashboard.pages.Overview.NodesTable.Node',
            defaultMessage: '节点',
          })}
        </h2>
        <Table
          columns={columns}
          dataSource={data as any}
          rowKey="name"
          pagination={{simple:true}}
          scroll={{ x: 1500 }}
          sticky
        />
      </Card>
    </Col>
  );
}
