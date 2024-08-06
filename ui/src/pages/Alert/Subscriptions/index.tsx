import { alert } from '@/api';
import type { RouteRouteResponse } from '@/api/generated';
import PreText from '@/components/PreText';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import ChannelDrawer from '../Channel/ChannelDrawer';
import { formatDuration } from '../helper';
import SubscripDrawerForm from './SubscripDrawerForm';

export default function Subscriptions() {
  const { data: listRoutesRes, refresh } = useRequest(alert.listRoutes);
  const access = useAccess();
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
      title: intl.formatMessage({
        id: 'src.pages.Alert.Subscriptions.3988FFB5',
        defaultMessage: '通道名',
      }),
      dataIndex: 'receiver',
      key: 'receiver',
      render: (receiver) => (
        <>
          {access.alarmwrite ? (
            <Button type="link" onClick={() => showChannelDrawer(receiver)}>
              {receiver}
            </Button>
          ) : (
            <span>{receiver}</span>
          )}
        </>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Subscriptions.B477E130',
        defaultMessage: '匹配配置',
      }),
      dataIndex: 'matchers',
      key: 'matchers',
      render: (matchers) => {
        if (!matchers.length) return '-';
        return <PreText cols={7} value={matchers} />;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Subscriptions.D2F05048',
        defaultMessage: '聚合配置',
      }),
      dataIndex: 'aggregateLabels',
      key: 'aggregateLabels',
      render: (labels) => <span>{labels.join(',')}</span>,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Subscriptions.19CF4790',
        defaultMessage: '推送周期',
      }),
      dataIndex: 'repeatInterval',
      render: (repeatIntervel) => <span>{formatDuration(repeatIntervel)}</span>,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Subscriptions.2BDE4581',
        defaultMessage: '操作',
      }),
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button
            onClick={() => editConfig(record.id)}
            disabled={!access.alarmwrite}
            style={{ paddingLeft: 0 }}
            type="link"
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Subscriptions.1BA1F775',
              defaultMessage: '编辑',
            })}
          </Button>
          <Button
            style={{ color: '#ff4b4b' }}
            disabled={!access.alarmwrite}
            onClick={() => {
              showDeleteConfirm({
                title: intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.6FD5AF22',
                  defaultMessage: '确定要删除推送配置吗？',
                }),
                content: intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.44ED2092',
                  defaultMessage: '删除后不可恢复，请谨慎操作',
                }),
                onOk: () => {
                  deleteRoute(record.id);
                },
                okText: intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.73CAD042',
                  defaultMessage: '删除',
                }),
              });
            }}
            type="link"
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Subscriptions.4800F4C6',
              defaultMessage: '删除',
            })}
          </Button>
        </>
      ),
    },
  ];

  return (
    <Card
      extra={
        access.alarmwrite ? (
          <Button type="primary" onClick={() => editConfig()}>
            {intl.formatMessage({
              id: 'src.pages.Alert.Subscriptions.DB2B8DA0',
              defaultMessage: '新建推送',
            })}
          </Button>
        ) : null
      }
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'src.pages.Alert.Subscriptions.3DB73ECC',
            defaultMessage: '推送配置',
          })}
        </h2>
      }
    >
      <Table
        columns={columns}
        dataSource={listRoutes}
        rowKey="id"
        pagination={{ simple: true }}
      />

      <SubscripDrawerForm
        title={intl.formatMessage({
          id: 'src.pages.Alert.Subscriptions.73FA804B',
          defaultMessage: '推送配置',
        })}
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
