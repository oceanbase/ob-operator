import { intl } from '@/utils/intl';
import { PlusOutlined } from '@ant-design/icons';
import { useRequest, useUpdateEffect } from 'ahooks';
import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  Row,
  Select,
  Tooltip,
  message,
} from 'antd';
import type { FormInstance } from 'antd/lib/form';

import AddNSModal from '@/components/customModal/AddNSModal';
import { getNameSpaces } from '@/services';
import copy from 'copy-to-clipboard';
import { useState } from 'react';
import {
  generateRandomPassword,
  passwordRules,
  resourceNameRule,
} from './helper';
import styles from './index.less';

interface BasicInfoProps {
  form: FormInstance<any>;
  passwordVal: string;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}

export default function BasicInfo({
  form,
  passwordVal,
  setPasswordVal,
}: BasicInfoProps) {
  const { setFieldValue } = form;
  //控制新增命名空间弹窗
  const [visible, setVisible] = useState(false);
  const [textVisile, setTextVisible] = useState<boolean>(false);
  const { data, run: getNS } = useRequest(getNameSpaces);
  const genaretaPassword = () => {
    let password = generateRandomPassword();
    setPasswordVal(password);
    setFieldValue('rootPassword', password);
    form.validateFields(['rootPassword']);
  };
  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase());

  const passwordCopy = () => {
    if (passwordVal) {
      copy(passwordVal);
      message.success(
        intl.formatMessage({
          id: 'OBDashboard.Cluster.New.BasicInfo.CopiedSuccessfully',
          defaultMessage: '复制成功',
        }),
      );
    }
  };

  const DropDownComponent = (menu: any) => {
    return (
      <div>
        {menu}
        <Divider style={{ margin: '10px 0' }} />
        <div
          onClick={() => setVisible(true)}
          style={{ padding: '8px', cursor: 'pointer' }}
        >
          <PlusOutlined />
          <span style={{ marginLeft: '6px' }}>
            {intl.formatMessage({
              id: 'OBDashboard.Cluster.New.BasicInfo.AddNamespace',
              defaultMessage: '新增命名空间',
            })}
          </span>
        </div>
      </div>
    );
  };

  const passwordChange = async (value: string) => {
    setPasswordVal(value);
  };

  const listenPasswordChange = async () => {
    try {
      await form.validateFields(['rootPassword']);
      setTextVisible(true);
    } catch (err: any) {
      const { errorFields } = err;
      if (errorFields[0].errors.length) setTextVisible(false);
    }
  };

  useUpdateEffect(() => {
    listenPasswordChange();
  }, [passwordVal]);

  return (
    <Col span={24}>
      <Card
        title={intl.formatMessage({
          id: 'dashboard.Cluster.New.BasicInfo.BasicInformation',
          defaultMessage: '基本信息',
        })}
      >
        <Row gutter={[16, 32]}>
          <Col span={8} style={{ height: 48 }}>
            <Form.Item
              label={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.BasicInfo.Namespace',
                defaultMessage: '命名空间',
              })}
              name="namespace"
              validateTrigger="onBlur"
              validateFirst
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BasicInfo.EnterANamespace',
                    defaultMessage: '请输入命名空间',
                  }),
                  validateTrigger: 'onChange',
                },
                resourceNameRule,
              ]}
            >
              <Select
                showSearch
                placeholder={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                optionFilterProp="label"
                filterOption={filterOption}
                dropdownRender={DropDownComponent}
                options={data}
              />
            </Form.Item>
          </Col>
          <Col span={8} style={{ height: 48 }}>
            <Tooltip
              color="#fff"
              overlayInnerStyle={{ color: 'rgba(0,0,0,.85)' }}
              overlayClassName={styles.toolTipContent}
              placement="bottomLeft"
              title={
                <ul>
                  <li>
                    {intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BasicInfo.ToCharactersInLength',
                      defaultMessage: '长度为 8~32 个字符',
                    })}
                  </li>
                  <li>
                    {intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BasicInfo.CanOnlyContainLettersNumbers',
                      defaultMessage:
                        '只能包含字母、数字和特殊字符（~!@#%^&*_-+=|()',
                    })}
                    {}[]:;,.?/）
                  </li>
                  <li>
                    {intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BasicInfo.AtLeastUppercaseAndLowercase',
                      defaultMessage:
                        '大小写字母、数字和特殊字符都至少包含 2 个',
                    })}
                  </li>
                </ul>
              }
            >
              <Form.Item
                label={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.Password',
                  defaultMessage: '密码',
                })}
                name="rootPassword"
                rules={passwordRules}
                className={styles.passwordFormItem}
                validateFirst
              >
                <Row gutter={8}>
                  <Col style={{ flex: 1 }}>
                    <Input.Password
                      value={passwordVal}
                      onChange={(val) => passwordChange(val.target.value)}
                      placeholder={intl.formatMessage({
                        id: 'OBDashboard.Cluster.New.BasicInfo.PleaseEnterOrGenerateRandomly',
                        defaultMessage: '请输入或随机生成',
                      })}
                    />

                    {textVisile && (
                      <p style={{ color: 'rgba(0, 0, 0, 0.45)' }}>
                        {intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.BasicInfo.PleaseRememberThePasswordOr',
                          defaultMessage: '请牢记密码，也可',
                        })}
                        <a onClick={passwordCopy}>
                          {intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.BasicInfo.CopyPassword',
                            defaultMessage: '复制密码',
                          })}
                        </a>
                        {intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.BasicInfo.AndKeepItProperly',
                          defaultMessage: '并妥善保存',
                        })}
                      </p>
                    )}
                  </Col>
                  <Col>
                    <Button onClick={genaretaPassword}>
                      {intl.formatMessage({
                        id: 'OBDashboard.Cluster.New.BasicInfo.RandomlyGenerated',
                        defaultMessage: '随机生成',
                      })}
                    </Button>
                  </Col>
                </Row>
              </Form.Item>
            </Tooltip>
          </Col>
          <Col span={8} style={{ height: 48 }}>
            <Tooltip
              color="#fff"
              overlayInnerStyle={{ color: 'rgba(0,0,0,.85)' }}
              overlayClassName={styles.toolTipContent}
              placement="topLeft"
              title={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.BasicInfo.TheNameOfTheResource',
                defaultMessage: 'k8s中资源的名称',
              })}
            >
              <Form.Item
                label={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.ResourceName',
                  defaultMessage: '资源名称',
                })}
                name="name"
                validateTrigger="onChange"
                validateFirst
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BasicInfo.EnterAKSResource',
                      defaultMessage: '请输入k8s资源名称',
                    }),
                  },
                  resourceNameRule,
                ]}
              >
                <Input
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BasicInfo.EnterAResourceName',
                    defaultMessage: '请输入资源名',
                  })}
                />
              </Form.Item>
            </Tooltip>
          </Col>

          <Col span={8} style={{ height: 72 }}>
            <Form.Item
              label={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.BasicInfo.ClusterName',
                defaultMessage: '集群名',
              })}
              name="clusterName"
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BasicInfo.EnterAClusterName',
                    defaultMessage: '请输入集群名',
                  }),
                },
              ]}
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.BasicInfo.EnterAClusterName',
                  defaultMessage: '请输入集群名',
                })}
              />
            </Form.Item>
          </Col>
        </Row>
      </Card>
      <AddNSModal
        visible={visible}
        setVisible={setVisible}
        successCallback={getNS}
      />
    </Col>
  );
}
