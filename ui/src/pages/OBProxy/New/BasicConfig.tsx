import IconTip from '@/components/IconTip';
import SelectNSFromItem from '@/components/SelectNSFromItem';
import TooltipPretty from '@/components/TooltipPretty';
import { resourceNameRule } from '@/constants/rules';
import { getSimpleClusterList } from '@/services';
import { useRequest } from 'ahooks';
import { Card, Col, Form, Input, Row, Select } from 'antd';
import type { FormInstance } from 'antd/lib/form';
import { useEffect } from 'react';

interface BasicConfigProps {
  form: FormInstance<any>;
}

export default function BasicConfig({ form }: BasicConfigProps) {
  const { data: clusterListRes } = useRequest(getSimpleClusterList);
  const selectCluster = Form.useWatch('obCluster');
  const clisterList = clusterListRes?.data.map((cluster) => ({
    label: cluster.name,
    value: JSON.stringify({ name: cluster.name, namespace: cluster.namespace }),
  }));

  useEffect(() => {
    if (selectCluster && !form.getFieldValue('namespace')) {
      form.setFieldValue('namespace', selectCluster.split('+')?.[1]);
    }
  }, [selectCluster]);
  return (
    <Card title="基本设置">
      <Row gutter={[16, 32]}>
        <Col span={8}>
          <TooltipPretty title={'k8s中资源的名称'}>
            <Form.Item
              label="资源名称"
              name={'name'}
              rules={[
                {
                  required: true,
                  message: '请输入k8s资源名称',
                },
                {
                  pattern: /\D/,
                  message: '资源名不能使用纯数字',
                },
                resourceNameRule,
              ]}
            >
              <Input placeholder="请输入" />
            </Form.Item>
          </TooltipPretty>
        </Col>
        <Col span={8}>
          <Form.Item
            label="OBProxy 集群名"
            name={'proxyClusterName'}
            rules={[
              {
                required: true,
                message: '请输入集群名',
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
                message: '请选择 OB 集群',
              },
            ]}
          >
            <Select options={clisterList} />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            name={'proxySysPassword'}
            label={
              <IconTip content="OBProxy root 密码" tip="root@proxysys 密码" />
            }
            rules={[
              {
                required: true,
                message: '请输入 OBProxy root 密码',
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
