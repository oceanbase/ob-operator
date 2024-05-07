import { alert } from '@/api';
import type {
  OceanbaseOBInstance,
  SilenceSilencerResponse,
  SilenceStatus,
} from '@/api/generated';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import ShieldDrawerForm from './ShieldDrawerForm';
const { Text } = Typography;

const columns: ColumnsType<SilenceSilencerResponse> = [
  {
    title: '屏蔽应用/对象类型',
    dataIndex: 'instance',
    key: 'type',
    render: (instance: OceanbaseOBInstance) => <Text>{instance.type}</Text>,
  },
  {
    title: '屏蔽对象',
    dataIndex: 'instance',
    key: 'targetObj',
    render: (instance: OceanbaseOBInstance) => (
      <Text>{instance[instance.type]}</Text>
    ),
  },
  {
    title: '屏蔽告警规则',
    dataIndex: 'matchers',
    key: 'matchers',
  },
  {
    title: '屏蔽结束时间',
    dataIndex: 'endsAt',
    key: 'endsAt',
    render: (endsAt) => (
      <Text>{moment.unix(endsAt).format('YYYY-MM-DD HH:MM:SS')}</Text>
    ),
  },
  {
    title: '创建人',
    dataIndex: 'createdBy',
    key: 'createdBy',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: SilenceStatus) => <Text>{status.state}</Text>,
  },
  {
    title: '创建时间',
    dataIndex: 'startsAt',
    key: 'startsAt',
    render: (startsAt) => (
      <Text>{moment.unix(startsAt).format('YYYY-MM-DD HH:MM:SS')}</Text>
    ),
  },
  {
    title: '备注',
    dataIndex: 'comment',
    key: 'comment',
  },
  {
    title: '操作',
    key: 'action',
    render: () => (
      <>
        {' '}
        <Button type="link">编辑</Button>
        <Button type="link">删除</Button>{' '}
      </>
    ),
  },
];

export default function Shield() {
  const [form] = Form.useForm();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { data: listSilencersRes } = useRequest(alert.listSilencers);
  const listSilencers = listSilencersRes?.data || [];
  const drawerClose = () => {
    setDrawerOpen(false);
  };
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter form={form} type="shield" />
      </Card>
      <Card
        title={<h2 style={{ marginBottom: 0 }}>屏蔽列表</h2>}
        extra={
          <Button type="primary" onClick={() => setDrawerOpen(true)}>
            新建屏蔽
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={listSilencers}
          rowKey="id"
          pagination={{ simple: true }}
          // scroll={{ x: 1500 }}
        />
      </Card>
      <ShieldDrawerForm width={880} onClose={drawerClose} open={drawerOpen} />
    </Space>
  );
}
