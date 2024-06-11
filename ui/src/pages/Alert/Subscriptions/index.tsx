import { alert } from '@/api';
import type { RouteRouteResponse } from '@/api/generated';
import PreText from '@/components/PreText';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { Alert } from '@/type/alert';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import ChannelDrawer from '../Channel/ChannelDrawer';
import { formatDuration } from '../helper';
import SubscripDrawerForm from './SubscripDrawerForm';

export default function Subscriptions() {
  const { data: listRoutesRes, refresh } = useRequest(alert.listRoutes);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [channelDrawerOpen, setChannelDrawerOpen] = useState(false);
  const [clickedId, setClickedId] = useState<string>();
  const [drawerStatus, setDrawerStatus] =
    useState<Alert.DrawerStatus>('display');
  const [editChannelName, setEditChannelName] = useState<string>();
  const { run: deleteRoute } = useRequest(alert.deleteRoute, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const drawerClose = () => {
    setClickedId(undefined);
    setDrawerOpen(false);
  };
  const channelDrawerClose = () => {
    setChannelDrawerOpen(false);
    setEditChannelName(undefined);
  };

  /**
   * If the id is undefined, it means creating
   */
  const editConfig = (id?: string) => {
    setClickedId(id);
    setDrawerOpen(true);
  };

  const showChannelDrawer = (receiver: string) => {
    setEditChannelName(receiver);
    setDrawerStatus('display');
    setChannelDrawerOpen(true);
  };

  const listRoutes = listRoutesRes?.data;

  const columns: ColumnsType<RouteRouteResponse> = [
    {
      title: '通道名',
      dataIndex: 'receiver',
      key: 'receiver',
      render: (receiver) => (
        <Button type="link" onClick={() => showChannelDrawer(receiver)}>
          {receiver}
        </Button>
      ),
    },
    {
      title: '匹配配置',
      dataIndex: 'matchers',
      key: 'matchers',
      render: (matchers) => {
        if (!matchers.length) return '-';
        return <PreText cols={7} value={matchers} />;
      },
    },
    {
      title: '聚合配置',
      dataIndex: 'aggregateLabels',
      key: 'aggregateLabels',
      render: (labels) => <span>{labels.join(',')}</span>,
    },
    {
      title: '推送周期',
      dataIndex: 'repeatInterval',
      render: (repeatIntervel) => <span>{formatDuration(repeatIntervel)}</span>,
    },
    {
      title: '操作',
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button
            onClick={() => editConfig(record.id)}
            style={{ paddingLeft: 0 }}
            type="link"
          >
            编辑
          </Button>
          <Button
            style={{ color: '#ff4b4b' }}
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
      extra={
        <Button type="primary" onClick={() => editConfig()}>
          新建推送
        </Button>
      }
      title={<h2 style={{ marginBottom: 0 }}>推送配置</h2>}
    >
      <Table
        columns={columns}
        dataSource={listRoutes}
        rowKey="id"
        pagination={{ simple: true }}
      />
      <SubscripDrawerForm
        title="推送配置"
        width={880}
        onClose={drawerClose}
        submitCallback={refresh}
        open={drawerOpen}
        id={clickedId}
      />
      <ChannelDrawer
        width={880}
        status={drawerStatus}
        setStatus={setDrawerStatus}
        name={editChannelName}
        onClose={channelDrawerClose}
        open={channelDrawerOpen}
      />
    </Card>
  );
}
