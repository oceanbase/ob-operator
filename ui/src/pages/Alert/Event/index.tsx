import { alert } from '@/api';
import type {
  AlarmSeverity,
  AlertAlert,
  AlertStatus,
  OceanbaseOBInstance,
} from '@/api/generated';
import { ALERT_STATE_MAP, SEVERITY_MAP } from '@/constants';
import { intl } from '@/utils/intl';
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
import dayjs from 'dayjs';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import RuleDrawerForm from '../Rules/RuleDrawerForm';
import { sortEvents } from '../helper';
const { Text } = Typography;

export default function Event() {
  const [form] = Form.useForm();
  const [drawerOpen, setDrawerOpen] = useState<boolean>(false);
  const [editRuleName, setEditRuleName] = useState<string>();
  const { data: listAlertsRes, run: getListAlerts } = useRequest(
    alert.listAlerts,
  );
  const editRule = (rule: string) => {
    setEditRuleName(rule);
    setDrawerOpen(true);
  };

  const listAlerts = sortEvents(listAlertsRes?.data || []);
  const columns: ColumnsType<AlertAlert> = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.19D28466',
        defaultMessage: '告警事件',
      }),
      dataIndex: 'summary',
      key: 'summary',
      width:'25%',
      render: (val, record) => {
        return (
          <Button onClick={() => editRule(record.rule)} type="link">
            <Tooltip title={record.description}>
              <div
                style={{
                  whiteSpace: 'break-spaces',
                  textAlign: 'left',
                }}
              >
                {val}
              </div>
            </Tooltip>
          </Button>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.EDED9514',
        defaultMessage: '告警对象',
      }),
      dataIndex: 'instance',
      key: 'instance',
      render: (instance: OceanbaseOBInstance) => (
        <Text>
          {intl.formatMessage({
            id: 'src.pages.Alert.Event.3EAC0543',
            defaultMessage: '对象：',
          })}
          {instance[instance.type]}
          <br />
          {intl.formatMessage({
            id: 'src.pages.Alert.Event.AB6EB56A',
            defaultMessage: '类型：',
          })}
          {instance.type}
        </Text>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.8BDBE511',
        defaultMessage: '告警等级',
      }),
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
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.26E65D10',
        defaultMessage: '告警状态',
      }),
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
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.90B9AC55',
        defaultMessage: '产生时间',
      }),
      dataIndex: 'startsAt',
      key: 'startsAt',
      sorter: (preRecord, curRecord) => curRecord.startsAt - preRecord.startsAt,
      render: (startsAt: number) => (
        <Text>{dayjs.unix(startsAt).format('YYYY-MM-DD HH:mm:ss')}</Text>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.929C9905',
        defaultMessage: '结束时间',
      }),
      dataIndex: 'endsAt',
      key: 'endsAt',
      sorter: (preRecord, curRecord) => curRecord.endsAt - preRecord.endsAt,
      render: (endsAt: number) => (
        <Text>{dayjs.unix(endsAt).format('YYYY-MM-DD HH:mm:ss')}</Text>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Event.BD73F875',
        defaultMessage: '操作',
      }),
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
          {intl.formatMessage({
            id: 'src.pages.Alert.Event.2BBFF587',
            defaultMessage: '屏蔽',
          })}
        </Button>
      ),
    },
  ];

  const drawerClose = () => {
    setEditRuleName(undefined);
    setDrawerOpen(false);
  };
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListAlerts} form={form} type="event" />
      </Card>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'src.pages.Alert.Event.0358EEE4',
              defaultMessage: '事件列表',
            })}
          </h2>
        }
      >
        <Table
          columns={columns}
          dataSource={listAlerts}
          rowKey="fingerprint"
          pagination={{ simple: true }}
          // scroll={{ x: 1500 }}
        />
      </Card>
      <RuleDrawerForm
        width={880}
        open={drawerOpen}
        ruleName={editRuleName}
        onClose={drawerClose}
      />
    </Space>
  );
}
