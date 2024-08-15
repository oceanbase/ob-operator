import { alert } from '@/api';
import type { ReceiverReceiver } from '@/api/generated';
import PreText from '@/components/PreText';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import ChannelDrawer from './ChannelDrawer';

export default function Channel() {
  const [drawerStatus, setDrawerStatus] =
    useState<Alert.DrawerStatus>('create');
  const access = useAccess();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [clickedChannelName, setClickedChannelName] = useState<string>();
  const { data: listReceiversRes, refresh } = useRequest(alert.listReceivers);
  const { run: deleteReceiver } = useRequest(alert.deleteReceiver, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  const listReceivers = listReceiversRes?.data;
  const editChannel = (name: string) => {
    setClickedChannelName(name);
    setDrawerStatus('edit');
    setDrawerOpen(true);
  };
  const openChannel = (name: string) => {
    setClickedChannelName(name);
    setDrawerStatus('display');
    setDrawerOpen(true);
  };
  const columns: ColumnsType<ReceiverReceiver> = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Channel.6E368EE7',
        defaultMessage: '通道名',
      }),
      dataIndex: 'name',
      key: 'name',
      render: (_, record) => (
        <>
          {access.alarmwrite ? (
            <Button type="link" onClick={() => openChannel(record.name)}>
              {record.name}
            </Button>
          ) : (
            <span> {record.name}</span>
          )}
        </>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Channel.44150BAB',
        defaultMessage: '通道类型',
      }),
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Channel.02ABEE1E',
        defaultMessage: '通道配置',
      }),
      dataIndex: 'config',
      key: 'config',
      render: (value) => <PreText value={value} />,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Alert.Channel.75AFBE43',
        defaultMessage: '操作',
      }),
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button
            type="link"
            disabled={!access.alarmwrite}
            onClick={() => editChannel(record.name)}
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Channel.39A85374',
              defaultMessage: '编辑',
            })}
          </Button>
          <Button
            style={access.alarmwrite ? { color: '#ff4b4b' } : {}}
            disabled={!access.alarmwrite}
            onClick={() => {
              showDeleteConfirm({
                title: intl.formatMessage({
                  id: 'src.pages.Alert.Channel.32271D0F',
                  defaultMessage: '确定要删除“钉钉群”告警通道吗？',
                }),
                content: intl.formatMessage({
                  id: 'src.pages.Alert.Channel.0B830488',
                  defaultMessage:
                    '删除后使用该通道的消息推送都将失效，请谨慎操作',
                }),
                onOk: () => {
                  deleteReceiver(record.name);
                },
                okText: intl.formatMessage({
                  id: 'src.pages.Alert.Channel.152DCAE7',
                  defaultMessage: '删除',
                }),
              });
            }}
            type="link"
          >
            {intl.formatMessage({
              id: 'src.pages.Alert.Channel.93AFF36C',
              defaultMessage: '删除',
            })}
          </Button>
        </>
      ),
    },
  ];

  const createNewChannel = () => {
    setDrawerOpen(true);
    setDrawerStatus('create');
  };
  const drawerClose = () => {
    setClickedChannelName(undefined);
    setDrawerOpen(false);
  };

  return (
    <Card
      extra={
        access.alarmwrite ? (
          <Button type="primary" onClick={createNewChannel}>
            {intl.formatMessage({
              id: 'src.pages.Alert.Channel.DBCB373E',
              defaultMessage: '新建告警通道',
            })}
          </Button>
        ) : null
      }
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'src.pages.Alert.Channel.552AE97C',
            defaultMessage: '告警通道',
          })}
        </h2>
      }
    >
      <Table
        columns={columns}
        dataSource={listReceivers}
        rowKey="fingerprint"
        pagination={{ simple: true }}
      />

      <ChannelDrawer
        width={880}
        status={drawerStatus}
        setStatus={setDrawerStatus}
        name={clickedChannelName}
        submitCallback={refresh}
        onClose={drawerClose}
        open={drawerOpen}
      />
    </Card>
  );
}
