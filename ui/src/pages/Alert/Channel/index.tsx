import { alert } from '@/api';
import type { ReceiverReceiver } from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import ChannelDrawer from './ChannelDrawer';

export default function Channel() {
  const [drawerStatus, setDrawerStatus] = useState<Alert.DrawerStatus>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [clickedChannelName, setClickedChannelName] = useState('');
  const { data: listReceiversRes, refresh } = useRequest(alert.listReceivers);
  const { run: deleteReceiver } = useRequest(alert.deleteReceiver, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });
  let listReceivers = listReceiversRes?.data;
  listReceivers = [
    {
      config: 'string',
      name: 'string',
      type: 'dingtalk',
    },
  ];
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
      title: '通道名',
      dataIndex: 'name',
      key: 'name',
      render: (_, record) => (
        <Button type='link' onClick={() => openChannel(record.name)}>{record.name}</Button>
      ),
    },
    {
      title: '通道类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '通道配置',
      dataIndex: 'config',
      key: 'config',
    },
    {
      title: '操作',
      dataIndex: 'action',
      render: (_, record) => (
        <>
          <Button type="link" onClick={() => editChannel(record.name)}>
            编辑
          </Button>
          <Button
            style={{ color: '#ff4b4b' }}
            onClick={() => {
              showDeleteConfirm({
                title: '确定要删除“钉钉群”告警通道吗？',
                content: '删除后使用该通道的消息推送都将失效，请谨慎操作',
                onOk: () => {
                  deleteReceiver(record.name);
                },
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
  const createNewChannel = () => {
    setDrawerOpen(true);
    setDrawerStatus('create');
  };

  return (
    <Card
      extra={
        <Button type="primary" onClick={createNewChannel}>
          新建告警通道
        </Button>
      }
      title={<h2 style={{ marginBottom: 0 }}>告警通道</h2>}
    >
      <Table
        columns={columns}
        dataSource={listReceivers}
        rowKey="fingerprint"
        pagination={{ simple: true }}
      />
      {drawerStatus && (
        <ChannelDrawer
          width={880}
          status={drawerStatus}
          setStatus={setDrawerStatus} 
          name={clickedChannelName}
          onClose={() => setDrawerOpen(false)}
          open={drawerOpen}
        />
      )}
    </Card>
  );
}
