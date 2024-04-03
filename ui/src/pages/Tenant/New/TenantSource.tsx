import { getAllTenants } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import {
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  Row,
  Select,
  Space,
  Switch,
  TimePicker,
  Tooltip,
} from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import { FormInstance } from 'antd/lib/form';
import dayjs from 'dayjs';
import moment from 'moment';
import { useEffect, useState } from 'react';
import styles from './index.less';

interface TenantSourceProps {
  form: FormInstance<any>;
  clusterName: string;
}
type RoleType = 'PRIMARY' | 'STANDBY';
const { Password } = Input;
export default function TenantSource({ form, clusterName }: TenantSourceProps) {
  const [isChecked, setIsChecked] = useState<boolean>(false);
  const [recoverChecked, setRecoverChecked] = useState<boolean>(false);
  const [synchronizeChecked, setSynchronizeChecked] = useState<boolean>(false);
  const [recoverTimeChecked, setRecoverTimeChecked] = useState<boolean>(false);
  const [selectRole, setSelectRole] = useState<RoleType | undefined>();
  const subTitleStyle = {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 24,
  };
  const h3Style = { marginBottom: 0, marginRight: 12 };
  const tenantRole = [
    {
      label: 'PRIMARY',
      value: 'PRIMARY',
    },
    {
      label: 'STANDBY',
      value: 'STANDBY',
    },
  ];

  const recoverType = [
    {
      label: 'NFS',
      value: 'NFS',
    },
    {
      label: 'OSS',
      value: 'OSS',
    },
  ];

  const tenantRoleChange = (value: RoleType) => {
    setSelectRole(value);
    if (value === 'PRIMARY' && synchronizeChecked) {
      setSynchronizeChecked(false);
    }
  };

  const { run: getTenants, data: tenantListRes } = useRequest(getAllTenants);

  const tenantList = tenantListRes?.data
    .filter((item) => item.status === 'running' && item.tenantRole === 'PRIMARY')
    .map((tenant) => ({
      label: tenant.name,
      value: tenant.name,
    }));

  const range = (start: number, end: number) => {
    const result = [];
    for (let i = start; i < end; i++) {
      result.push(i);
    }
    return result;
  };

  const disabledDateTime: RangePickerProps['disabledTime'] = (_) => {
    const isToday = _?.date() === moment().date();
    if (!isToday)
      return {
        disabledHours: () => [],
        disabledMinutes: () => [],
        disabledSeconds: () => [],
      };
    return {
      disabledHours: () => range(0, 24).splice(moment().hour() + 1, 24),
      disabledMinutes: (hour) => {
        if (hour === moment().hour()) {
          return range(0, 60).splice(moment().minute() + 1, 60);
        }
        return [];
      },
      disabledSeconds: (hour, minute) => {
        if (hour === moment().hour() && minute === moment().minute()) {
          return range(0, 60).splice(moment().second(), 60);
        }
        return [];
      },
    };
  };

  const disabledDate: RangePickerProps['disabledDate'] = (current) => {
    return current && current > dayjs().endOf('day');
  };

  useEffect(() => {
    if (synchronizeChecked && clusterName) {
      getTenants(clusterName);
    }
  }, [synchronizeChecked, clusterName]);

  useEffect(() => {
    let [cardBody] = document.querySelectorAll(
      '#tenant-card .ant-card-body',
    ) as NodeListOf<HTMLElement>;
    if (!isChecked && cardBody) {
      cardBody.style.padding = '0px';
    } else {
      cardBody.style.padding = '24px';
    }
  }, [isChecked]);

  return (
    <Card
      title={
        <Space>
          <span>
            {intl.formatMessage({
              id: 'Dashboard.Tenant.New.TenantSource.TenantRecovery',
              defaultMessage: '租户恢复',
            })}
          </span>
          <Switch onChange={(val) => setIsChecked(val)} checked={isChecked} />
        </Space>
      }
      bordered={false}
      id="tenant-card"
    >
      {isChecked && (
        <>
          <Form.Item
            name={['tenantRole']}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.TenantSource.TenantRole',
              defaultMessage: '租户角色',
            })}
            style={{ width: '50%' }}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.TenantSource.SelectATenantRole',
                  defaultMessage: '请选择租户角色',
                }),
              },
            ]}
          >
            <Select
              placeholder={intl.formatMessage({
                id: 'Dashboard.Tenant.New.TenantSource.PleaseSelect',
                defaultMessage: '请选择',
              })}
              onChange={(value) => tenantRoleChange(value)}
              options={tenantRole}
            />
          </Form.Item>
          <div style={{ marginBottom: 24 }}>
            <div style={subTitleStyle}>
              <h3 style={h3Style}>
                {intl.formatMessage({
                  id: 'Dashboard.Tenant.New.TenantSource.RestoreFromBackup',
                  defaultMessage: '从备份恢复',
                })}
              </h3>
              <Switch
                onChange={(val) => setRecoverChecked(val)}
                checked={recoverChecked}
              />
            </div>

            {recoverChecked && (
              <div>
                <Row gutter={[16, 32]}>
                  <Col span={8}>
                    <Form.Item
                      name={['source', 'restore', 'type']}
                      label={intl.formatMessage({
                        id: 'Dashboard.Tenant.New.TenantSource.RecoveryType',
                        defaultMessage: '恢复类型',
                      })}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'Dashboard.Tenant.New.TenantSource.SelectARecoveryType',
                            defaultMessage: '请选择恢复类型',
                          }),
                        },
                      ]}
                    >
                      <Select
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Tenant.New.TenantSource.PleaseSelect',
                          defaultMessage: '请选择',
                        })}
                        options={recoverType}
                      />
                    </Form.Item>
                  </Col>
                  <Form.Item
                    noStyle
                    dependencies={[['source', 'restore', 'type']]}
                  >
                    {({ getFieldValue }) => {
                      if (
                        getFieldValue(['source', 'restore', 'type']) !== 'NFS'
                      ) {
                        return (
                          <>
                            <Col span={8}>
                              <Form.Item
                                label="OSS AccessID"
                                name={['source', 'restore', 'ossAccessId']}
                                rules={[
                                  {
                                    required: true,
                                    message: intl.formatMessage({
                                      id: 'Dashboard.Tenant.New.TenantSource.EnterOssAccessid',
                                      defaultMessage: '请输入 OSS AccessID',
                                    }),
                                  },
                                ]}
                              >
                                <Password
                                  placeholder={intl.formatMessage({
                                    id: 'Dashboard.Tenant.New.TenantSource.PleaseEnter',
                                    defaultMessage: '请输入',
                                  })}
                                />
                              </Form.Item>
                            </Col>
                            <Col span={8}>
                              <Form.Item
                                label="OSS AccessKey"
                                name={['source', 'restore', 'ossAccessKey']}
                                rules={[
                                  {
                                    required: true,
                                    message: intl.formatMessage({
                                      id: 'Dashboard.Tenant.New.TenantSource.EnterOssAccesskey',
                                      defaultMessage: '请输入 OSS AccessKey',
                                    }),
                                  },
                                ]}
                              >
                                <Password
                                  placeholder={intl.formatMessage({
                                    id: 'Dashboard.Tenant.New.TenantSource.PleaseEnter',
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
                    <Tooltip
                      overlayStyle={{ maxWidth: 300 }}
                      title={
                        <div>
                          <span>
                            {intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.NfsTypeSeePathFormat',
                              defaultMessage:
                                'NFS 类型，参考路径格式：path/to/dir',
                            })}
                          </span>
                          <br />
                          <span>
                            {intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.OssTypeForMoreInformation',
                              defaultMessage:
                                'OSS 类型，参考路径格式：oss://bucket/dir?host=xxx',
                            })}
                          </span>
                        </div>
                      }
                    >
                      <Form.Item
                        label={intl.formatMessage({
                          id: 'Dashboard.Tenant.New.TenantSource.LogArchivePath',
                          defaultMessage: '日志归档路径',
                        })}
                        name={['source', 'restore', 'archiveSource']}
                        rules={[
                          {
                            required: true,
                            message: intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.EnterTheLogArchivePath',
                              defaultMessage: '请输入日志归档路径',
                            }),
                          },
                        ]}
                      >
                        <Input
                          placeholder={intl.formatMessage({
                            id: 'Dashboard.Tenant.New.TenantSource.PleaseEnter',
                            defaultMessage: '请输入',
                          })}
                        />
                      </Form.Item>
                    </Tooltip>
                  </Col>
                  <Col span={8}>
                    <Tooltip
                      overlayStyle={{ maxWidth: 300 }}
                      title={
                        <div>
                          <span>
                            {intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.NfsTypeSeePathFormat',
                              defaultMessage:
                                'NFS 类型，参考路径格式：path/to/dir',
                            })}
                          </span>
                          <br />
                          <span>
                            {intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.OssTypeForMoreInformation',
                              defaultMessage:
                                'OSS 类型，参考路径格式：oss://bucket/dir?host=xxx',
                            })}
                          </span>
                        </div>
                      }
                    >
                      <Form.Item
                        label={intl.formatMessage({
                          id: 'Dashboard.Tenant.New.TenantSource.DataBackupPath',
                          defaultMessage: '数据备份路径',
                        })}
                        name={['source', 'restore', 'bakDataSource']}
                        rules={[
                          {
                            required: true,
                            message: intl.formatMessage({
                              id: 'Dashboard.Tenant.New.TenantSource.EnterTheDataBackupPath',
                              defaultMessage: '请输入数据备份路径',
                            }),
                          },
                        ]}
                      >
                        <Input
                          placeholder={intl.formatMessage({
                            id: 'Dashboard.Tenant.New.TenantSource.PleaseEnter',
                            defaultMessage: '请输入',
                          })}
                        />
                      </Form.Item>
                    </Tooltip>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      label={intl.formatMessage({
                        id: 'Dashboard.Tenant.New.TenantSource.EncryptedPassword',
                        defaultMessage: '加密密码',
                      })}
                      name={['source', 'restore', 'bakEncryptionPassword']}
                    >
                      <Password
                        placeholder={intl.formatMessage({
                          id: 'Dashboard.Tenant.New.TenantSource.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    {/* <PasswordInput
                  value={passwordVal}
                  onChange={setPasswordVal}
                  form={form}
                  name="rootPassword"
                  /> */}
                  </Col>
                </Row>
                <div>
                  {' '}
                  <div style={subTitleStyle}>
                    <h4 style={h3Style}>
                      {intl.formatMessage({
                        id: 'Dashboard.Tenant.New.TenantSource.RestoreToASpecificTime',
                        defaultMessage: '恢复至特定时间',
                      })}
                    </h4>{' '}
                    <Switch
                      onChange={(val) => setRecoverTimeChecked(val)}
                      checked={recoverTimeChecked}
                      disabled={synchronizeChecked === true}
                    />
                  </div>
                  {recoverTimeChecked && (
                    <Row gutter={[16, 32]}>
                      <Col span={12}>
                        <Form.Item
                          className={styles.dateContainer}
                          label={intl.formatMessage({
                            id: 'Dashboard.Tenant.New.TenantSource.RecoveryDate',
                            defaultMessage: '恢复日期',
                          })}
                          style={{ width: '100%' }}
                          rules={[
                            {
                              required: true,
                              message: intl.formatMessage({
                                id: 'Dashboard.Tenant.New.TenantSource.SelectARecoveryDate',
                                defaultMessage: '请选择恢复日期',
                              }),
                            },
                          ]}
                          name={['source', 'restore', 'until', 'date']}
                        >
                          <DatePicker disabledDate={disabledDate} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          className={styles.dateContainer}
                          style={{ width: '100%' }}
                          rules={[
                            {
                              required: true,
                              message: intl.formatMessage({
                                id: 'Dashboard.Tenant.New.TenantSource.SelectARecoveryTime',
                                defaultMessage: '请选择恢复时间',
                              }),
                            },
                          ]}
                          label={intl.formatMessage({
                            id: 'Dashboard.Tenant.New.TenantSource.MinutesAndSeconds',
                            defaultMessage: '时分秒',
                          })}
                          name={['source', 'restore', 'until', 'time']}
                        >
                          <TimePicker disabledTime={disabledDateTime} />
                        </Form.Item>
                      </Col>
                    </Row>
                  )}
                </div>
              </div>
            )}
          </div>
          <div>
            <div style={subTitleStyle}>
              <h3 style={h3Style}>
                {intl.formatMessage({
                  id: 'Dashboard.Tenant.New.TenantSource.SynchronizePrimaryTenants',
                  defaultMessage: '同步主租户',
                })}
              </h3>
              <Switch
                onChange={(val) => setSynchronizeChecked(val)}
                checked={synchronizeChecked}
                disabled={
                  recoverTimeChecked === true || selectRole === 'PRIMARY'
                }
              />
            </div>
            {synchronizeChecked && (
              <Form.Item
                style={{ width: '50%' }}
                label={intl.formatMessage({
                  id: 'Dashboard.Tenant.New.TenantSource.MasterTenant',
                  defaultMessage: '主租户',
                })}
                name={['source', 'tenant']}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.TenantSource.PleaseSelect',
                      defaultMessage: '请选择',
                    }),
                  },
                ]}
              >
                <Select
                  options={tenantList}
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Tenant.New.TenantSource.PleaseSelect',
                    defaultMessage: '请选择',
                  })}
                />
              </Form.Item>
            )}
          </div>
        </>
      )}
    </Card>
  );
}
