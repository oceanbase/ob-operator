import { alert } from '@/api';
import type { AlarmSeverity, RuleRuleResponse } from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { SEVERITY_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import { formatDuration } from '../helper';
import RuleDrawerForm from './RuleDrawerForm';

const { Text } = Typography;

export default function Rules() {
  const [form] = Form.useForm();
  const {
    data: listRulesRes,
    refresh,
    run: getListRules,
  } = useRequest(alert.listRules);
  const [editRuleName, setEditRuleName] = useState<string>();
  const { run: deleteRule } = useRequest(alert.deleteRule, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const listRules =
    listRulesRes?.data.sort(
      (pre, cur) =>
        SEVERITY_MAP[cur.severity].weight - SEVERITY_MAP[pre.severity].weight,
    ) || [];
  const editRule = (ruleName: string) => {
    setEditRuleName(ruleName);
    setDrawerOpen(true);
  };
  const drawerClose = () => {
    setEditRuleName(undefined);
    setDrawerOpen(false);
  };

  const columns: ColumnsType<RuleRuleResponse> = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.77E702BB',
        defaultMessage: '告警规则名',
      }),
      dataIndex: 'name',
      fixed: true,
      key: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.C6233D40',
        defaultMessage: '触发规则',
      }),
      dataIndex: 'query',
      width: '30%',
      key: 'query',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.188CEA51',
        defaultMessage: '持续时间',
      }),
      dataIndex: 'duration',
      key: 'duration',
      render: (value) => <Text>{formatDuration(value)}</Text>,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.C8C937C1',
        defaultMessage: '对象类型',
      }),
      dataIndex: 'instanceType',
      key: 'instanceType',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.18FF1D51',
        defaultMessage: '告警等级',
      }),
      dataIndex: 'severity',
      key: 'severity',
      sorter: (preRecord, curRecord) =>
        SEVERITY_MAP[curRecord.severity].weight -
        SEVERITY_MAP[preRecord.severity].weight,
      render: (severity: AlarmSeverity) => (
        <Tag color={SEVERITY_MAP[severity]?.color}>
          {SEVERITY_MAP[severity]?.label}
        </Tag>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.AEB0287D',
        defaultMessage: '类型',
      }),
      dataIndex: 'type',
      key: 'type',
      filters: [
        {
          text: intl.formatMessage({
            id: 'src.pages.Alert.Rules.140E8C1E',
            defaultMessage: '自定义',
          }),
          value: 'customized',
        },
        {
          text: intl.formatMessage({
            id: 'src.pages.Alert.Rules.7789B7EF',
            defaultMessage: '默认',
          }),
          value: 'builtin',
        },
      ],

      onFilter: (value, record) => record.type === value,
      render: (type) => (
        <Text>
          {type === 'builtin'
            ? intl.formatMessage({
                id: 'src.pages.Alert.Rules.9B28C134',
                defaultMessage: '默认',
              })
            : intl.formatMessage({
                id: 'src.pages.Alert.Rules.224DA83F',
                defaultMessage: '自定义',
              })}
        </Text>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Rules.F333E1DF',
        defaultMessage: '操作',
      }),
      fixed: 'right',
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button
            onClick={() => editRule(record.name)}
            style={{ paddingLeft: 0 }}
            type="link"
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Rules.873B1514',
              defaultMessage: '编辑',
            })}
          </Button>
          <Button
            type="link"
            style={{ color: '#ff4b4b' }}
            onClick={() => {
              showDeleteConfirm({
                title: intl.formatMessage(
                  {
                    id: 'src.pages.Alert.Rules.3FB7D236',
                    defaultMessage: '确定要删除 ${record.name} 告警规则吗？',
                  },
                  { recordName: record.name },
                ),
                content: intl.formatMessage({
                  id: 'src.pages.Alert.Rules.DA9CF3DF',
                  defaultMessage:
                    '删除后，引用该告警规则的规则分组与告警模版将同步删除关于此告警规则的配置，请谨慎操作。',
                }),

                okText: intl.formatMessage({
                  id: 'src.pages.Alert.Rules.EF19C6D7',
                  defaultMessage: '删除',
                }),
                onOk: () => {
                  deleteRule(record.name);
                },
              });
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Rules.D7681DA5',
              defaultMessage: '删除',
            })}
          </Button>
        </>
      ),
    },
  ];

  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListRules} form={form} type="rules" />
      </Card>
      <Card
        extra={
          <Button onClick={() => setDrawerOpen(true)} type="primary">
            {intl.formatMessage({
              id: 'src.pages.Alert.Rules.90D4952A',
              defaultMessage: '新建告警规则',
            })}
          </Button>
        }
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'src.pages.Alert.Rules.B943644E',
              defaultMessage: '规则列表',
            })}
          </h2>
        }
      >
        <Table
          columns={columns}
          rowKey="name"
          dataSource={listRules}
          pagination={{ simple: true }}
          scroll={{ x: 1400 }}
        />
      </Card>
      <RuleDrawerForm
        width={880}
        open={drawerOpen}
        ruleName={editRuleName}
        onClose={drawerClose}
        submitCallback={refresh}
      />
    </Space>
  );
}
