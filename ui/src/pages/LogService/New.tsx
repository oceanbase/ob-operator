import {
  BucketInput,
  CredentialPicker,
  StoragePreflight,
} from '@/pages/SharedStorage/ObjectStorage';
import {
  bucketURLRules,
  buildBucketURL,
} from '@/pages/SharedStorage/storageConfig';
import { LSItem, Zone, lsRequest } from '@/services/logservice';
import { PageContainer } from '@ant-design/pro-components';
import { history, useAccess } from '@umijs/max';
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Space,
  message,
} from 'antd';
import { useEffect, useState } from 'react';
import { validBootstrapReplicas } from './bootstrap';
import { L, errorText } from './common';

const required = [{ required: true, message: 'Required' }];
const dns = [
  {
    required: true,
    pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
    max: 63,
    message: 'DNS label, 1–63 characters',
  },
];
type Values = {
  name: string;
  namespace: string;
  clusterId: number;
  image: string;
  cpu: number;
  memory: number;
  storeSize: number;
  logSize: number;
  storageClass: string;
  bucketURL: string;
  secret: string;
  topology: Zone[];
  parameters?: { name: string; value: string }[];
};
export default function NewLogService({
  embedded = false,
  namespace,
  connectionDefaults,
  onCreated,
  onCancel,
  onBusyChange,
}: {
  embedded?: boolean;
  namespace?: string;
  connectionDefaults?: { endpoint: string; region: string; secret?: string };
  onCreated?: (ls: LSItem) => void;
  onCancel?: () => void;
  onBusyChange?: (busy: boolean) => void;
}) {
  const [form] = Form.useForm<Values>();
  const access = useAccess();
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const selectedNS = Form.useWatch('namespace', form) || namespace;
  const bucketURL = Form.useWatch('bucketURL', form);
  const secretName = Form.useWatch('secret', form);
  const cancel = () => (onCancel ? onCancel() : history.push('/logservice'));
  const defaultLocation = connectionDefaults
    ? {
        endpoint: connectionDefaults.endpoint,
        region: connectionDefaults.region,
        bucket: '',
        prefix: '',
      }
    : undefined;
  useEffect(() => {
    if (connectionDefaults)
      form.setFieldsValue({
        bucketURL: buildBucketURL({
          endpoint: connectionDefaults.endpoint,
          region: connectionDefaults.region,
          bucket: '',
          prefix: '',
        }),
        secret: connectionDefaults.secret,
      });
  }, [
    connectionDefaults?.endpoint,
    connectionDefaults?.region,
    connectionDefaults?.secret,
  ]);
  const submit = async (v: Values) => {
    setSaving(true);
    onBusyChange?.(true);
    setError('');
    try {
      const result = await lsRequest<LSItem>('', {
        method: 'POST',
        data: {
          namespace: v.namespace,
          name: v.name,
          spec: {
            clusterId: v.clusterId,
            logService: {
              image: v.image,
              resource: { cpu: String(v.cpu), memory: `${v.memory}Gi` },
              storage: {
                storeStorage: {
                  size: `${v.storeSize}Gi`,
                  storageClass: v.storageClass,
                },
                logStorage: {
                  size: `${v.logSize}Gi`,
                  storageClass: v.storageClass,
                },
              },
            },
            topology: v.topology,
            objectStoreUrl: {
              bucketURL: v.bucketURL,
              secretRef: { name: v.secret },
            },
            parameters: v.parameters || [],
          },
        },
      });
      message.success(
        L(
          '创建请求已提交，正在等待 Operator 就绪',
          'Creation submitted; waiting for Operator readiness',
        ),
      );
      if (onCreated) onCreated(result);
      else history.push(`/logservice/${result.namespace}/${result.name}`);
    } catch (e) {
      setError(errorText(e));
    } finally {
      setSaving(false);
      onBusyChange?.(false);
    }
  };
  const content = (
    <>
      <Alert
        showIcon
        type="info"
        message={L(
          '日志服务使用独立的对象存储位置。选择或新建同命名空间凭据；创建后仅副本数量可变更。LogService 单独创建，取消后续 OB 向导不会自动删除它。',
          'Use a separate log storage location and credentials in the same namespace. Only replicas may change after creation. LogService is created independently and retained if the OB wizard is cancelled.',
        )}
        style={{ marginBottom: 16 }}
      />
      {error && (
        <Alert
          type="error"
          message={error}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      <Form
        form={form}
        layout="vertical"
        onFinish={submit}
        disabled={!access.oblogservicewrite || saving}
        initialValues={{
          namespace,
          secret: connectionDefaults?.secret,
          bucketURL: defaultLocation
            ? buildBucketURL(defaultLocation)
            : undefined,
          cpu: 1,
          memory: 8,
          storeSize: 30,
          logSize: 15,
          topology: ['zone1', 'zone2', 'zone3'].map((zone) => ({
            zone,
            replica: 1,
            rpcPort: 50051,
            httpPort: 50052,
          })),
        }}
      >
        <Card title={L('基本配置', 'Basic configuration')}>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="namespace"
                label={L('命名空间', 'Namespace')}
                rules={dns}
              >
                <Input disabled={embedded} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="name"
                label={L('资源名', 'Resource name')}
                rules={dns}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="clusterId" label="Cluster ID" rules={required}>
                <InputNumber
                  min={1}
                  max={Number.MAX_SAFE_INTEGER}
                  precision={0}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="image"
            label={L('LogService 镜像', 'LogService image')}
            rules={required}
          >
            <Input placeholder="registry/oblogservice:version" />
          </Form.Item>
          <Row gutter={24}>
            {(
              [
                ['cpu', 'CPU', 0.1],
                ['memory', L('内存 (GiB)', 'Memory (GiB)'), 1],
                ['storeSize', L('Store PVC (GiB)', 'Store PVC (GiB)'), 1],
                ['logSize', L('Log PVC (GiB)', 'Log PVC (GiB)'), 1],
              ] as const
            ).map(([name, label, min]) => (
              <Col span={6} key={name}>
                <Form.Item name={name} label={label} rules={required}>
                  <InputNumber min={min} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            ))}
          </Row>
          <Form.Item
            name="storageClass"
            label="StorageClass"
            extra={L(
              '留空使用集群默认 StorageClass',
              'Leave empty to use the default StorageClass',
            )}
          >
            <Input />
          </Form.Item>
        </Card>
        <Card
          title={L('日志对象存储', 'Log object storage')}
          style={{ marginTop: 16 }}
        >
          <Form.Item
            name="bucketURL"
            label={L('日志存储位置', 'Log storage location')}
            rules={bucketURLRules}
          >
            <BucketInput defaultLocation={defaultLocation} />
          </Form.Item>
          <Form.Item
            name="secret"
            label={L(
              '同命名空间的对象存储凭据',
              'Object storage credentials in the same namespace',
            )}
            rules={required}
          >
            <CredentialPicker namespace={selectedNS} />
          </Form.Item>
          <StoragePreflight
            namespace={selectedNS}
            bucketURL={bucketURL}
            secretName={secretName}
          />
        </Card>
        <Card title={L('Zone 拓扑', 'Zone topology')} style={{ marginTop: 16 }}>
          <Alert
            showIcon
            type="info"
            style={{ marginBottom: 16 }}
            message={L(
              '首次初始化必须合计 3 个节点（已验证 oblogservice:1.3.0）；运行后可在现有 Zone 内扩缩容副本。',
              'Initialize with exactly 3 nodes (verified with oblogservice:1.3.0); replicas may be scaled in existing zones after initialization.',
            )}
          />
          <Form.List
            name="topology"
            rules={[
              {
                validator: async (_, zones: Zone[]) => {
                  if (
                    !zones?.length ||
                    new Set(zones.map((z) => z.zone)).size !== zones.length
                  )
                    throw new Error(
                      L(
                        '至少一个 Zone，且名称不能重复',
                        'At least one unique zone is required',
                      ),
                    );
                  if (!validBootstrapReplicas(zones))
                    throw new Error(
                      L(
                        '首次初始化的副本总数必须为 3；运行后再进行扩缩容',
                        'Initial replica count must total 3; scale after initialization',
                      ),
                    );
                },
              },
            ]}
          >
            {(fields, { add, remove }, { errors }) => (
              <>
                {fields.map((f) => (
                  <Row gutter={16} key={f.key}>
                    <Col span={6}>
                      <Form.Item
                        name={[f.name, 'zone']}
                        label="Zone"
                        rules={dns}
                      >
                        <Input />
                      </Form.Item>
                    </Col>
                    <Col span={4}>
                      <Form.Item
                        name={[f.name, 'replica']}
                        label={L('副本数', 'Replicas')}
                        rules={required}
                      >
                        <InputNumber min={1} precision={0} />
                      </Form.Item>
                    </Col>
                    <Col span={5}>
                      <Form.Item
                        name={[f.name, 'rpcPort']}
                        label="RPC port"
                        rules={required}
                      >
                        <InputNumber min={1} max={65535} precision={0} />
                      </Form.Item>
                    </Col>
                    <Col span={5}>
                      <Form.Item
                        name={[f.name, 'httpPort']}
                        label="HTTP port"
                        rules={required}
                      >
                        <InputNumber min={1} max={65535} precision={0} />
                      </Form.Item>
                    </Col>
                    <Col span={4}>
                      <Button
                        disabled={fields.length < 2}
                        onClick={() => remove(f.name)}
                        style={{ marginTop: 30 }}
                      >
                        {L('移除', 'Remove')}
                      </Button>
                    </Col>
                  </Row>
                ))}
                <Form.ErrorList errors={errors} />
                <Button
                  onClick={() =>
                    add({
                      zone: '',
                      replica: 1,
                      rpcPort: 50051,
                      httpPort: 50052,
                    })
                  }
                >
                  {L('添加 Zone', 'Add zone')}
                </Button>
              </>
            )}
          </Form.List>
        </Card>
        <Card
          title={L(
            '启动参数（可选；不支持热修改）',
            'Startup parameters (optional; no online changes)',
          )}
          style={{ marginTop: 16 }}
        >
          <Form.List name="parameters">
            {(fields, { add, remove }) => (
              <>
                {fields.map((f) => (
                  <Space key={f.key} align="baseline">
                    <Form.Item name={[f.name, 'name']} rules={required}>
                      <Input placeholder="name" />
                    </Form.Item>
                    <Form.Item name={[f.name, 'value']} rules={required}>
                      <Input placeholder="value" />
                    </Form.Item>
                    <Button onClick={() => remove(f.name)}>
                      {L('移除', 'Remove')}
                    </Button>
                  </Space>
                ))}
                <Button onClick={() => add()}>
                  {L('添加参数', 'Add parameter')}
                </Button>
              </>
            )}
          </Form.List>
        </Card>
        <Space style={{ marginTop: 24 }}>
          <Button disabled={saving} onClick={cancel}>
            {L('取消', 'Cancel')}
          </Button>
          <Button type="primary" htmlType="submit" loading={saving}>
            {L('创建 LogService', 'Create LogService')}
          </Button>
        </Space>
      </Form>
    </>
  );
  return embedded ? (
    content
  ) : (
    <PageContainer
      title={L('创建 LogService', 'Create LogService')}
      onBack={cancel}
    >
      {content}
    </PageContainer>
  );
}
