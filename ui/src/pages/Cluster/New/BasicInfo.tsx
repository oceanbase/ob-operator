import { intl } from '@/utils/intl';
import { PlusOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { Card, Col, Divider, Form, Input, Row, Select, Tooltip } from 'antd';
import type { FormInstance } from 'antd/lib/form';

import PasswordInput from '@/components/PasswordInput';
import AddNSModal from '@/components/customModal/AddNSModal';
import { MODE_MAP } from '@/constants';
import { resourceNameRule } from '@/constants/rules';
import { getNameSpaces } from '@/services';
import { useState } from 'react';
import styles from './index.less';

interface BasicInfoProps {
  form: FormInstance<API.CreateClusterData>;
  passwordVal: string;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}

export default function BasicInfo({
  form,
  passwordVal,
  setPasswordVal,
}: BasicInfoProps) {
  // control the modal for adding a new namespace
  const [visible, setVisible] = useState(false);
  const { data, run: getNS } = useRequest(getNameSpaces);

  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase());

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

  const addNSCallback = (newNS: string) => {
    form.setFieldValue('namespace', newNS);
    form.validateFields(['namespace']);
    getNS();
  };

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
            <PasswordInput
              value={passwordVal}
              onChange={setPasswordVal}
              form={form}
              name="rootPassword"
            />
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
                  {
                    pattern: /\D/,
                    message: intl.formatMessage({
                      id: 'Dashboard.Cluster.New.BasicInfo.ResourceNamesCannotUsePure',
                      defaultMessage: '资源名不能使用纯数字',
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
          <Col span={8}>
            <Form.Item
              label={intl.formatMessage({
                id: 'Dashboard.Cluster.New.BasicInfo.ClusterMode',
                defaultMessage: '集群模式',
              })}
              name="mode"
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.Cluster.New.BasicInfo.SelectClusterMode',
                    defaultMessage: '请选择集群模式',
                  }),
                },
              ]}
            >
              <Select
                placeholder={intl.formatMessage({
                  id: 'Dashboard.Cluster.New.BasicInfo.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                optionLabelProp="selectLabel"
                options={Array.from(MODE_MAP.keys()).map((key) => ({
                  value: key,
                  selectLabel:MODE_MAP.get(key)?.text,
                  label: (
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <span>{MODE_MAP.get(key)?.text}</span>
                      <span>{MODE_MAP.get(key)?.limit}</span>
                    </div>
                  ),
                }))}
              />
            </Form.Item>
          </Col>
        </Row>
      </Card>
      <AddNSModal
        visible={visible}
        setVisible={setVisible}
        successCallback={addNSCallback}
      />
    </Col>
  );
}
