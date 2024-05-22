import AlertDrawer from '@/components/AlertDrawer';
import InputLabel from '@/components/InputLabel';
import { QuestionCircleOutlined } from '@ant-design/icons';
import type { DrawerProps } from 'antd';
import { Col, Form, Input, InputNumber, Row, Select } from 'antd';
import { useEffect } from 'react';

type AlertRuleDrawerProps = {
  ruleName?: string;
} & DrawerProps;
const { TextArea } = Input;
export default function RuleDrawerForm({
  ruleName,
  ...props
}: AlertRuleDrawerProps) {
  const [form] = Form.useForm();
  const initialValues = {
    labels: [
      {
        key: '',
        value: '',
      },
    ],
  };

  useEffect(() => {
    if (ruleName) {
      // Something to do
    }
  }, []);

  return (
    <AlertDrawer
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      {...props}
    >
      <Form
        initialValues={initialValues}
        preserve={false}
        style={{ marginBottom: 64 }}
        layout="vertical"
        form={form}
      >
        <Row gutter={[24, 24]}>
          <Col span={16}>
            <Form.Item
              name={'name'}
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              label="告警规则名"
            >
              <Input placeholder="请输入" />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item name={'serverity'} label="告警级别">
              <Select placeholder="请选择" />
            </Form.Item>
          </Col>
          <Col span={16}>
            <Form.Item
              name={'query'}
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              label={
                <div>
                  <span>指标计算表达式</span>
                  <QuestionCircleOutlined />
                </div>
              }
            >
              <Input placeholder="请输入" />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              label="持续时间"
              name={'description'}
            >
              <InputNumber placeholder="请输入" addonAfter="分钟" min={1} />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              name={'summary'}
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              label={
                <div>
                  <span>summary 信息</span>
                  <QuestionCircleOutlined />
                </div>
              }
            >
              <TextArea rows={4} placeholder="请输入" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              name={'description'}
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              label={
                <div>
                  <span>告警详情信息</span>
                  <QuestionCircleOutlined />
                </div>
              }
            >
              <TextArea rows={4} placeholder="请输入" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              label={
                <div>
                  <span>标签</span>
                  <QuestionCircleOutlined />
                  <span style={{ color: 'rgba(0,0,0,0.45)' }}>(可选)</span>
                </div>
              }
            >
              <InputLabel
                wrapFormName="labels"
                labelFormName="key"
                valueFormName="value"
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </AlertDrawer>
  );
}
