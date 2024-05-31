import { alert } from '@/api';
import type { RuleRule } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import InputLabel from '@/components/InputLabel';
import { LEVER_OPTIONS_ALARM, SERVERITY_MAP } from '@/constants';
import { QuestionCircleOutlined } from '@ant-design/icons';
import type { DrawerProps } from 'antd';
import {
  Col,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
  Tag,
  message,
} from 'antd';
import { useEffect } from 'react';

type AlertRuleDrawerProps = {
  ruleName?: string;
  submitCallback?: () => void;
} & DrawerProps;
const { TextArea } = Input;
export default function RuleDrawerForm({
  ruleName,
  submitCallback,
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
    instanceType: 'obcluster',
  };
  const submit = (values: RuleRule) => {
    alert.createOrUpdateRule(values).then(({ successful }) => {
      if (successful) {
        message.success('操作成功！');
        submitCallback && submitCallback();
      }
    });
  };

  useEffect(() => {
    if (ruleName) {
      alert.getRule(ruleName).then(({ data, successful }) => {
        if (successful) {
          form.setFieldsValue({ ...data });
        }
      });
    }
  }, [ruleName]);

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
        onFinish={submit}
        validateTrigger="onBlur"
        form={form}
      >
        <Row gutter={[24, 24]}>
          <Col span={24}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请选择',
                },
              ]}
              name={'instanceType'}
              label="屏蔽对象类型"
            >
              <Radio.Group>
                <Radio value="obcluster"> 集群 </Radio>
                <Radio value="obtenant"> 租户 </Radio>
                <Radio value="observer"> OBServer </Radio>
              </Radio.Group>
            </Form.Item>
          </Col>

          <Col span={16}>
            <Form.Item
              name={'name'}
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
                {
                  validator: async (_, value) => {
                    const res = await alert.listRules();
                    if (res.successful) {
                      for (const rule of res.data) {
                        if (rule.name === value) {
                          return Promise.reject(
                            new Error('告警规则已存在，请重新输入'),
                          );
                        }
                      }
                    }
                    return Promise.resolve();
                  },
                },
              ]}
              label="告警规则名"
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
              name={'serverity'}
              label="告警级别"
            >
              <Select
                options={LEVER_OPTIONS_ALARM?.map((item) => ({
                  value: item.value,
                  label: (
                    <Tag color={SERVERITY_MAP[item?.value]?.color}>
                      {item.label}
                    </Tag>
                  ),
                }))}
                placeholder="请选择"
              />
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
              name={'duration'}
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
