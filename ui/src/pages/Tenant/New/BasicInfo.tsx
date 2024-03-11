import PasswordInput from '@/components/PasswordInput';
import { intl } from '@/utils/intl';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import InputNumber from '@/components/InputNumber';
import type { FormInstance } from 'antd/lib/form';

interface BasicInfoProps {
  form: FormInstance<any>;
  passwordVal: string;
  clusterList: API.SimpleClusterList;
  setSelectClusterId: React.Dispatch<React.SetStateAction<number | undefined>>;

  setPasswordVal: React.Dispatch<React.SetStateAction<string>>;
}
export default function BasicInfo({
  form,
  passwordVal,
  clusterList,
  setPasswordVal,
  setSelectClusterId,
}: BasicInfoProps) {
  const clusterOptions = clusterList.map((cluster) => ({
    value: cluster.clusterId,
    label: cluster.name,
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
              onChange={(value) => selectClusterChange(value)}
              options={clusterOptions}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={['name']}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'Dashboard.Tenant.New.BasicInfo.EnterAResourceName',
                  defaultMessage: '请输入资源名',
                }),
              },
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
            <InputNumber style={{ width: '100%' }} />
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
        {/* <Col span={8}>
           <Form.Item name={["charset"]} label="字符集">
             <Select />
           </Form.Item>
          </Col> */}
      </Row>
    </Card>
  );
}
