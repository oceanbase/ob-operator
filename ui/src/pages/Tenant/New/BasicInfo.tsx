import InputNumber from '@/components/InputNumber';
import PasswordInput from '@/components/PasswordInput';
import { LOADTYPE_LIST, TZ_NAME_REG } from '@/constants';
import { resourceNameRule } from '@/constants/rules';
import { intl } from '@/utils/intl';
import { Card, Checkbox, Col, Form, Input, Row, Select, Space } from 'antd';
import type { FormInstance } from 'antd/lib/form';

const { Option } = Select;

interface BasicInfoProps {
  form: FormInstance<API.NewTenantForm>;
  passwordVal: string;
  clusterList: API.SimpleClusterList;
  setSelectClusterId: React.Dispatch<React.SetStateAction<number | undefined>>;
  setDeleteVal: boolean;
  deleteVal: boolean;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}
export default function BasicInfo({
  form,
  passwordVal,
  clusterList,
  deleteVal,
  setDeleteVal,
  setPasswordVal,
  setSelectClusterId,
}: BasicInfoProps) {
  const clusterOptions = clusterList
    .filter((cluster) => cluster.status !== 'failed')
    .map((cluster) => ({
      value: cluster.id,
      label: cluster.name,
      status: cluster.status,
    }));
  const selectClusterChange = (id: number) => {
    setSelectClusterId(id);
  };
  return (
    <Card
      title={intl.formatMessage({
        id: 'Dashboard.Tenant.New.BasicInfo.BasicInformation',
        defaultMessage: '基本信息',
      })}
    >
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <Form.Item
            name={['obcluster']}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.EnterAnObCluster',
                  defaultMessage: '请输入OB集群',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.BasicInfo.ObCluster',
              defaultMessage: 'OB集群',
            })}
          >
            <Select
              placeholder={intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.PleaseSelect',
                defaultMessage: '请选择',
              })}
              onChange={(value) => selectClusterChange(value)}
              optionLabelProp="selectLabel"
              options={clusterOptions.map((option) => ({
                value: option.value,
                selectLabel: option.label,
                disabled: option.status !== 'running',
                label: (
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                    }}
                  >
                    <span>{option.label}</span>
                    <span>{option.status}</span>
                  </div>
                ),
              }))}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={['name']}
            validateFirst
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.EnterAResourceName',
                  defaultMessage: '请输入资源名',
                }),
              },
              {
                pattern: /\D/,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.ResourceNamesCannotUsePure',
                  defaultMessage: '资源名不能使用纯数字',
                }),
              },
              resourceNameRule,
            ]}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.BasicInfo.ResourceName',
              defaultMessage: '资源名',
            })}
          >
            <Input />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={['tenantName']}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.BasicInfo.TenantName',
              defaultMessage: '租户名',
            })}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.EnterATenantName',
                  defaultMessage: '请输入租户名',
                }),
              },
              {
                pattern: TZ_NAME_REG,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.TheFirstCharacterMustBe',
                  defaultMessage: '首字符必须是字母或者下划线，不能包含 -',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.PleaseEnter',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <PasswordInput
            value={passwordVal}
            onChange={setPasswordVal}
            form={form}
            name="rootPassword"
          />
        </Col>
        <Col span={8}>
          <Form.Item
            name={['unitNum']}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.PleaseEnterTheNumberOf',
                  defaultMessage: '请输入Unit 数量',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.BasicInfo.NumberOfUnits',
              defaultMessage: 'Unit 数量',
            })}
          >
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={['connectWhiteList']}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.BasicInfo.ConnectionWhitelist',
              defaultMessage: '连接白名单',
            })}
          >
            <Select mode="tags" />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={'优化场景'}
            name={'loadType'}
            initialValue="HTAP"
            rules={[
              {
                required: true,
                message: '请选择优化场景',
              },
            ]}
          >
            <Select>
              {LOADTYPE_LIST?.map((item) => (
                <Option key={item.value} value={item.value}>
                  {item.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
        <Space>
          删除保护
          <Checkbox
            defaultChecked={deleteVal}
            onChange={(e) => {
              setDeleteVal(e.target.checked);
            }}
          />
        </Space>
        {/* <Col span={8}>
              <Form.Item name={["charset"]} label="字符集">
                <Select />
              </Form.Item>
             </Col> */}
      </Row>
    </Card>
  );
}
