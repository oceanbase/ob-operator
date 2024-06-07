import { alert } from '@/api';
import type { AlarmSeverity, RuleRuleResponse } from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { SEVERITY_MAP } from '@/constants';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
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
  const listRules = listRulesRes?.data || [];

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
      title: '告警规则名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '触发规则',
      dataIndex: 'query',
      width: '30%',
      key: 'query',
    },
    {
      title: '持续时间',
      dataIndex: 'duration',
      key: 'duration',
      render: (value) => <Text>{value}分钟</Text>,
    },
    {
      title: '对象类型',
      dataIndex: 'instanceType',
      key: 'instanceType',
    },
    {
      title: '告警等级',
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
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      filters: [
        {
          text: '自定义',
          value: 'customized',
        },
        {
          text: '默认',
          value: 'builtin',
        },
      ],
      onFilter: (value, record) => record.type === value,
      render: (type) => <Text>{type === 'builtin' ? '默认' : '自定义'}</Text>,
    },
    {
      title: '操作',
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button
            onClick={() => editRule(record.name)}
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
                title: `确定要删除 ${record.name} 告警规则吗？`,
                content:
                  '删除后，引用该告警规则的规则分组与告警模版将同步删除关于此告警规则的配置，请谨慎操作。',
                okText: '删除',
                onOk: () => {
                  deleteRule(record.name);
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

  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListRules} form={form} type="rules" />
      </Card>
      <Card
        extra={
          <Button onClick={() => setDrawerOpen(true)} type="primary">
            新建告警规则
          </Button>
        }
        title={<h2 style={{ marginBottom: 0 }}>规则列表</h2>}
      >
        <Table
          columns={columns}
          rowKey="name"
          dataSource={listRules}
          pagination={{ simple: true }}
          // scroll={{ x: 1500 }}
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
