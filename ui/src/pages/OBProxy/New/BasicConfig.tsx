import IconTip from '@/components/IconTip';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import { getSimpleClusterList } from '@/services';
import { useRequest } from 'ahooks';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';

interface BasicConfigProps {
  form: FormInstance<any>;
}

export default function BasicConfig({ form }: BasicConfigProps) {
  const { data: clusterListRes } = useRequest(getSimpleClusterList);
  const clisterList = clusterListRes?.data.map((cluster) => ({
    label: cluster.name,
    value: `${cluster.name}+${cluster.namespace}`,
  }));
  return (
    <Card title="基本设置">
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <Form.Item
            label="资源名称"
            rules={[
              {
                required: true,
                message: '请输入',
              },
            ]}
          >
            <Input placeholder="请输入" />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label="OBProxy 集群名"
            rules={[
              {
                required: true,
                message: '请输入',
              },
            ]}
          >
            <Input placeholder="请输入" />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label="连接 OB 集群"
            name="obCluster"
            rules={[
              {
                required: true,
                message: '请选择',
              },
            ]}
          >
            <Select options={clisterList} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label={
              <IconTip content="OBProxy root 密码" tip="root@proxysys 密码" />
            }
            rules={[
              {
                required: true,
                message: '请输入',
              },
            ]}
          >
            <Input.Password placeholder="请输入" />
          </Form.Item>
        </Col>
        <Col span={8}>
          <SelectNSFromItem form={form} />
        </Col>
      </Row>
    </Card>
  );
}
