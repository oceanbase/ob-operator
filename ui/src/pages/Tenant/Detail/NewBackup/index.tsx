import { usePublicKey } from '@/hook/usePublicKey';
import { createBackupReportWrap } from '@/services/reportRequest/backupReportReq';
import { getTenant } from '@/services/tenant';
import { strTrim } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Col, Form, Input, Row, Select, message } from 'antd';
import { checkScheduleDatesHaveFull, formatBackupForm } from '../../helper';
import BasicInfo from '../Overview/BasicInfo';
import AdvancedConfiguration from './AdvancedConfiguration';
import BakMethodsList from './BakMethodsList';
import SchduleSelectFormItem from './SchduleSelectFormItem';
import ScheduleTimeFormItem from './ScheduleTimeFormItem';

export type ScheduleDates = {
  [T: number]: API.BackupType;
  days: number[];
  mode: API.ScheduleType;
};

const { Password } = Input;
export default function NewBackup() {
  const navigate = useNavigate();
  const { ns, name, tenantName } = useParams();
  const [form] = Form.useForm<API.NewBackupForm>();
  const publicKey = usePublicKey();
  const scheduleValue = Form.useWatch(['scheduleDates'], form);

  const distTypes = [
    { label: 'NFS', value: 'NFS' },
    { label: 'OSS', value: 'OSS' },
    { label: 'COS', value: 'COS' },
    { label: 'S3', value: 'S3' },
    { label: 'S3-Compatible', value: 'S3-Compatible' },
  ];

  const handleSubmit = async (values: API.NewBackupForm) => {
    if (!checkScheduleDatesHaveFull(values.scheduleDates)) {
      message.warning(
        intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.ConfigureAtLeastOneFull',
          defaultMessage: '请至少配置 1 个全量备份',
        }),
      );
      return;
    }
    const res = await createBackupReportWrap({
      ns: ns!,
      name: name!,
      ...formatBackupForm(strTrim(values), publicKey),
    });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.CreatedSuccessfully',
          defaultMessage: '创建成功',
        }),
        3,
      );
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
    defaultParams: [{ ns: ns!, name: name! }],
  });

  const tenantDetail = tenantDetailResponse?.data;

  return (
    <PageContainer
      style={{ paddingBottom: 70 }}
      header={{
        title: intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.CreateATenantBackupPolicy',
          defaultMessage: '创建租户备份策略',
        }),
        onBack: () => navigate(`/tenant/${ns}/${name}/${tenantName}/backup`),
      }}
      footer={[
        <Button
          onClick={() => navigate(`/tenant/${ns}/${name}/${tenantName}/backup`)}
          key="cancel"
        >
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button key="submit" type="primary" onClick={() => form.submit()}>
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.Submit',
            defaultMessage: '提交',
          })}
        </Button>,
      ]}
    >
      <Form initialValues={initialValues} form={form} onFinish={handleSubmit}>
        {tenantDetail && (
          <Col span={24}>
            <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
          </Col>
        )}

        <Card style={{ marginTop: 24, marginBottom: 24 }}>
          <Row>
            <Col span={24}>
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
                      options={distTypes}
                    />
                  </Form.Item>
                </Col>
                <Form.Item noStyle dependencies={['destType']}>
                  {({ getFieldValue }) => {
                    if (getFieldValue(['destType']) !== 'NFS') {
                      return (
                        <>
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
                              label="Host"
                              name={['host']}
                              rules={[
                                {
                                  required: true,
                                  message: intl.formatMessage({
                                    id: 'src.pages.Tenant.Detail.NewBackup.F6D664A0',
                                    defaultMessage: '请输入 Host',
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
                        </>
                      );
                    }
                    return null;
                  }}
                </Form.Item>

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

                <Form.Item noStyle dependencies={['destType']}>
                  {({ getFieldValue }) => {
                    if (getFieldValue(['destType']) === 'COS') {
                      return (
                        <Col span={8}>
                          <Form.Item
                            label="AppID"
                            name={['appID']}
                            rules={[
                              {
                                required: true,
                                message: intl.formatMessage({
                                  id: 'src.pages.Tenant.Detail.NewBackup.7A899F0F',
                                  defaultMessage: '请输入AppID',
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
                      );
                    }
                  }}
                </Form.Item>
                <Form.Item noStyle dependencies={['destType']}>
                  {({ getFieldValue }) => {
                    if (getFieldValue(['destType']) === 'S3') {
                      return (
                        <Col span={8}>
                          <Form.Item
                            label="Region"
                            name={['Region']}
                            rules={[
                              {
                                required: true,
                                message: intl.formatMessage({
                                  id: 'src.pages.Tenant.Detail.NewBackup.5A4FD22B',
                                  defaultMessage: '请输入 Region',
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
                      );
                    }
                  }}
                </Form.Item>
              </Row>
            </Col>
            <Col span={12}>
              <SchduleSelectFormItem
                form={form}
                scheduleValue={scheduleValue}
              />
            </Col>
            <Col span={12}>
              <ScheduleTimeFormItem />
            </Col>
            {scheduleValue && (
              <Col span={24}>
                <BakMethodsList scheduleValue={scheduleValue} />
              </Col>
            )}
          </Row>
        </Card>
        <AdvancedConfiguration />
      </Form>
    </PageContainer>
  );
}
