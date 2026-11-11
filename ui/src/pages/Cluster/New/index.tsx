import { getStorageClasses } from '@/services';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useNavigate, useSearchParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Alert, Button, Col, Form, Row, Steps, message } from 'antd';
import { useEffect, useState } from 'react';

import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { L, errorText } from '@/pages/LogService/common';
import { createClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import { strTrim } from '@/utils/helper';
import BackUp from './BackUp';
import BasicInfo from './BasicInfo';
import LakehouseReview from './LakehouseReview';
import Monitor from './Monitor';
import Observer from './Observer';
import Parameters from './Parameters';
import SharedStorage, { StorageArchitecture } from './SharedStorage';
import Topo from './Topo';
import { normalizeStorageMode } from './storageMode';

export default function New() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [passwordVal, setPasswordVal] = useState<string>('');
  const [proxyroPasswordVal, setProxyroPasswordVal] = useState<string>('');
  const [form] = Form.useForm<API.CreateClusterData>();
  const access = useAccess();
  const isSS = Form.useWatch('deploymentMode', form) === 'shared_storage';
  const [step, setStep] = useState(0);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState('');
  useEffect(() => {
    setStep(0);
    setSubmitError('');
  }, [isSS]);
  const next = async () => {
    try {
      const fields =
        step === 0
          ? [
              'namespace',
              'name',
              'clusterName',
              'mode',
              'scenario',
              'rootPassword',
              'proxyroPassword',
              'sharedStorageInfo',
            ]
          : step === 1
          ? ['logServiceRef']
          : undefined;
      await form.validateFields(fields, { recursive: true });
      setStep((s) => Math.min(s + 1, 3));
    } catch {
      message.warning(
        L('请检查当前步骤的配置', 'Check the configuration in this step'),
      );
    }
  };
  const [pvcValue, setPvcValue] = useState<boolean>(false);
  const [deleteValue, setDeleteValue] = useState<boolean>(false);
  const { data: storageClassesRes, run: fetchStorageClasses } = useRequest(
    getStorageClasses,
    {
      onSuccess: ({ successful, data }) => {
        if (successful && data.length === 1) {
          const { value } = data[0];
          form.setFieldValue(['observer', 'storage'], {
            data: {
              storageClass: value,
            },
            log: {
              storageClass: value,
            },
            redoLog: {
              storageClass: value,
            },
          });
        }
      },
      manual: true,
    },
  );
  const publicKey = usePublicKey();
  const storageClasses = storageClassesRes?.data;
  const onFinish = async (values: API.CreateClusterData) => {
    if (isSS && step !== 3) {
      await next();
      return;
    }
    if (saving || !access.obclusterwrite) return;
    setSaving(true);
    setSubmitError('');
    try {
      values = normalizeStorageMode(structuredClone(values));
      values.clusterId = new Date().getTime() % 4294901759;
      const rootPassword = encryptText(values.rootPassword, publicKey);
      const proxyroPassword = values.proxyroPassword
        ? encryptText(values.proxyroPassword, publicKey)
        : undefined;
      if (!rootPassword || proxyroPassword === false)
        throw new Error(
          L(
            '密码加密失败，请刷新页面后重试',
            'Password encryption failed; reload and retry',
          ),
        );
      values.rootPassword = rootPassword;
      values.proxyroPassword = proxyroPassword;
      values.deletionProtection = deleteValue;
      values.pvcIndependent = pvcValue;

      strTrim(values);
      const topologyValue = values.topology.map((item) => ({
        ...item,
        nodeSelector: undefined,
        affinities: [...(item.affinities || []), ...(item.nodeSelector || [])],
      }));

      values.topology = topologyValue;

      const res = await createClusterReportWrap(values);
      if (res.successful) {
        message.success(res.message, 3);
        form.resetFields();
        setPasswordVal('');
        setProxyroPasswordVal('');
        setPvcValue(false);
        setDeleteValue(true);
        history.back();
      } else setSubmitError(res.message || L('创建失败', 'Creation failed'));
    } catch (e) {
      setSubmitError(errorText(e));
    } finally {
      setSaving(false);
    }
  };
  const initialValues = {
    deploymentMode:
      searchParams.get('deploymentMode') === 'shared_storage'
        ? 'shared_storage'
        : 'normal',
    mode: 'NORMAL',
    topology: [
      {
        zone: 'zone1',
        replicas: 1,
      },
      {
        zone: 'zone2',
        replicas: 1,
      },
      {
        zone: 'zone3',
        replicas: 1,
      },
    ],
  };

  useEffect(() => {
    fetchStorageClasses();
  }, []);
  return (
    <PageContainer
      header={{
        title: intl.formatMessage({
          id: 'dashboard.Cluster.New.CreateACluster',
          defaultMessage: '创建集群',
        }),
        onBack: () => {
          navigate('/cluster');
        },
      }}
      footer={[
        <Button
          disabled={saving}
          onClick={() => navigate('/cluster')}
          key="cancel"
        >
          {intl.formatMessage({
            id: 'dashboard.Cluster.New.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        ...(isSS && step > 0
          ? [
              <Button
                key="previous"
                disabled={saving}
                onClick={() => setStep((s) => s - 1)}
              >
                {L('上一步', 'Previous')}
              </Button>,
            ]
          : []),
        <Button
          type="primary"
          key="submit"
          loading={saving}
          disabled={!access.obclusterwrite}
          onClick={() => (isSS && step < 3 ? next() : form.submit())}
        >
          {isSS
            ? step < 3
              ? L('下一步', 'Next')
              : L('确认创建湖库集群', 'Create lakehouse cluster')
            : intl.formatMessage({
                id: 'dashboard.Cluster.New.Submit',
                defaultMessage: '提交',
              })}
        </Button>,
      ]}
    >
      <Form
        form={form}
        onFinish={onFinish}
        onFinishFailed={() => {
          if (isSS) setStep(0);
          message.warning(
            L(
              '配置校验失败，请检查各步骤',
              'Validation failed; check each step',
            ),
          );
        }}
        disabled={saving || !access.obclusterwrite}
        layout="vertical"
        initialValues={initialValues}
        style={{ marginBottom: 56 }}
      >
        {submitError && (
          <Alert
            type="error"
            showIcon
            message={submitError}
            style={{ marginBottom: 16 }}
          />
        )}
        <StorageArchitecture />
        {isSS && (
          <Steps
            current={step}
            style={{ marginBottom: 24 }}
            items={[
              L('基本信息与数据存储', 'Basics & data storage'),
              L('日志服务', 'Log service'),
              L('计算与缓存', 'Compute & cache'),
              L('检查并创建', 'Review & create'),
            ].map((title) => ({ title }))}
          />
        )}
        <div style={{ display: !isSS || step === 0 ? 'block' : 'none' }}>
          <Row gutter={[16, 16]}>
            <Col span={24}>
              <BasicInfo
                passwordVal={passwordVal}
                proxyroPasswordVal={proxyroPasswordVal}
                deleteValue={deleteValue}
                setPasswordVal={setPasswordVal}
                setProxyroPasswordVal={setProxyroPasswordVal}
                setDeleteValue={setDeleteValue}
                form={form}
              />
            </Col>
            {isSS && <SharedStorage form={form} section="storage" />}
          </Row>
        </div>
        {isSS && (
          <div style={{ display: step === 1 ? 'block' : 'none' }}>
            <Row gutter={[16, 16]}>
              <SharedStorage form={form} section="logservice" />
            </Row>
          </div>
        )}
        <div
          style={{
            display: !isSS || step === 2 ? 'block' : 'none',
            marginTop: isSS ? 0 : 16,
          }}
        >
          <Row gutter={[16, 16]}>
            {isSS && (
              <Col span={24}>
                <Alert
                  type="info"
                  showIcon
                  message={L(
                    'SS 的 data 卷用作本地数据缓存，log 卷为运行日志；远端数据存储与 LogService 日志存储单独配置，不创建本地 redoLog 卷。',
                    'SS uses the data volume as local cache and the log volume for runtime logs. Remote data and LogService storage are separate; no local redoLog volume is created.',
                  )}
                />
              </Col>
            )}
            <Topo form={form} />
            <Observer
              storageClasses={storageClasses}
              form={form}
              pvcValue={pvcValue}
              setPvcValue={setPvcValue}
            />
            <Monitor />
            <Parameters />
            <BackUp />
          </Row>
        </div>
        {isSS && (
          <div style={{ display: step === 3 ? 'block' : 'none' }}>
            <LakehouseReview form={form} />
          </div>
        )}
      </Form>
    </PageContainer>
  );
}
