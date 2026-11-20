import NewLogService from '@/pages/LogService/New';
import { L } from '@/pages/LogService/common';
import {
  BucketInput,
  CredentialPicker,
  StoragePreflight,
  StorageSummary,
} from '@/pages/SharedStorage/ObjectStorage';
import {
  bucketURLRules,
  parseBucketURL,
  storageLocationsOverlap,
} from '@/pages/SharedStorage/storageConfig';
import { LSItem, lsRequest } from '@/services/logservice';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { FormInstance } from 'antd';
import {
  Alert,
  Button,
  Card,
  Col,
  Drawer,
  Form,
  Input,
  Select,
  Space,
  Tag,
} from 'antd';
import { useEffect, useRef, useState } from 'react';

export function StorageArchitecture() {
  return (
    <Card
      title={L('存储架构', 'Storage architecture')}
      style={{ marginBottom: 16 }}
    >
      <Form.Item
        name="deploymentMode"
        label={L('部署模式', 'Deployment mode')}
        rules={[{ required: true }]}
      >
        <Select
          options={[
            {
              value: 'normal',
              label: L('存算一体（normal）', 'Shared nothing (normal)'),
            },
            {
              value: 'shared_storage',
              label: L('湖库共享存储（SS）', 'Lakehouse shared storage (SS)'),
            },
          ]}
        />
      </Form.Item>
    </Card>
  );
}

export default function SharedStorage({
  form,
  section,
}: {
  form: FormInstance<API.CreateClusterData>;
  section: 'storage' | 'logservice';
}) {
  const namespace = Form.useWatch('namespace', form);
  const bucketURL = Form.useWatch(['sharedStorageInfo', 'bucketURL'], form);
  const secretName = Form.useWatch(
    ['sharedStorageInfo', 'secretRef', 'name'],
    form,
  );
  const selected = Form.useWatch(['logServiceRef', 'name'], form);
  const access = useAccess();
  const [creating, setCreating] = useState(false);
  const [creatingBusy, setCreatingBusy] = useState(false);
  const [reuseConnection, setReuseConnection] = useState(false);
  const previousNS = useRef(namespace);
  const {
    data: logServices,
    error,
    loading,
    refresh,
  } = useRequest(
    async () => ({
      namespace,
      items: await lsRequest<LSItem[]>('', { params: { namespace } }),
    }),
    {
      ready:
        !!namespace && section === 'logservice' && !!access.oblogserviceread,
      refreshDeps: [namespace],
      pollingInterval: 5000,
      pollingWhenHidden: false,
    },
  );
  const items =
    logServices && logServices.namespace === namespace ? logServices.items : [];
  const ls = items.find((v) => v.name === selected);
  useEffect(() => {
    if (previousNS.current !== namespace) {
      form.setFieldValue(['logServiceRef', 'name'], undefined);
      previousNS.current = namespace;
      setCreating(false);
    }
  }, [namespace]);
  if (section === 'storage')
    return (
      <Col span={24}>
        <Card title={L('OB 数据对象存储', 'OB data object storage')}>
          <Alert
            showIcon
            type="info"
            style={{ marginBottom: 16 }}
            message={L(
              '这里配置 OB 数据存储，不是 LogService 日志存储。创建后不支持在线更换存储位置、关联 LogService 或修改 IOPS/带宽。',
              'Configure OB data storage, not LogService log storage. Storage location, LogService reference and IOPS/bandwidth cannot be changed online after creation.',
            )}
          />
          <Form.Item
            name={['sharedStorageInfo', 'bucketURL']}
            label={L('数据存储位置', 'Data storage location')}
            rules={bucketURLRules}
          >
            <BucketInput />
          </Form.Item>
          <Form.Item
            name={['sharedStorageInfo', 'secretRef', 'name']}
            label={L('对象存储凭据', 'Object storage credentials')}
            rules={[{ required: true }]}
          >
            <CredentialPicker namespace={namespace} />
          </Form.Item>
          <Space align="start">
            <Form.Item
              name={['sharedStorageInfo', 'maxIOPS']}
              label={L('最大 IOPS（可选）', 'Max IOPS (optional)')}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name={['sharedStorageInfo', 'maxBandwidth']}
              label={L('最大带宽（可选）', 'Max bandwidth (optional)')}
            >
              <Input />
            </Form.Item>
          </Space>
          <StoragePreflight
            namespace={namespace}
            bucketURL={bucketURL}
            secretName={secretName}
          />
        </Card>
      </Col>
    );
  return (
    <Col span={24}>
      <Card title={L('关联日志服务', 'Log service association')}>
        <Alert
          showIcon
          type="info"
          style={{ marginBottom: 16 }}
          message={L(
            '一个 OB 集群关联一个 LogService 集群，不是一个节点。仅可选择同命名空间、运行中的 LogService。',
            'An OB cluster references one LogService cluster, not one node. Select a running LogService in the same namespace.',
          )}
        />
        {error && <Alert type="error" message={error.message} showIcon />}
        <Form.Item
          name={['logServiceRef', 'name']}
          label="LogService"
          rules={[
            { required: true },
            {
              validator: async (_, value) => {
                if (
                  access.oblogserviceread &&
                  value &&
                  (!ls || ls.deleting || ls.status.status !== 'running')
                )
                  throw new Error(
                    L(
                      '请等待 LogService running 后再继续',
                      'Wait for LogService to become running before continuing',
                    ),
                  );
              },
            },
          ]}
        >
          {access.oblogserviceread ? (
            <Select
              showSearch
              loading={loading}
              placeholder="LogService"
              options={items.map((v) => ({
                value: v.name,
                label: v.name + ' (' + (v.status.status || 'pending') + ')',
                disabled: v.status.status !== 'running' || v.deleting,
              }))}
            />
          ) : (
            <Input placeholder="ls-test" />
          )}
        </Form.Item>
        <Space style={{ marginBottom: 16 }}>
          <Button
            onClick={refresh}
            disabled={!access.oblogserviceread || !namespace}
          >
            {L('刷新状态', 'Refresh status')}
          </Button>
          <Button
            disabled={!access.oblogservicewrite || !namespace}
            onClick={() => setCreating(true)}
          >
            {L('在此创建 LogService', 'Create LogService here')}
          </Button>
        </Space>
        {ls && (
          <Card
            size="small"
            title={
              <Space>
                {ls.name}
                <Tag>{ls.status.status || 'pending'}</Tag>
              </Space>
            }
          >
            <p>
              {ls.spec.topology
                .map((z) => z.zone + ': ' + z.replica)
                .join(' · ')}
            </p>
            <StorageSummary
              bucketURL={ls.spec.objectStoreUrl.bucketURL}
              secretName={ls.spec.objectStoreUrl.secretRef.name}
            />
            {storageLocationsOverlap(
              bucketURL,
              ls.spec.objectStoreUrl.bucketURL,
            ) && (
              <Alert
                type="warning"
                showIcon
                message={L(
                  '数据与日志的 Bucket/前缀重叠，建议隔离存储位置。',
                  'Data and log bucket/prefix locations overlap; use separate locations.',
                )}
              />
            )}
          </Card>
        )}
        <Drawer
          title={L('创建 LogService', 'Create LogService')}
          open={creating}
          width="min(1100px, 95vw)"
          closable={!creatingBusy}
          keyboard={!creatingBusy}
          onClose={() => {
            if (!creatingBusy) {
              setCreating(false);
              setReuseConnection(false);
            }
          }}
          maskClosable={false}
          destroyOnClose
        >
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message={L(
              '可复用数据存储的 Endpoint、Region 和凭据；日志 Bucket/前缀需另行填写。',
              'You may reuse the data endpoint, region and credentials; enter a separate log bucket/prefix.',
            )}
            action={
              <Button
                disabled={
                  creatingBusy ||
                  reuseConnection ||
                  !parseBucketURL(bucketURL || '')
                }
                onClick={() => setReuseConnection(true)}
              >
                {L('复用连接与凭据', 'Reuse connection & credentials')}
              </Button>
            }
          />
          <NewLogService
            embedded
            namespace={namespace}
            connectionDefaults={
              reuseConnection
                ? { ...parseBucketURL(bucketURL)!, secret: secretName }
                : undefined
            }
            onBusyChange={setCreatingBusy}
            onCancel={() => {
              setCreating(false);
              setReuseConnection(false);
            }}
            onCreated={(v) => {
              form.setFieldValue(['logServiceRef', 'name'], v.name);
              setCreating(false);
              setReuseConnection(false);
              refresh();
            }}
          />
        </Drawer>
      </Card>
    </Col>
  );
}
