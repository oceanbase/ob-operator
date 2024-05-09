import { alert } from '@/api';
import type { RouteRouteResponse } from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';

export default function Subscriptions() {
  const { data: listRoutesRes, refresh } = useRequest(alert.listRoutes);
  const { run: deleteRoute } = useRequest(alert.deleteRoute, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const listRoutes = listRoutesRes?.data;
  const columns: ColumnsType<RouteRouteResponse> = [
    {
      title: '通道名',
      dataIndex: 'receiver',
      key: 'receiver',
    },
    {
      title: '匹配配置',
      dataIndex: 'matchers',
      key: 'matchers',
    },
    {
      title: '聚合配置',
      dataIndex: 'aggregateLabels',
      key: 'aggregateLabels',
    },
    {
      title: '推送周期',
      dataIndex: 'repeatInterval',
    },
    {
      title: '操作',
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button type="link">编辑</Button>
          <Button
            style={{color:'#ff4b4b'}}
            onClick={() => {
              showDeleteConfirm({
                title: '确定要删除推送配置吗？',
                content: '删除后不可恢复，请谨慎操作',
                onOk: () => {
                  deleteRoute(record.id);
                },
                okText: '删除',
              });
            }}
            type="link"
          >
            删除
          </Button>
        </>
      ),
    },
  ];
  return (
    <Card
      extra={<Button type="primary">新建推送</Button>}
      title={<h2 style={{ marginBottom: 0 }}>推送配置</h2>}
    >
      <Table
        columns={columns}
        dataSource={listRoutes}
        rowKey="fingerprint"
        pagination={{ simple: true }}
      />
    </Card>
  );
}
