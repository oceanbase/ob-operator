import { Button, Card, message, Space, Table, Typography } from 'antd';

import { K8sClusterApi } from '@/api';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { getK8sObclusterListReq } from '@/services';
import { formatTime, isNullValue } from '@oceanbase/util';
import { Link, useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import Createk8sClusterModal from './Createk8sClusterModal';

const { Text } = Typography;
export default function K8sClusterList() {
  const { initialState } = useModel('@@initialState');
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
          refresh();
        }
      },
    },
  );

  const allPolicies = initialState?.policies?.filter(
    (policy) => policy.domain === 'k8s-cluster',
  );
  const k8sClusterAccess = allPolicies?.map(
    (item) => `k8sCluster${item.action}`,
  );

  const k8sClusterwrite = k8sClusterAccess?.includes('k8sClusterwrite');
  const k8sClusterread = k8sClusterAccess?.includes('k8sClusterread');

  const columns: ColumnsType<API.TenantDetail> = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (text) => (
        <Text ellipsis={{ tooltip: text }}>
          {k8sClusterread ? <Link to={`${text}`}>{text}</Link> : text}
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
        return k8sClusterwrite ? (
          <Space>
            <a
              onClick={() => {
                setVisible(true);
                setEditData(record);
              }}
            >
              编辑
            </a>
            <Button
              type="link"
              onClick={() =>
                showDeleteConfirm({
                  title: '确定要删除该 K8s 集群吗？',
                  onOk: () => deleteK8sCluster(record.name),
                })
              }
            >
              删除
            </Button>
          </Space>
        ) : null;
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
        k8sClusterwrite ? (
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
