import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { L, errorText } from '@/pages/LogService/common';
import {
  CredentialReference,
  StorageCheck,
  storageRequest,
} from '@/services/objectstorage';
import { useAccess } from '@umijs/max';
import {
  Alert,
  Button,
  Col,
  Descriptions,
  Form,
  Input,
  Modal,
  Radio,
  Row,
  Select,
  Space,
  Typography,
  message,
} from 'antd';
import { useEffect, useRef, useState } from 'react';
import {
  StorageLocation,
  buildBucketURL,
  emptyLocation,
  parseBucketURL,
} from './storageConfig';

export function BucketInput({
  value = '',
  onChange,
  defaultLocation,
}: {
  value?: string;
  onChange?: (v: string) => void;
  defaultLocation?: StorageLocation;
}) {
  const [advanced, setAdvanced] = useState(false);
  const [fields, setFields] = useState<StorageLocation>(
    () => parseBucketURL(value) || defaultLocation || emptyLocation(),
  );
  useEffect(() => {
    if (!value) setFields(defaultLocation || emptyLocation());
    else if (defaultLocation && value === buildBucketURL(defaultLocation))
      setFields(defaultLocation);
    else if (value !== buildBucketURL(fields)) {
      const parsed = parseBucketURL(value);
      if (parsed) setFields(parsed);
    }
  }, [value, defaultLocation?.endpoint, defaultLocation?.region]);
  const edit = (key: keyof StorageLocation, v: string) => {
    const next = { ...fields, [key]: v };
    setFields(next);
    onChange?.(buildBucketURL(next));
  };
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Typography.Text>
        {L(
          '类型：S3 兼容（当前验证环境为 MinIO）；其他供应商和 STS 尚未验证。',
          'Type: S3 compatible (validated with MinIO); other providers and STS are not verified.',
        )}
      </Typography.Text>
      <Radio.Group
        value={advanced}
        onChange={(e) => setAdvanced(e.target.value)}
        options={[
          { label: L('结构化配置', 'Structured fields'), value: false },
          { label: L('高级：Bucket URL', 'Advanced: Bucket URL'), value: true },
        ]}
      />
      {advanced ? (
        <Input.TextArea
          aria-label="Bucket URL"
          rows={2}
          value={value}
          onChange={(e) => onChange?.(e.target.value)}
          placeholder="s3://bucket/prefix?host=https://endpoint&s3_region=us-east-1"
        />
      ) : (
        <Row gutter={[16, 12]} style={{ width: '100%' }}>
          {(
            [
              ['endpoint', 'Endpoint (http / https)', 'https://s3.example.com'],
              ['bucket', 'Bucket', 'sharedstorage'],
              ['region', 'Region', 'us-east-1'],
              [
                'prefix',
                L('路径前缀（可选）', 'Prefix (optional)'),
                'cluster-a',
              ],
            ] as const
          ).map(([key, label, placeholder]) => (
            <Col xs={24} md={12} key={key}>
              <label>
                {label}
                <Input
                  aria-label={label}
                  value={fields[key]}
                  onChange={(e) => edit(key, e.target.value)}
                  placeholder={placeholder}
                />
              </label>
            </Col>
          ))}
        </Row>
      )}
      {!advanced && value && (
        <Typography.Text type="secondary" style={{ overflowWrap: 'anywhere' }}>
          {parseBucketURL(value)
            ? value
            : L(
                '请补全有效的 Endpoint、Bucket、Region。',
                'Complete a valid Endpoint, Bucket and Region.',
              )}
        </Typography.Text>
      )}
      {fields.endpoint.startsWith('http://') && (
        <Alert
          type="warning"
          showIcon
          message={L(
            '对象存储使用 HTTP；正式环境建议 HTTPS。',
            'Object storage uses HTTP; use HTTPS in production.',
          )}
        />
      )}
    </Space>
  );
}

export function CredentialPicker({
  namespace,
  value,
  onChange,
}: {
  namespace?: string;
  value?: string;
  onChange?: (v?: string) => void;
}) {
  const access = useAccess();
  const publicKey = usePublicKey();
  const [items, setItems] = useState<CredentialReference[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [open, setOpen] = useState(false);
  const [revision, setRevision] = useState(0);
  const [form] = Form.useForm();
  const previousNS = useRef(namespace);
  useEffect(() => {
    if (previousNS.current !== namespace) {
      onChange?.(undefined);
      previousNS.current = namespace;
      setOpen(false);
      form.resetFields();
    }
    setItems([]);
    setError('');
    let active = true;
    if (namespace && access.objectstorageread) {
      setLoading(true);
      storageRequest<CredentialReference[]>(namespace, 'credentials')
        .then((v) => {
          if (active) setItems(v);
        })
        .catch((e) => {
          if (active) setError(errorText(e));
        })
        .finally(() => {
          if (active) setLoading(false);
        });
    }
    return () => {
      active = false;
    };
  }, [namespace, access.objectstorageread, revision]);
  const close = () => {
    setOpen(false);
    form.resetFields();
  };
  const create = async () => {
    const v = await form.validateFields();
    if (!namespace || !publicKey) return;
    setSaving(true);
    try {
      const encryptedAccessID = encryptText(v.accessID, publicKey),
        encryptedAccessKey = encryptText(v.accessKey, publicKey);
      if (!encryptedAccessID || !encryptedAccessKey)
        throw new Error(
          L(
            '凭据加密失败，请刷新页面',
            'Credential encryption failed; reload the page',
          ),
        );
      const ref = await storageRequest<CredentialReference>(
        namespace,
        'credentials',
        {
          method: 'POST',
          data: { name: v.name, encryptedAccessID, encryptedAccessKey },
        },
      );
      onChange?.(ref.name);
      close();
      setRevision((r) => r + 1);
      message.success(
        L(
          '凭据已创建；已有 Secret 未修改',
          'Credentials created; existing Secrets were not changed',
        ),
      );
    } catch (e) {
      message.error(errorText(e));
    } finally {
      setSaving(false);
    }
  };
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Space.Compact style={{ width: '100%' }}>
        {access.objectstorageread ? (
          <Select
            aria-label={L('对象存储凭据', 'Object storage credentials')}
            style={{ flex: 1 }}
            value={value}
            allowClear
            showSearch
            loading={loading}
            disabled={!namespace}
            onChange={onChange}
            placeholder={L(
              '选择同命名空间 Secret',
              'Select a Secret in this namespace',
            )}
            options={items.map((v) => ({ value: v.name, label: v.name }))}
          />
        ) : (
          <Input
            aria-label="Secret name"
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            placeholder={L(
              '输入同命名空间 Secret 名称',
              'Secret name in the same namespace',
            )}
            disabled={!namespace}
          />
        )}
        {access.objectstorageread && (
          <Button
            disabled={!namespace}
            loading={loading}
            onClick={() => setRevision((r) => r + 1)}
          >
            {L('刷新', 'Refresh')}
          </Button>
        )}
        {access.objectstoragewrite && (
          <Button disabled={!namespace} onClick={() => setOpen(true)}>
            {L('新建凭据', 'New credentials')}
          </Button>
        )}
      </Space.Compact>
      {error && <Alert type="error" message={error} showIcon />}
      <Typography.Text type="secondary">
        {L(
          '仅显示 Secret 名称，不读取或回显密钥；跨命名空间不可复用。',
          'Only Secret names are displayed, never credential values. References must be in the same namespace.',
        )}
      </Typography.Text>
      <Modal
        title={L('新建对象存储凭据', 'Create object storage credentials')}
        open={open}
        onCancel={close}
        onOk={create}
        confirmLoading={saving}
        closable={!saving}
        maskClosable={!saving}
        cancelButtonProps={{ disabled: saving }}
        okButtonProps={{ disabled: !publicKey || !namespace }}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message={L(
            '凭据单独创建并保留；取消后续向导不会删除它。此处不支持覆盖、轮换或删除已有凭据。正式环境请使用 HTTPS 访问 Dashboard。',
            'Credentials are created independently and retained if the wizard is cancelled. No overwrite, rotation or deletion. Use Dashboard over HTTPS in production.',
          )}
        />
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item
            name="name"
            label="Secret name"
            rules={[
              {
                required: true,
                max: 63,
                pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
              },
            ]}
          >
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="accessID"
            label="Access ID"
            rules={[{ required: true, whitespace: true, max: 200 }]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="accessKey"
            label="Access Key"
            rules={[{ required: true, whitespace: true, max: 200 }]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}

const codes: Record<string, [string, string]> = {
  bucket_accessible: [
    'Bucket 可访问（仅 HeadBucket 检查通过）',
    'Bucket accessible (HeadBucket only)',
  ],
  connection_failed: [
    '连接失败：检查 DNS、网络、TLS 或受限地址',
    'Connection failed: check DNS, network, TLS or restricted addresses',
  ],
  access_denied: [
    '访问被拒绝：检查凭据、Region 和 Bucket 权限',
    'Access denied: check credentials, region and bucket permissions',
  ],
  bucket_not_found_or_denied: [
    'Bucket 不存在或访问被拒绝',
    'Bucket not found or access denied',
  ],
  redirect_refused: [
    'Endpoint 返回重定向；请填写正确的 Endpoint/Region',
    'Redirect refused; use the correct endpoint and region',
  ],
  endpoint_error: [
    '对象存储返回异常状态',
    'Object storage returned an unexpected status',
  ],
};
export function StoragePreflight({
  namespace,
  bucketURL,
  secretName,
}: {
  namespace?: string;
  bucketURL?: string;
  secretName?: string;
}) {
  const access = useAccess();
  const [result, setResult] = useState<StorageCheck>();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const generation = useRef(0);
  useEffect(() => {
    generation.current++;
    setResult(undefined);
    setError('');
    setLoading(false);
    return () => {
      generation.current++;
    };
  }, [namespace, bucketURL, secretName]);
  const check = async () => {
    if (!namespace || !bucketURL || !secretName) return;
    const requestGeneration = ++generation.current;
    setResult(undefined);
    setLoading(true);
    setError('');
    try {
      const r = await storageRequest<StorageCheck>(namespace, 'check', {
        method: 'POST',
        data: { bucketURL, secretName },
      });
      if (requestGeneration === generation.current) setResult(r);
    } catch (e) {
      if (requestGeneration === generation.current) setError(errorText(e));
    } finally {
      if (requestGeneration === generation.current) setLoading(false);
    }
  };
  const text = result ? codes[result.code] || codes.endpoint_error : undefined;
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Button
        loading={loading}
        onClick={check}
        disabled={
          !access.objectstoragewrite ||
          !namespace ||
          !secretName ||
          !bucketURL ||
          !parseBucketURL(bucketURL)
        }
      >
        {L('检查连接与 Bucket 访问', 'Check connection and bucket access')}
      </Button>
      <Typography.Text type="secondary">
        {L(
          '从 Dashboard 后端发起只读 HeadBucket；不写入对象，不验证对象读写/删除和路径权限。此结果不是持续健康监控。',
          'Read-only HeadBucket from Dashboard, with no object writes. Object read/write/delete and prefix permissions are not verified. This is not continuous health monitoring.',
        )}
        {!access.objectstoragewrite &&
          ` ${L(
            '需要对象存储写权限。',
            'Requires objectstorage write permission.',
          )}`}
      </Typography.Text>
      {error && <Alert type="error" showIcon message={error} />}
      {result && text && (
        <Alert
          type={result.ok ? 'success' : 'warning'}
          showIcon
          message={L(text[0], text[1])}
          description={`${new Date(result.checkedAt).toLocaleString()}${
            result.httpStatus ? ` · HTTP ${result.httpStatus}` : ''
          }`}
        />
      )}
    </Space>
  );
}

export function StorageSummary({
  bucketURL,
  secretName,
  maxIOPS,
  maxBandwidth,
}: {
  bucketURL?: string;
  secretName?: string;
  maxIOPS?: string;
  maxBandwidth?: string;
}) {
  const loc = bucketURL && parseBucketURL(bucketURL);
  return (
    <>
      {!loc && (
        <Alert
          type="warning"
          showIcon
          message={L(
            '无法按当前 S3 表单解析此配置；未修改原配置，也不会回显可能含凭据的原始 URL。',
            'This configuration cannot be parsed by the S3 form. Original settings are unchanged and potentially sensitive raw URLs are not displayed.',
          )}
        />
      )}
      <Descriptions
        column={{ xs: 1, sm: 2 }}
        size="small"
        style={{ marginTop: 12 }}
      >
        <Descriptions.Item label="Endpoint">
          {loc ? loc.endpoint : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Bucket">
          {loc ? loc.bucket : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Region">
          {loc ? loc.region : '-'}
        </Descriptions.Item>
        <Descriptions.Item label={L('路径前缀', 'Prefix')}>
          {loc ? loc.prefix || '/' : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Secret">
          {secretName || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="TLS">
          {loc ? (loc.endpoint.startsWith('https:') ? 'HTTPS' : 'HTTP') : '-'}
        </Descriptions.Item>
        {maxIOPS !== undefined && (
          <Descriptions.Item label={L('最大 IOPS', 'Max IOPS')}>
            {maxIOPS || L('未配置', 'Not configured')}
          </Descriptions.Item>
        )}
        {maxBandwidth !== undefined && (
          <Descriptions.Item label={L('最大带宽', 'Max bandwidth')}>
            {maxBandwidth || L('未配置', 'Not configured')}
          </Descriptions.Item>
        )}
      </Descriptions>
    </>
  );
}
