import PasswordInput from '@/components/PasswordInput';
import { Card, Col, Form, Input, InputNumber, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';

interface BasicInfoProps {
  form: FormInstance<any>;
  passwordVal: string;
  clusterList: API.SimpleClusterList;
  setSelectClusterId: React.Dispatch<
    React.SetStateAction<
      | number
      | undefined
    >
  >;
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
    value:cluster.clusterId,
    label: cluster.name,
  }));
  const selectClusterChange = (id:number) => {
    setSelectClusterId(id)
  };
  return (
    <Card title="基本信息">
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <Form.Item
            name={["obcluster"]}
            rules={[
              {
                required: true,
                message: '请输入OB集群',
              },
            ]}
            label="OB集群"
          >
            <Select
              onChange={(value) => selectClusterChange(value)}
              options={clusterOptions}
            />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={["name"]}
            rules={[
              {
                required: true,
                message: '请输入资源名',
              },
            ]}
            label="资源名"
          >
            <Input />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={["tenantName"]}
            label="租户名"
            rules={[
              {
                required: true,
                message: '请输入租户名',
              },
            ]}
          >
            <Input placeholder="请输入" />
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
            name={["unitNum"]}
            rules={[
              {
                required: true,
                message: '请输入Unit 数量',
              },
            ]}
            label="Unit 数量"
          >
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item name={["connectWhiteList"]} label="连接白名单">
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
