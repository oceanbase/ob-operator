import { Button, Card, message, Space, Table, Typography } from 'antd';

import { K8sClusterApi } from '@/api';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { getK8sObclusterListReq } from '@/services';
import { formatTime, isNullValue } from '@oceanbase/util';
import { Link, useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import Createk8sClusterModal from './Createk8sClusterModal';

const { Text } = Typography;
export default function K8sClusterList() {
  const access = useAccess();
  const [visible, setVisible] = useState<boolean>(false);
  const [editData, setEditData] = useState<string>({});
  const { data, refresh, loading } = useRequest(getK8sObclusterListReq);

  const K8sClustersList = data?.data;

  const { run: deleteK8sCluster } = useRequest(
    K8sClusterApi.deleteRemoteK8sCluster,
    {
      manual: true,
      onSuccess: ({ successful }) => {
        if (successful) {
          message.success('删除 k8s 集群成功');
        }
      },
    },
  );

  const columns: ColumnsType<API.TenantDetail> = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (text) => (
        <Text ellipsis={{ tooltip: text }}>
          <Link to={`${text}`}>{text}</Link>
        </Text>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
    },
    {
      title: '创建时间',
      width: 178,
      dataIndex: 'createdAt',
      render: (text) => {
        return <Text>{isNullValue(text) ? '-' : formatTime(text)}</Text>;
      },
    },

    {
      title: '操作',
      dataIndex: 'operation',
      render: (_, record) => {
        return (
          <Space>
            <Button
              onClick={() => {
                setVisible(true);
                setEditData(record);
              }}
              type="link"
            >
              编辑
            </Button>
            <Button
              type="link"
              onClick={() =>
                showDeleteConfirm({
                  title: '确定要删除该用户吗？',
                  onOk: () => deleteK8sCluster(record.name),
                })
              }
            >
              删除
            </Button>
          </Space>
        );
      },
    },
  ];

  return (
    <Card
      loading={loading}
      title={
        <div>
          <h2 style={{ marginBottom: 0 }}>K8s 集群管理</h2>
        </div>
      }
      extra={
        access.obclusterwrite ? (
          <Button
            type="primary"
            onClick={() => {
              setVisible(true);
            }}
          >
            创建 K8S 集群
          </Button>
        ) : null
      }
    >
      <Table
        columns={columns}
        dataSource={K8sClustersList}
        scroll={{ x: 1200 }}
        pagination={{ simple: true }}
        rowKey="name"
        bordered
        sticky
      />
      <Createk8sClusterModal
        visible={visible}
        onSuccess={() => {
          setVisible(false);
          setEditData({});
          refresh();
        }}
        onCancel={() => {
          setVisible(false);
          setEditData({});
        }}
        editData={editData}
      />
    </Card>
  );
}
