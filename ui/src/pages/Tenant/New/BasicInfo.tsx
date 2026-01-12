import PasswordInput from '@/components/PasswordInput';
import { LOADTYPE_LIST, TZ_NAME_REG } from '@/constants';
import { resourceNameRule } from '@/constants/rules';
import { intl } from '@/utils/intl';
import { Card, Checkbox, Col, Form, Input, Row, Select, Space } from 'antd';
import type { FormInstance } from 'antd/lib/form';
import { useEffect } from 'react';

const { Option } = Select;

interface BasicInfoProps {
  form: FormInstance<API.NewTenantForm>;
  passwordVal: string;
  clusterList: API.SimpleClusterList;
  setSelectClusterId: React.Dispatch<React.SetStateAction<number | undefined>>;
  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
  deleteValue: boolean;
  setDeleteValue: (e: boolean) => void;
}
export default function BasicInfo({
  form,
  passwordVal,
  clusterList,
  setPasswordVal,
  setSelectClusterId,
  deleteValue,
  setDeleteValue,
  sqlDiagnoseValue,
  setSqlDiagnoseValue,
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
  const path = window.location.hash.split('=')[1];

  useEffect(() => {
    if (path) {
      const cluster = clusterOptions?.find((item) => item.label === path);
      setSelectClusterId(cluster?.value || undefined);
      form.setFieldsValue({
        obcluster: cluster?.value,
      });
    }
  }, [clusterList, path]);
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
            label={intl.formatMessage({
              id: 'src.pages.Tenant.New.6071B46A',
              defaultMessage: '优化场景',
            })}
            name={['scenario']}
            initialValue="HTAP"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.Tenant.New.7D0448C6',
                  defaultMessage: '请选择优化场景',
                }),
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
        <Col span={2}>
          <Space>
            {intl.formatMessage({
              id: 'src.pages.Tenant.New.3979BAB6',
              defaultMessage: '删除保护',
            })}
            <Checkbox
              defaultChecked={deleteValue}
              onChange={(e) => {
                setDeleteValue(e.target.checked);
              }}
            />
          </Space>
        </Col>
        <Col span={2}>
          <Space>
            SQL 诊断
            <Checkbox
              defaultChecked={sqlDiagnoseValue}
              onChange={(e) => {
                setSqlDiagnoseValue(e.target.checked);
              }}
            />
          </Space>
        </Col>
      </Row>
    </Card>
  );
}
