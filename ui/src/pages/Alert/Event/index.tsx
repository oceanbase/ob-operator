import { alert } from '@/api';
import type {
  AlarmServerity,
  AlertAlert,
  AlertStatus,
  OceanbaseOBInstance,
} from '@/api/generated';
import { SERVERITY_MAP } from '@/constants';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';
import AlarmFilter from '../AlarmFilter';
const { Text } = Typography;



export default function Event() {
  const [form] = Form.useForm();
  const { data: listAlertsRes } = useRequest(alert.listAlerts);
  const listAlerts = listAlertsRes?.data || [];
  const columns: ColumnsType<AlertAlert> = [
    {
      title: '告警事件',
      dataIndex: 'summary',
      key: 'summary',
      render: (val) => <Button type="link">{val}</Button>,
    },
    {
      title: '告警对象',
      dataIndex: 'instance',
      key: 'instance',
      render: (instance: OceanbaseOBInstance) => (
        <Text>
          对象：{instance[instance.type]}
          <br />
          类型：{instance.type}
        </Text>
      ),
    },
    {
      title: '告警等级',
      dataIndex: 'serverity',
      key: 'serverity',
      render: (serverity: AlarmServerity) => (
        <Tag color={SERVERITY_MAP[serverity]?.color}>
          {SERVERITY_MAP[serverity]?.label}
        </Tag>
      ),
    },
    {
      title: '告警状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: AlertStatus) => <Tag>{status.state}</Tag>,
    },
    {
      title: '产生时间',
      dataIndex: 'startsAt',
      key: 'startsAt',
      defaultSortOrder: 'ascend',
      sorter: (pre: number, cur: number) => cur - pre,
      render: (startsAt: number) => (
        <Text>{moment.unix(startsAt).format('YYYY-MM-DD HH:MM:SS')}</Text>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: () => <Button type="link">屏蔽</Button>,
    },
  ];
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter form={form} type="event" />
      </Card>
      <Card
        title={<h2 style={{ marginBottom: 0 }}>事件列表</h2>}
      >
        <Table
          columns={columns}
          dataSource={listAlerts}
          rowKey="fingerprint"
          pagination={{ simple: true }}
          // scroll={{ x: 1500 }}
        />
      </Card>
    </Space>
  );
}
