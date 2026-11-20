import type { ResponseOBCluster } from '@/api/generated';
import { L } from '@/pages/LogService/common';
import SSMonitor from '@/pages/SharedStorage/Monitor';
import {
  StoragePreflight,
  StorageSummary,
} from '@/pages/SharedStorage/ObjectStorage';
import { storageLocationsOverlap } from '@/pages/SharedStorage/storageConfig';
import { LSDetail, lsPath, lsRequest } from '@/services/logservice';
import { PageContainer } from '@ant-design/pro-components';
import { history, request, useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Row,
  Space,
  Table,
  Tabs,
  Tag,
} from 'antd';

const gib = (bytes?: number) =>
  bytes === undefined ? '-' : `${Number((bytes / 1024 ** 3).toFixed(2))} GiB`;
export default function Lakehouse() {
  const { ns = '', name = '', clusterName = '' } = useParams();
  const access = useAccess();
  const base = `/cluster/${encodeURIComponent(ns)}/${encodeURIComponent(
    name,
  )}/${encodeURIComponent(clusterName)}`;
  const {
    data: clusterData,
    error,
    loading,
    refresh,
  } = useRequest(
    async () => {
      const r = await request<{
        successful: boolean;
        message: string;
        data: ResponseOBCluster;
      }>(
        `/api/v1/obclusters/namespace/${encodeURIComponent(
          ns,
        )}/name/${encodeURIComponent(name)}`,
      );
      if (!r.successful) throw new Error(r.message);
      return r.data;
    },
    { refreshDeps: [ns, name] },
  );
  const data =
    clusterData?.namespace === ns && clusterData.name === name
      ? clusterData
      : undefined;
  const lsName = data?.logServiceRef?.name;
  const {
    data: lsData,
    error: lsError,
    refresh: refreshLS,
  } = useRequest(
    async () => ({
      namespace: ns,
      name: lsName,
      detail: await lsRequest<LSDetail>(lsPath(ns, lsName!)),
    }),
    { ready: !!lsName && !!access.oblogserviceread, refreshDeps: [ns, lsName] },
  );
  const ls =
    lsData?.namespace === ns && lsData.name === lsName
      ? lsData.detail
      : undefined;
  const store = data?.sharedStorageInfo;
  const logs = ls?.spec.objectStoreUrl;
  const config = (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Alert
        type="info"
        showIcon
        message={L(
          '只读配置：当前 Operator 不支持在线更换数据存储、LogService 关联、IOPS/带宽或升级 SS 镜像。LogService 仅支持已有 Zone 内扩缩容副本。',
          'Read-only configuration: online data storage/LogService reassignment, IOPS/bandwidth changes and SS image upgrades are unsupported. LogService supports replica scaling within existing zones only.',
        )}
      />
      <Card
        title={L('计算与本地缓存', 'Compute and local cache')}
        extra={
          <Button onClick={() => history.push(base + '/topo')}>
            {L('查看拓扑', 'View topology')}
          </Button>
        }
      >
        <Descriptions column={2}>
          <Descriptions.Item label={L('集群', 'Cluster')}>
            {data?.clusterName} <Tag>{data?.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item
            label={L('命名空间 / 资源名', 'Namespace / resource')}
          >
            {ns} / {name}
          </Descriptions.Item>
          <Descriptions.Item
            label={L('单节点 CPU / 内存', 'Per-node CPU / memory')}
          >
            {data?.resource.cpu} CPU / {gib(data?.resource.memory)}
          </Descriptions.Item>
          <Descriptions.Item
            label={L('本地数据缓存 PVC', 'Local data cache PVC')}
          >
            {gib(data?.storage.dataStorage.size)} ·{' '}
            {data?.storage.dataStorage.storageClass ||
              L('默认 StorageClass', 'Default StorageClass')}
          </Descriptions.Item>
          <Descriptions.Item label={L('运行日志 PVC', 'Runtime log PVC')}>
            {gib(data?.storage.sysLogStorage.size)}
          </Descriptions.Item>
          <Descriptions.Item label={L('镜像', 'Image')}>
            {data?.image}
          </Descriptions.Item>
        </Descriptions>
        <Table
          size="small"
          rowKey="zone"
          dataSource={data?.topology}
          pagination={false}
          columns={[
            { title: 'Zone', dataIndex: 'zone' },
            {
              title: L('期望 OB 节点', 'Desired OB nodes'),
              dataIndex: 'replicas',
            },
            { title: L('状态', 'Status'), dataIndex: 'status' },
          ]}
        />
      </Card>
      <Card
        title={L('关联 LogService 集群', 'Associated LogService cluster')}
        extra={
          lsName && access.oblogserviceread ? (
            <Button
              onClick={() => history.push('/logservice' + lsPath(ns, lsName))}
            >
              {L('管理 / 扩缩容副本', 'Manage / scale replicas')}
            </Button>
          ) : null
        }
      >
        <Descriptions>
          <Descriptions.Item label="LogService">
            {lsName || '-'}
          </Descriptions.Item>
          <Descriptions.Item label={L('状态', 'Status')}>
            {ls?.status.status || '-'}
          </Descriptions.Item>
          <Descriptions.Item
            label={L('Ready / 期望副本', 'Ready / desired replicas')}
          >
            {ls
              ? `${
                  ls.nodes.filter((n) => n.status.ready && !n.deleting).length
                } / ${ls.spec.topology.reduce((n, z) => n + z.replica, 0)}`
              : '-'}
          </Descriptions.Item>
        </Descriptions>
        {!access.oblogserviceread && (
          <Alert
            type="info"
            message={L(
              '需要 LogService 读权限才能查看日志服务详情。',
              'LogService read permission is required for details.',
            )}
          />
        )}
        {lsError && <Alert type="error" showIcon message={lsError.message} />}
        {ls && (
          <Table
            rowKey="zone"
            size="small"
            pagination={false}
            dataSource={ls.spec.topology}
            columns={[
              { title: 'Zone', dataIndex: 'zone' },
              {
                title: L('期望副本', 'Desired replicas'),
                dataIndex: 'replica',
              },
              {
                title: 'Ready',
                render: (_, z) =>
                  ls.nodes.filter(
                    (n) => n.zone === z.zone && n.status.ready && !n.deleting,
                  ).length,
              },
            ]}
          />
        )}
      </Card>
      {storageLocationsOverlap(store?.bucketURL, logs?.bucketURL) && (
        <Alert
          type="warning"
          showIcon
          message={L(
            'OB 数据与 LS 日志的 Bucket/前缀重叠；建议隔离存储位置。现有配置未修改。',
            'OB data and LS log bucket/prefix locations overlap. Existing settings are unchanged.',
          )}
        />
      )}
      <Row gutter={[16, 16]} style={{ width: '100%' }}>
        <Col xs={24} xl={12}>
          <Card title={L('OB 数据对象存储', 'OB data object storage')}>
            <StorageSummary
              bucketURL={store?.bucketURL}
              secretName={store?.secretRef.name}
              maxIOPS={store?.maxIOPS || ''}
              maxBandwidth={store?.maxBandwidth || ''}
            />
            <StoragePreflight
              namespace={ns}
              bucketURL={store?.bucketURL}
              secretName={store?.secretRef.name}
            />
          </Card>
        </Col>
        <Col xs={24} xl={12}>
          <Card
            title={L(
              'LogService 日志对象存储',
              'LogService log object storage',
            )}
          >
            {logs ? (
              <>
                <StorageSummary
                  bucketURL={logs.bucketURL}
                  secretName={logs.secretRef.name}
                />
                <StoragePreflight
                  namespace={ns}
                  bucketURL={logs.bucketURL}
                  secretName={logs.secretRef.name}
                />
              </>
            ) : (
              <Alert
                type="info"
                message={L(
                  '日志存储配置尚未获取，见上方关联状态或权限提示。',
                  'Log storage settings have not been loaded; see association status or permissions above.',
                )}
              />
            )}
          </Card>
        </Col>
      </Row>
    </Space>
  );
  return (
    <PageContainer
      title={L('湖库配置', 'Lakehouse configuration')}
      loading={loading && !data}
      extra={
        <Button
          onClick={() => {
            refresh();
            if (lsName && access.oblogserviceread) refreshLS();
          }}
        >
          {L('刷新配置', 'Refresh configuration')}
        </Button>
      }
    >
      {error && <Alert type="error" showIcon message={error.message} />}
      {data &&
        (data.deploymentMode !== 'shared_storage' ? (
          <Alert
            type="info"
            message={L(
              '该集群不是湖库共享存储模式。',
              'This cluster is not in shared-storage mode.',
            )}
          />
        ) : (
          <Tabs
            items={[
              {
                key: 'configuration',
                label: L('配置与依赖', 'Configuration & dependencies'),
                children: config,
              },
              {
                key: 'monitor',
                label: L(
                  '对象存储与缓存监控',
                  'Object storage & cache monitoring',
                ),
                children: (
                  <SSMonitor kind="obcluster" namespace={ns} name={name} />
                ),
              },
              ...(lsName && access.oblogserviceread
                ? [
                    {
                      key: 'logs',
                      label: L('日志服务监控', 'Log service monitoring'),
                      children: (
                        <SSMonitor
                          kind="logservice"
                          namespace={ns}
                          name={lsName}
                        />
                      ),
                    },
                  ]
                : []),
            ]}
          />
        ))}
    </PageContainer>
  );
}
