import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Descriptions, Form, Input, InputNumber, Modal, Space, Table, Tag, message } from 'antd';
import { history, useAccess, useLocation, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useState } from 'react';
import { LSDetail, lsPath, lsRequest } from '@/services/logservice';
import SSMonitor from '@/pages/SharedStorage/Monitor';
import { L, errorText } from './common';
import { StoragePreflight, StorageSummary } from '@/pages/SharedStorage/ObjectStorage';

export default function LogServiceDetail() {
  const { ns = '', name = '' } = useParams();
  const section = useLocation().pathname.split('/').filter(Boolean).pop();
  const access = useAccess();
  const [scaleForm] = Form.useForm();
  const [scaleVersion, setScaleVersion] = useState('');
  const [deleteVersion, setDeleteVersion] = useState('');
  const [confirmName, setConfirmName] = useState('');
  const [saving, setSaving] = useState(false);
  const [actionError, setActionError] = useState('');
  const { data, error, loading, refresh } = useRequest(() => lsRequest<LSDetail>(lsPath(ns, name)), { refreshDeps: [ns, name], pollingInterval: 5000 });
  const scale = async () => {
    try {
      const values = await scaleForm.validateFields(); setSaving(true); setActionError('');
      await lsRequest(`${lsPath(ns, name)}/replicas`, { method: 'PATCH', data: { resourceVersion: scaleVersion, replicas: values.replicas } });
      setScaleVersion(''); refresh(); message.success(L('扩缩容请求已提交；请观察 Zone 和节点状态', 'Scaling submitted; watch zone and node status'));
    } catch (e) { setActionError(errorText(e)); } finally { setSaving(false); }
  };
  const remove = async () => {
    setSaving(true); setActionError('');
    try {
      await lsRequest(lsPath(ns, name), { method: 'DELETE', data: { resourceVersion: deleteVersion, confirmName } });
      setDeleteVersion(''); message.success(L('删除请求已提交，由 Operator 清理资源', 'Deletion submitted; Operator will clean up resources')); history.push('/logservice');
    } catch (e) { setActionError(errorText(e)); } finally { setSaving(false); }
  };
  const writable = access.oblogservicewrite && !data?.deleting;
  const busy = data?.status.status !== 'running';
  return <PageContainer title={`LogService · ${name}`} loading={loading && !data} onBack={() => history.push('/logservice')} extra={<Space>
    <Button onClick={refresh}>{L('刷新', 'Refresh')}</Button>
    <Button disabled={!writable || busy} onClick={() => { scaleForm.setFieldsValue({ replicas: Object.fromEntries(data!.spec.topology.map(z => [z.zone, z.replica])) }); setActionError(''); setScaleVersion(data!.resourceVersion); }}>{L('扩缩容副本', 'Scale replicas')}</Button>
    <Button danger disabled={!writable || !data || data.protected || !!data.references.length} onClick={() => { setConfirmName(''); setActionError(''); setDeleteVersion(data!.resourceVersion); }}>{L('删除', 'Delete')}</Button>
  </Space>}>
    {error && <Alert type="error" showIcon message={error.message} />}
    {data && <>
      {data.references.length > 0 && <Alert type="info" showIcon message={`${L('被以下 OB 集群引用，禁止删除', 'Referenced by these OBClusters; deletion is blocked')}: ${data.references.join(', ')}`} style={{ marginBottom: 16 }} />}
      {busy && <Alert type={data.status.status === 'failed' ? 'error' : 'info'} showIcon message={`${L('当前状态', 'Current state')}: ${data.deleting ? 'deleting' : data.status.status || 'pending'}`} description={`${data.status.operationContext?.task || ''} ${data.status.operationContext?.taskStatus || ''}`} style={{ marginBottom: 16 }} />}
      {[
        { key: 'storage', label: L('日志对象存储', 'Log object storage'), children: <Card><Alert type="info" showIcon message={L('这是 LogService 日志存储，不是 OB 数据存储。创建后不支持在线修改位置或切换凭据。', 'This is LogService log storage, not OB data storage. Online location or credential-reference changes are unsupported.')} /><StorageSummary bucketURL={data.spec.objectStoreUrl.bucketURL} secretName={data.spec.objectStoreUrl.secretRef.name} /><StoragePreflight namespace={ns} bucketURL={data.spec.objectStoreUrl.bucketURL} secretName={data.spec.objectStoreUrl.secretRef.name} /></Card> },
        { key: 'overview', label: L('概览与拓扑', 'Overview & topology'), children: <Space direction="vertical" style={{ width: '100%' }} size="large">
          <Card><Descriptions column={2} items={[
            { key: 'ns', label: L('命名空间', 'Namespace'), children: data.namespace },
            { key: 'id', label: 'Cluster ID', children: data.spec.clusterId },
            { key: 'status', label: L('状态', 'State'), children: <Tag color={busy ? 'orange' : 'green'}>{data.status.status || 'pending'}</Tag> },
            { key: 'image', label: L('镜像', 'Image'), children: data.spec.logService.image },
            { key: 'resource', label: L('单节点资源', 'Per-node resources'), children: `${data.spec.logService.resource.cpu} CPU / ${data.spec.logService.resource.memory}` },
            { key: 'storage', label: 'Store / Log PVC', children: `${data.spec.logService.storage.storeStorage.size} / ${data.spec.logService.storage.logStorage.size}` },
            { key: 'bucket', label: 'Bucket URL', children: <span style={{ overflowWrap: 'anywhere' }}>{data.spec.objectStoreUrl.bucketURL}</span> },
            { key: 'secret', label: 'Secret', children: data.spec.objectStoreUrl.secretRef.name },
          ]} /></Card>
          <Card title={L('Zone 拓扑', 'Zone topology')}><Table rowKey="zone" pagination={false} dataSource={data.spec.topology} columns={[
            { title: 'Zone', dataIndex: 'zone' }, { title: L('期望副本', 'Desired replicas'), dataIndex: 'replica' },
            { title: L('Ready 节点', 'Ready nodes'), render: (_, z) => data.nodes.filter(n => n.zone === z.zone && n.status.ready && !n.deleting).length },
            { title: 'RPC / HTTP', render: (_, z) => `${z.rpcPort || 50051} / ${z.httpPort || 50052}` },
          ]} /></Card>
          <Card title={L('节点', 'Nodes')}><Table rowKey="name" dataSource={data.nodes} columns={[
            { title: L('节点', 'Node'), dataIndex: 'name' }, { title: 'Zone', dataIndex: 'zone' },
            { title: L('状态', 'Status'), render: (_, n) => `${n.deleting ? 'deleting' : n.status.status} / ${n.status.podPhase || '-'}` },
            { title: 'Ready', render: (_, n) => <Tag color={n.status.ready ? 'green' : 'orange'}>{String(n.status.ready)}</Tag> },
            { title: 'Pod IP / Service IP', render: (_, n) => `${n.status.podIP || '-'} / ${n.status.serviceIP || '-'}` },
          ]} /></Card>
          <Card title="PVC"><Table rowKey="name" dataSource={data.volumes} columns={[{ title: L('名称', 'Name'), dataIndex: 'name' }, { title: L('状态', 'Phase'), dataIndex: 'phase' }, { title: L('大小', 'Size'), dataIndex: 'size' }]} /></Card>
        </Space> },
        { key: 'monitor', label: L('LS 专项监控', 'LS monitoring'), children: <SSMonitor kind="logservice" namespace={ns} name={name} /> },
        { key: 'parameters', label: L('启动参数', 'Startup parameters'), children: <Card><Alert type="info" message={L('只读：当前 Operator 不支持参数热更新。', 'Read only: Operator does not support online parameter changes.')} /><Table rowKey="name" dataSource={data.spec.parameters} columns={[{ title: 'Name', dataIndex: 'name' }, { title: 'Value', dataIndex: 'value' }]} /></Card> },
        { key: 'events', label: L('事件', 'Events'), children: <Table rowKey={(e, i) => `${e.time}-${i}`} dataSource={data.events} columns={[
          { title: L('时间', 'Time'), render: (_, e) => new Date(e.time).toLocaleString() }, { title: L('对象', 'Object'), dataIndex: 'object' },
          { title: L('类型', 'Type'), dataIndex: 'type' }, { title: L('原因', 'Reason'), dataIndex: 'reason' }, { title: L('详情', 'Message'), dataIndex: 'message' },
        ]} /> },
      ].find(item => item.key === section)?.children}
    </>}
    <Modal open={!!scaleVersion} title={L('扩缩容 LogService 副本', 'Scale LogService replicas')} confirmLoading={saving} onCancel={() => !saving && setScaleVersion('')} onOk={scale} destroyOnClose>
      <Alert showIcon type="warning" message={L('缩容会注销并删除对应 LN 和节点资源。每个已有 Zone 至少保留一个副本；不更改 Zone、镜像或存储。', 'Scale-in unregisters the selected LN and removes its node resources. Every existing zone keeps at least one replica. Zones, image and storage remain unchanged.')} />
      {actionError && <Alert type="error" showIcon message={actionError} />}
      <Form form={scaleForm} layout="vertical" style={{ marginTop: 16 }}>{data?.spec.topology.map(z => <Form.Item key={z.zone} name={['replicas', z.zone]} label={`${z.zone} (${L('当前期望', 'currently desired')}: ${z.replica})`} rules={[{ required: true, type: 'number', min: 1 }]}><InputNumber min={1} precision={0} /></Form.Item>)}</Form>
    </Modal>
    <Modal open={!!deleteVersion} title={L('删除 LogService', 'Delete LogService')} confirmLoading={saving} okButtonProps={{ danger: true, disabled: confirmName !== name }} onCancel={() => !saving && setDeleteVersion('')} onOk={remove} destroyOnClose>
      <Alert type="warning" showIcon message={L('此操作会删除 LogService 及其托管资源，可能造成数据不可恢复。请确认没有使用者，并输入资源名继续。', 'This deletes LogService and its managed resources and may irreversibly remove data. Confirm it has no consumers, then type its name to continue.')} />
      {actionError && <Alert type="error" message={actionError} />}
      <Input aria-label={L('确认删除的资源名', 'Confirm resource name')} placeholder={name} value={confirmName} onChange={e => setConfirmName(e.target.value)} style={{ marginTop: 16 }} />
    </Modal>
  </PageContainer>;
}
