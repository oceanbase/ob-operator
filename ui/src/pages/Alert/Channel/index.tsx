import { alert } from '@/api';
import type { ReceiverReceiver } from '@/api/generated';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { useRequest } from 'ahooks';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';

export default function Channel() {
  const { data: listReceiversRes,refresh } = useRequest(alert.listReceivers);
  const { run:deleteReceiver } = useRequest(alert.deleteReceiver,{
    onSuccess:({successful})=>{
        if(successful){
            refresh();
        }
    }
  });
  let listReceivers = listReceiversRes?.data;
  listReceivers = [
    {
      config: 'string',
      name: 'string',
      type: 'dingtalk',
    },
  ];
  const columns: ColumnsType<ReceiverReceiver> = [
    {
      title: '通道名',
      dataIndex: 'name',
      key: 'name',
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
      render: (_,record) => (
        <>
          <Button type="link">编辑</Button>
          <Button
            style={{color:'#ff4b4b'}}
            onClick={() => {
              showDeleteConfirm({
                title: '确定要删除“钉钉群”告警通道吗？',
                content: '删除后使用该通道的消息推送都将失效，请谨慎操作',
                onOk: () => {
                    deleteReceiver(record.name);
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
      extra={<Button type="primary">新建告警通道</Button>}
      title={<h2 style={{ marginBottom: 0 }}>告警通道</h2>}
    >
      <Table
        columns={columns}
        dataSource={listReceivers}
        rowKey="fingerprint"
        pagination={{ simple: true }}
      />
    </Card>
  );
}
