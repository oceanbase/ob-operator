import { alert } from '@/api';
import type {
  OceanbaseOBInstance,
  SilenceSilencerResponse,
  SilenceStatus,
} from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { SHILED_STATUS_MAP } from '@/constants';
import { Alert } from '@/type/alert';
import { useSearchParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import {
  Button,
  Card,
  Form,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { clone } from 'lodash';
import moment from 'moment';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import { sortAlarmShielding } from '../helper';
import ShieldDrawerForm from './ShieldDrawerForm';
const { Text } = Typography;

type InstancesRender = {
  type?: Alert.InstancesKey;
  observer?: string[];
  obtenant?: string[];
  obcluster?: string[];
  obzone?: string[];
};

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
      width: 200,
      render: (instances: OceanbaseOBInstance[]) => {
        const temp: InstancesRender = {};
        for (const instance of instances) {
          Object.keys(instance).forEach((key: keyof OceanbaseOBInstance) => {
            if (temp[key]) {
              temp[key] = [...temp[key], instance[key]];
            } else {
              temp[key] = [instance[key]];
            }
          });
        }
        delete temp.type;

        const InstancesRender = () => (
          <div>
            {Object.keys(temp).map((key,index) => (
              <p key={index}>
                {key}：{temp[key].join(',')}
              </p>
            ))}
          </div>
        );
        return (
          <Tooltip title={<InstancesRender />}>
            <div>
              {Object.keys(temp).map((key) => (
                <Text ellipsis style={{ width: 200 }}>
                  {key}：{temp[key].join(',')}
                </Text>
              ))}
            </div>
          </Tooltip>
        );
      },
    },
    {
      title: '屏蔽告警规则',
      dataIndex: 'matchers',
      key: 'matchers',
      render: (rules) => {
        const newRules = clone(rules);
        if (newRules.length)
          newRules.splice(0, 0, { name: '规则名', value: '规则' });
        return (
          <Space style={{ width: '100%' }} direction="vertical">
            {newRules?.map((rule) => {
              return (
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div style={{ flex: 1 }}>{rule.name}</div>
                  <div style={{ flex: 1 }}> {rule.value}</div>
                </div>
              );
            })}
          </Space>
        );
      },
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
            style={
              record.status.state !== 'expired' ? { color: '#ff4b4b' } : {}
            }
            disabled={record.status.state === 'expired'}
            onClick={() => {
              showDeleteConfirm({
                title: '确定解除该告警屏蔽条件吗？',
                content: '解除后不可恢复，请谨慎操作',
                okText: '解除',
                onOk: () => {
                  deleteSilencer(record.id);
                },
              });
            }}
          >
            解除屏蔽
          </Button>
        </>
      ),
    },
  ];
  const formatInstanceParam = (instanceParam: Alert.InstanceParamType) => {
    const { obcluster, observer, obtenant, type } = instanceParam;
    const res: Alert.InstancesType = {
      type,
      obcluster: [obcluster!],
    };
    if (observer) res.observer = [observer];
    if (obtenant) res.obtenant = [obtenant];
    return res;
  };
  const initialValues: Alert.ShieldDrawerInitialValues = {};
  if (searchParams.get('instance')) {
    initialValues.instances = formatInstanceParam(
      JSON.parse(searchParams.get('instance')!),
    );
  }
  if (searchParams.get('label')) {
    initialValues.matchers = JSON.parse(searchParams.get('label')!);
  }
  if (searchParams.get('rule')) {
    initialValues.rules = searchParams.get('rule')
      ? [searchParams.get('rule')!]
      : undefined;
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
