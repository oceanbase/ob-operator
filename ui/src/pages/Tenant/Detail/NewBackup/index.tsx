import { usePublicKey } from '@/hook/usePublicKey';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { createBackupPolicyOfTenant,getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import {
Button,
Card,
Col,
Form,
Input,
Row,
Select,
Space,
message,
} from 'antd';
import { formatBackupForm } from '../../helper';
import BasicInfo from '../Overview/BasicInfo';
import AdvancedConfiguration from './AdvancedConfiguration';
import BakMethodsList from './BakMethodsList';
import SchduleSelectFormItem from './SchduleSelectFormItem';
import ScheduleTimeFormItem from './ScheduleTimeFormItem';
const { Password } = Input;
export default function NewBackup() {
  const navigate = useNavigate();
  const [ns, name] = getNSName();
  const [form] = Form.useForm();
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
      ...formatBackupForm(values, publicKey),
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
    defaultParams: [{ ns, name }],
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
      <Form initialValues={initialValues} form={form} onFinish={handleSubmit}>
        {tenantDetail && (
          <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
        )}
        <Card style={{ marginTop: 24, marginBottom: 24 }}>
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
