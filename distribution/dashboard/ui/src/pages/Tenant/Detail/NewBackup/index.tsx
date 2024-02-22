import { WEEK_TEXT_MAP } from '@/constants/schedule';
import { usePublicKey } from '@/hook/usePublicKey';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { createBackupPolicyOfTenant, getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest, useUpdateEffect } from 'ahooks';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
  Space,
  Switch,
  TimePicker,
  message,
} from 'antd';
import { clone } from 'lodash';
import { useState } from 'react';
import { formatNewBackupForm } from '../../helper';
import BasicInfo from '../Overview/BasicInfo';
import type { ParamsType } from './ScheduleSelectComp';
import ScheduleSelectComp from './ScheduleSelectComp';
import styles from './index.less';
const { Password } = Input;
export default function NewBackup() {
  const navigate = useNavigate();
  const [ns, name] = getNSName();
  const [form] = Form.useForm();
  const [isExpand, setIsExpand] = useState(false);
  const [passInputExpand, setPassInputExpand] = useState(false);
  const [jobInputExpand, setJobInputExpand] = useState(false);
  const [recoveryInputExpand, setRecoveryInputExpand] = useState(false);
  const publicKey = usePublicKey();
  const scheduleValue = Form.useWatch(['scheduleDates'], form);
  const distType = [
    { label: 'OSS', value: 'OSS' },
    { label: 'NFS', value: 'NFS' },
  ];

  const handleSubmit = async (values: any) => {
    const res = await createBackupPolicyOfTenant({
      ns,
      name,
      ...formatNewBackupForm(values, publicKey),
    });
    if (res.successful) {
      message.success('创建成功', 3);
      form.resetFields();
      history.back();
    }
  };

  const initialValues = {
    scheduleDates: {
      mode: 'Weekly',
      days: [],
    },
  };
  const { data: tenantDetailResponse } = useRequest(getTenant, {
    defaultParams: [{ ns, name }],
  });

  const tenantDetail = tenantDetailResponse?.data;

  /**
   * When the scheduling cycle changes,
   * ensure that the backup data method
   * can be changed accordingly.
   */
  useUpdateEffect(() => {
    let newScheduleValue = clone(scheduleValue);
    scheduleValue.days.forEach((key: number) => {
      if (!scheduleValue[String(key)]) {
        newScheduleValue[key] = 'Full';
        form.setFieldValue('scheduleDates', newScheduleValue);
      }
    });
    Object.keys(scheduleValue).forEach((key) => {
      if (/^[\d]+$/.test(key) && !scheduleValue?.days.includes(Number(key))) {
        delete newScheduleValue[key];
        form.setFieldValue('scheduleDates', newScheduleValue);
      }
    });
  }, [scheduleValue]);

  return (
    <PageContainer
      style={{ paddingBottom: 70 }}
      header={{
        title: intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.CreateATenantBackupPolicy',
          defaultMessage: '创建租户备份策略',
        }),
        onBack: () => navigate(`/tenant/ns=${ns}&nm=${name}/backup`),
      }}
      footer={[
        <Button
          onClick={() => navigate(`/tenant/ns=${ns}&nm=${name}/backup`)}
          key="cancel"
        >
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button key="submit" onClick={() => form.submit()}>
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.Submit',
            defaultMessage: '提交',
          })}
        </Button>,
      ]}
    >
      {tenantDetail && (
        <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
      )}

      <Form initialValues={initialValues} form={form} onFinish={handleSubmit}>
        <Card style={{ marginBottom: 24 }}>
          <Row>
            <Col span={24}>
              <Space direction="vertical">
                <h3>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.BackupPolicyConfiguration',
                    defaultMessage: '备份策略配置',
                  })}
                </h3>
                <Row gutter={[16, 32]}>
                  <Col span={8}>
                    <Form.Item
                      name={['destType']}
                      label={intl.formatMessage({
                        id: 'Dashboard.Detail.NewBackup.BackupType',
                        defaultMessage: '备份类型',
                      })}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.SelectABackupType',
                            defaultMessage: '请选择备份类型',
                          }),
                        },
                      ]}
                    >
                      <Select
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseSelect',
                          defaultMessage: '请选择',
                        })}
                        options={distType}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      label="OSS AccessID"
                      name={['ossAccessId']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.EnterOssAccessid',
                            defaultMessage: '请输入 OSS AccessID',
                          }),
                        },
                      ]}
                    >
                      <Password
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      label="OSS AccessKey"
                      name={['ossAccessKey']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.EnterOssAccesskey',
                            defaultMessage: '请输入 OSS AccessKey',
                          }),
                        },
                      ]}
                    >
                      <Password
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      label={intl.formatMessage({
                        id: 'Dashboard.Detail.NewBackup.LogArchivePath',
                        defaultMessage: '日志归档路径',
                      })}
                      name={['archivePath']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.EnterTheLogArchivePath',
                            defaultMessage: '请输入日志归档路径',
                          }),
                        },
                      ]}
                    >
                      <Input
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      label={intl.formatMessage({
                        id: 'Dashboard.Detail.NewBackup.DataBackupPath',
                        defaultMessage: '数据备份路径',
                      })}
                      name={['bakDataPath']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.EnterTheDataBackupPath',
                            defaultMessage: '请输入数据备份路径',
                          }),
                        },
                      ]}
                    >
                      <Input
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              </Space>
            </Col>
            <Col span={12}>
              <Space direction="vertical">
                <h3>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.SchedulingCycle',
                    defaultMessage: '调度周期',
                  })}
                </h3>

                <Form.Item
                  rules={[
                    () => ({
                      validator: (_: any, value: ParamsType) => {
                        if (!value.days.length) {
                          return Promise.reject(
                            new Error(
                              intl.formatMessage({
                                id: 'Dashboard.Detail.NewBackup.SelectASchedulingCycle',
                                defaultMessage: '请选择调度周期',
                              }),
                            ),
                          );
                        }
                        return Promise.resolve();
                      },
                    }),
                  ]}
                  name={['scheduleDates']}
                >
                  <ScheduleSelectComp />
                </Form.Item>
              </Space>
            </Col>
            <Col span={12}>
              <Space direction="vertical">
                <h3>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.SchedulingTime',
                    defaultMessage: '调度时间',
                  })}
                </h3>
                <Form.Item
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'Dashboard.Detail.NewBackup.SelectASchedulingTime',
                        defaultMessage: '请选择调度时间',
                      }),
                    },
                  ]}
                  name={'scheduleTime'}
                >
                  <TimePicker />
                </Form.Item>
              </Space>
            </Col>
            <Col span={24}>
              <Space direction="vertical">
                <h3>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.BackupData',
                    defaultMessage: '备份数据方式',
                  })}
                </h3>
                <p>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.WeRecommendThatYouConfigure',
                    defaultMessage: '建议至少配置 1 个全量备份',
                  })}
                </p>
                {scheduleValue?.days
                  .sort((pre: number, cur: number) => pre - cur)
                  .map((day: number, index: number) => (
                    <Form.Item
                      name={['scheduleDates', day]}
                      label={
                        scheduleValue?.mode === 'Monthly'
                          ? day
                          : WEEK_TEXT_MAP.get(day)
                      }
                      key={index}
                    >
                      <Radio.Group defaultValue="Full">
                        <Radio value="Full">
                          {intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.FullQuantity',
                            defaultMessage: '全量',
                          })}
                        </Radio>
                        <Radio value="Incremental">
                          {intl.formatMessage({
                            id: 'Dashboard.Detail.NewBackup.Increment',
                            defaultMessage: '增量',
                          })}
                        </Radio>
                      </Radio.Group>
                    </Form.Item>
                  ))}
              </Space>
            </Col>
          </Row>
        </Card>
        <Card
          title={
            <Space>
              <span>
                {intl.formatMessage({
                  id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration',
                  defaultMessage: '高级配置',
                })}
              </span>
              <Switch
                onChange={() => setIsExpand(!isExpand)}
                checked={isExpand}
              />
            </Space>
          }
          bordered={false}
        >
          {isExpand && (
            <Row>
              <Col className={styles.column} span={24}>
                <label className={styles.labelText}>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.EncryptedBackup',
                    defaultMessage: '加密备份：',
                  })}
                </label>
                <Switch
                  className={styles.switch}
                  onChange={() => setPassInputExpand(!passInputExpand)}
                  checked={passInputExpand}
                />

                {passInputExpand && (
                  <Form.Item
                    label={intl.formatMessage({
                      id: 'Dashboard.Detail.NewBackup.EncryptedPassword',
                      defaultMessage: '加密密码',
                    })}
                    rules={[
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.EnterAnEncryptionPassword',
                          defaultMessage: '请输入加密密码',
                        }),
                      },
                    ]}
                    name={['bakEncryptionPassword']}
                  >
                    <Password
                      style={{ width: 216 }}
                      placeholder={intl.formatMessage({
                        id: 'Dashboard.Detail.NewBackup.PleaseEnter',
                        defaultMessage: '请输入',
                      })}
                    />
                  </Form.Item>
                )}
              </Col>
              <Col className={styles.column} span={24}>
                <label className={styles.labelText}>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.BackupTaskRetention',
                    defaultMessage: '备份任务保留：',
                  })}
                </label>
                <Switch
                  className={styles.switch}
                  onChange={() => setJobInputExpand(!jobInputExpand)}
                  checked={jobInputExpand}
                />

                {jobInputExpand && (
                  <Form.Item
                    rules={[
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.PleaseEnterTheRetentionDays',
                          defaultMessage: '请输入备份任务保留天数',
                        }),
                      },
                    ]}
                    name={['jobKeepWindow']}
                  >
                    <InputNumber />
                  </Form.Item>
                )}
              </Col>
              <Col className={styles.column} span={24}>
                <label className={styles.labelText}>
                  {intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.DataRecoveryWindow',
                    defaultMessage: '数据恢复窗口：',
                  })}
                </label>
                <Switch
                  className={styles.switch}
                  onChange={() => setRecoveryInputExpand(!recoveryInputExpand)}
                  checked={recoveryInputExpand}
                />

                {recoveryInputExpand && (
                  <Form.Item
                    rules={[
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'Dashboard.Detail.NewBackup.EnterTheDataRecoveryWindow',
                          defaultMessage: '请输入数据恢复窗口：',
                        }),
                      },
                    ]}
                    name={['recoveryWindow']}
                  >
                    <InputNumber />
                  </Form.Item>
                )}
              </Col>
              <Form.Item
                name={['pieceInterval']}
                label={intl.formatMessage({
                  id: 'Dashboard.Detail.NewBackup.ArchiveSliceInterval',
                  defaultMessage: '归档切片间隔',
                })}
              >
                <InputNumber min={1} max={7} />
              </Form.Item>
            </Row>
          )}
        </Card>
      </Form>
    </PageContainer>
  );
}
