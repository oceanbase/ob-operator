import { alert } from '@/api';
import type {
  AlarmMatcher,
  OceanbaseOBInstance,
  SilenceSilencerResponse,
  SilenceStatus,
} from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { useSearchParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import ShieldDrawerForm from './ShieldDrawerForm';
import { Alert } from '@/type/alert';
const { Text } = Typography;

export default function Shield() {
  const [form] = Form.useForm();
  const [searchParams] = useSearchParams();

  const [drawerOpen, setDrawerOpen] = useState(
    Boolean(searchParams.get('instance')),
  );
  const { data: listSilencersRes, refresh } = useRequest(alert.listSilencers);
  const { run: deleteSilencer } = useRequest(alert.deleteSilencer, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const listSilencers = listSilencersRes?.data || [];
  const drawerClose = () => {
    setDrawerOpen(false);
  };
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
      render: (rules) => (
        <p>
          {rules
            .map((rule: AlarmMatcher) => rule.name! + rule.value!)
            .join(',')}
        </p>
      ),
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
      render: (_, record) => (
        <>
          <Button style={{ paddingLeft: 0 }} type="link">
            编辑
          </Button>
          <Button
            type="link"
            style={{ color: '#ff4b4b' }}
            onClick={() => {
              showDeleteConfirm({
                title: '确定删除该告警屏蔽条件吗？',
                content: '删除后不可恢复，请谨慎操作',
                okText: '删除',
                onOk: () => {
                  deleteSilencer(record.id);
                },
              });
            }}
          >
            删除
          </Button>
        </>
      ),
    },
  ];
  const initialValues: Alert.ShieldDrawerInitialValues = {};
  if (searchParams.get('instance')) {
    initialValues.instance = JSON.parse(searchParams.get('instance')!);
  }
  if (searchParams.get('label')) {
    initialValues.matchers = JSON.parse(searchParams.get('label')!);
  }
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
      <ShieldDrawerForm
        width={880}
        initialValues={initialValues}
        onClose={drawerClose}
        open={drawerOpen}
      />
    </Space>
  );
}
