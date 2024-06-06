import { alert } from '@/api';
import type {
  AlarmSeverity,
  AlertAlert,
  AlertStatus,
  OceanbaseOBInstance,
} from '@/api/generated';
import { ALERT_STATE_MAP, SEVERITY_MAP } from '@/constants';
import { history } from '@umijs/max';
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
import dayjs from 'dayjs'
import AlarmFilter from '../AlarmFilter';
import { sortEvents } from '../helper';
const { Text } = Typography;

export default function Event() {
  const [form] = Form.useForm();
  const { data: listAlertsRes, run: getListAlerts } = useRequest(
    alert.listAlerts,
  );
  const listAlerts = sortEvents(listAlertsRes?.data || []);
  const columns: ColumnsType<AlertAlert> = [
    {
      title: '告警事件',
      dataIndex: 'summary',
      key: 'summary',
      render: (val, record) => {
        return (
          <Button
            onClick={() => history.push(`/alert/rules?rule=${record.rule}`)}
            type="link"
          >
            <Tooltip title={record.description}>{val}</Tooltip>
          </Button>
        );
      },
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
      dataIndex: 'severity',
      key: 'severity',
      sorter: (preRecord, curRecord) => {
        return (
          SEVERITY_MAP[preRecord.severity].weight -
          SEVERITY_MAP[curRecord.severity].weight
        );
      },
      render: (severity: AlarmSeverity) => (
        <Tag color={SEVERITY_MAP[severity]?.color}>
          {SEVERITY_MAP[severity]?.label}
        </Tag>
      ),
    },
    {
      title: '告警状态',
      dataIndex: 'status',
      key: 'status',
      sorter: (preRecord, curRecord) => {
        return (
          ALERT_STATE_MAP[preRecord.status.state].weight -
          ALERT_STATE_MAP[curRecord.status.state].weight
        );
      },
      render: (status: AlertStatus) => (
        <Tag color={ALERT_STATE_MAP[status.state].color}>
          {ALERT_STATE_MAP[status.state].text || '-'}
        </Tag>
      ),
    },
    {
      title: '产生时间',
      dataIndex: 'startsAt',
      key: 'startsAt',
      sorter: (preRecord, curRecord) => curRecord.startsAt - preRecord.startsAt,
      render: (startsAt: number) => (
        <Text>{dayjs.unix(startsAt).format('YYYY-MM-DD HH:mm:ss')}</Text>
      ),
    },
    {
      title: '结束时间',
      dataIndex: 'endsAt',
      key: 'endsAt',
      sorter: (preRecord, curRecord) => curRecord.endsAt - preRecord.endsAt,
      render: (endsAt: number) => (
        <Text>{dayjs.unix(endsAt).format('YYYY-MM-DD HH:mm:ss')}</Text>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Button
          disabled={record.status.state !== 'active'}
          style={{ paddingLeft: 0 }}
          type="link"
          onClick={() => {
            history.push(
              `/alert/shield?instance=${JSON.stringify(
                record.instance,
              )}&label=${JSON.stringify(
                record.labels?.map((label) => ({
                  name: label.key,
                  value: label.value,
                })),
              )}&rule=${record.rule}`,
            );
          }}
        >
          屏蔽
        </Button>
      ),
    },
  ];
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListAlerts} form={form} type="event" />
      </Card>
      <Card title={<h2 style={{ marginBottom: 0 }}>事件列表</h2>}>
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
