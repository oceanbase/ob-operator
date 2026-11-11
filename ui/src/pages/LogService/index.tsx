import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Input, Space, Table, Tag } from 'antd';
import { history, useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useState } from 'react';
import { LSItem, lsRequest } from '@/services/logservice';
import { L } from './common';

export default function LogServiceList() {
  const access = useAccess();
  const [filter, setFilter] = useState('');
  const { data, error, loading, refresh } = useRequest(() => lsRequest<LSItem[]>(), { pollingInterval: 10000 });
  return <PageContainer title="LogService" extra={<Space>
    <Button onClick={refresh}>{L('刷新', 'Refresh')}</Button>
    <Button type="primary" disabled={!access.oblogservicewrite} onClick={() => history.push('/logservice/new')}>{L('创建 LogService', 'Create LogService')}</Button>
  </Space>}>
    <Alert type="info" showIcon message={L('支持现有 Zone 内副本扩缩容；暂不支持在线加减 Zone、升级镜像或调整 CPU、内存、磁盘。', 'Scale replicas in existing zones. Online zone changes, image upgrades, CPU, memory and disk changes are not supported.')} style={{ marginBottom: 16 }} />
    {error && <Alert type="error" showIcon message={error.message} />}
    <Input.Search aria-label={L('筛选 LogService', 'Filter LogServices')} placeholder={L('按名称或命名空间筛选', 'Filter by name or namespace')} onChange={e => setFilter(e.target.value)} style={{ width: 360, marginBottom: 16 }} />
    <Table<LSItem> rowKey={r => `${r.namespace}/${r.name}`} loading={loading} dataSource={data?.filter(r => `${r.namespace}/${r.name}`.includes(filter))} columns={[
      { title: L('名称', 'Name'), dataIndex: 'name', render: (_, r) => <a onClick={() => history.push(`/logservice/${r.namespace}/${r.name}`)}>{r.name}</a> },
      { title: L('命名空间', 'Namespace'), dataIndex: 'namespace' },
      { title: 'Cluster ID', render: (_, r) => r.spec.clusterId },
      { title: L('状态', 'Status'), render: (_, r) => <Tag color={r.status.status === 'running' ? 'green' : 'orange'}>{r.deleting ? 'deleting' : r.status.status || 'pending'}</Tag> },
      { title: L('拓扑 / 副本', 'Topology / replicas'), render: (_, r) => r.spec.topology.map(z => `${z.zone}: ${z.replica}`).join(' · ') },
      { title: L('镜像', 'Image'), render: (_, r) => r.spec.logService?.image },
      { title: L('创建时间', 'Created'), render: (_, r) => new Date(r.createdAt).toLocaleString() },
    ]} />
  </PageContainer>;
}
