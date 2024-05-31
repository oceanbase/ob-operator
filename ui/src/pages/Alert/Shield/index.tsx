import { alert } from '@/api';
import type {
  OceanbaseOBInstance,
  SilenceSilencerResponse,
  SilenceStatus,
} from '@/api/generated';
import PreText from '@/components/PreText';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { SHILED_STATUS_MAP } from '@/constants';
import { Alert } from '@/type/alert';
import { useSearchParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import { sortAlarmShielding } from '../helper';
import ShieldDrawerForm from './ShieldDrawerForm';
const { Text } = Typography;

export default function Shield() {
  const [form] = Form.useForm();
  const [searchParams, setSearchParams] = useSearchParams();
  const [editShieldId, setEditShieldId] = useState<string>();
  const [drawerOpen, setDrawerOpen] = useState(
    Boolean(searchParams.get('instance')),
  );
  const {
    data: listSilencersRes,
    refresh,
    run: getListSilencers,
  } = useRequest(alert.listSilencers);
  const { run: deleteSilencer } = useRequest(alert.deleteSilencer, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const listSilencers = sortAlarmShielding(listSilencersRes?.data || []);
  const drawerClose = () => {
    setSearchParams('');
    setEditShieldId(undefined);
    setDrawerOpen(false);
  };
  const editShield = (id: string) => {
    setEditShieldId(id);
    setDrawerOpen(true);
  };
  const columns: ColumnsType<SilenceSilencerResponse> = [
    {
      title: '屏蔽应用/对象类型',
      dataIndex: 'instances',
      key: 'type',
      render: (instances: OceanbaseOBInstance[]) => (
        <Text>{instances[0].type || '-'}</Text>
      ),
    },
    {
      title: '屏蔽对象',
      dataIndex: 'instances',
      key: 'instances',
      width:300,
      render: (instances: OceanbaseOBInstance[]) => (
        <Text>
          {instances.map((instance) => {
            delete instance.type;
            return <pre>{JSON.stringify(instance, null,2)}</pre>;
          }) || '-'}
        </Text>
      ),
    },
    {
      title: '屏蔽告警规则',
      dataIndex: 'matchers',
      key: 'matchers',
      width: 400,
      render: (rules) => <PreText value={rules} cols={7} />,
    },
    {
      title: '屏蔽结束时间',
      dataIndex: 'endsAt',
      key: 'endsAt',
      sorter: (preRecord, curRecord) => curRecord.startsAt - preRecord.startsAt,
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
      sorter: (preRecord, curRecord) =>
        SHILED_STATUS_MAP[curRecord.status.state].weight -
        SHILED_STATUS_MAP[preRecord.status.state].weight,
      render: (status: SilenceStatus) => (
        <Tag color={SHILED_STATUS_MAP[status.state].color}>
          {SHILED_STATUS_MAP[status.state]?.text || '-'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'startsAt',
      key: 'startsAt',
      sorter: (preRecord, curRecord) => curRecord.startsAt - preRecord.startsAt,
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
          <Button
            onClick={() => editShield(record.id)}
            style={{ paddingLeft: 0 }}
            type="link"
          >
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
  if (searchParams.get('instances')) {
    initialValues.instances = JSON.parse(searchParams.get('instances')!);
  }
  if (searchParams.get('label')) {
    initialValues.matchers = JSON.parse(searchParams.get('label')!);
  }
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListSilencers} form={form} type="shield" />
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
        submitCallback={refresh}
        open={drawerOpen}
        id={editShieldId}
      />
    </Space>
  );
}
