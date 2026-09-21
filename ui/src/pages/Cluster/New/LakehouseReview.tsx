import { L } from '@/pages/LogService/common';
import {
  StoragePreflight,
  StorageSummary,
} from '@/pages/SharedStorage/ObjectStorage';
import type { FormInstance } from 'antd';
import { Alert, Card, Descriptions, Form, Space, Table } from 'antd';

export default function LakehouseReview({
  form,
}: {
  form: FormInstance<API.CreateClusterData>;
}) {
  return (
    <Form.Item noStyle shouldUpdate>
      {() => {
        const v = form.getFieldsValue(true);
        return (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Alert
              showIcon
              type="warning"
              message={L(
                '确认后创建 OB 集群；不会修改已有 LogService 或凭据。向导中已创建的 LogService/Secret 会独立保留。',
                'Confirmation creates an OB cluster without changing existing LogServices or credentials. LogServices/Secrets already created in this wizard remain independent.',
              )}
            />
            <Card title={L('创建前检查', 'Review before creation')}>
              <Descriptions column={2}>
                <Descriptions.Item label={L('命名空间', 'Namespace')}>
                  {v.namespace}
                </Descriptions.Item>
                <Descriptions.Item label={L('资源名称', 'Resource name')}>
                  {v.name}
                </Descriptions.Item>
                <Descriptions.Item label={L('集群名', 'Cluster name')}>
                  {v.clusterName}
                </Descriptions.Item>
                <Descriptions.Item label="LogService">
                  {v.logServiceRef?.name}
                </Descriptions.Item>
                <Descriptions.Item label={L('镜像', 'Image')}>
                  {v.observer?.image}
                </Descriptions.Item>
                <Descriptions.Item
                  label={L('单节点资源', 'Per-node resources')}
                >
                  {v.observer?.resource?.cpu} CPU /{' '}
                  {v.observer?.resource?.memory} GiB
                </Descriptions.Item>
                <Descriptions.Item label={L('本地缓存盘', 'Local cache disk')}>
                  {v.observer?.storage?.data?.size} GiB ·{' '}
                  {v.observer?.storage?.data?.storageClass ||
                    L('默认 StorageClass', 'Default StorageClass')}
                </Descriptions.Item>
              </Descriptions>
              <Table
                rowKey="zone"
                size="small"
                pagination={false}
                dataSource={v.topology}
                columns={[
                  { title: 'Zone', dataIndex: 'zone' },
                  { title: L('节点数', 'Nodes'), dataIndex: 'replicas' },
                ]}
              />
            </Card>
            <Card title={L('OB 数据对象存储', 'OB data object storage')}>
              <StorageSummary
                bucketURL={v.sharedStorageInfo?.bucketURL}
                secretName={v.sharedStorageInfo?.secretRef?.name}
                maxIOPS={v.sharedStorageInfo?.maxIOPS || ''}
                maxBandwidth={v.sharedStorageInfo?.maxBandwidth || ''}
              />
              <StoragePreflight
                namespace={v.namespace}
                bucketURL={v.sharedStorageInfo?.bucketURL}
                secretName={v.sharedStorageInfo?.secretRef?.name}
              />
            </Card>
            <Alert
              showIcon
              type="info"
              message={L(
                '创建后不支持在线更换对象存储、关联 LogService 或升级 SS 镜像。连接检查不代表对象读写/删除和前缀权限均已验证。',
                'Online storage/LogService reassignment and SS image upgrades are unsupported. Connection checks do not prove object read/write/delete or prefix permissions.',
              )}
            />
          </Space>
        );
      }}
    </Form.Item>
  );
}
