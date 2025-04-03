import { intl } from '@/utils/intl';
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
          message.success(
            intl.formatMessage({
              id: 'src.pages.K8sCluster.F485EB45',
              defaultMessage: '删除 k8s 集群成功',
            }),
          );
          refresh();
        }
      },
    },
  );

  const columns: ColumnsType<API.TenantDetail> = [
    {
      title: intl.formatMessage({
        id: 'src.pages.K8sCluster.2588408F',
        defaultMessage: '名称',
      }),
      dataIndex: 'name',
      render: (text) => (
        <Text ellipsis={{ tooltip: text }}>
          <Link to={`${text}`}>{text}</Link>
        </Text>
      ),
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.K8sCluster.22D1DE51',
        defaultMessage: '描述',
      }),
      dataIndex: 'description',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.K8sCluster.0D18EFBC',
        defaultMessage: '创建时间',
      }),
      width: 178,
      dataIndex: 'createdAt',
      render: (text) => {
        return <Text>{isNullValue(text) ? '-' : formatTime(text)}</Text>;
      },
    },

    {
      title: intl.formatMessage({
        id: 'src.pages.K8sCluster.F4F859EC',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',

      render: (_, record) => {
        return (
          <Space>
            <a
              onClick={() => {
                setVisible(true);
                setEditData(record);
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.K8sCluster.D6091626',
                defaultMessage: '编辑',
              })}
            </a>
            <Button
              type="link"
              onClick={() =>
                showDeleteConfirm({
                  title: intl.formatMessage({
                    id: 'src.pages.K8sCluster.CE727163',
                    defaultMessage: '确定要删除该 K8s 集群吗？',
                  }),
                  onOk: () => deleteK8sCluster(record.name),
                })
              }
            >
              {intl.formatMessage({
                id: 'src.pages.K8sCluster.5C84A300',
                defaultMessage: '删除',
              })}
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
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'src.pages.K8sCluster.5689B3FE',
              defaultMessage: 'K8s 集群管理',
            })}
          </h2>
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
            {intl.formatMessage({
              id: 'src.pages.K8sCluster.F57B4076',
              defaultMessage: '创建 K8S 集群',
            })}
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
