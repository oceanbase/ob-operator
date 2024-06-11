import { alert } from '@/api';
import type { CommonKVPair, RuleRule } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import InputLabelComp from '@/components/InputLabelComp';
import InputTimeComp from '@/components/InputTimeComp';
import { LEVER_OPTIONS_ALARM, SEVERITY_MAP } from '@/constants';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Col, Form, Input, Radio, Row, Select, Tag, message } from 'antd';
import { useEffect } from 'react';
import { validateLabelValues } from '../helper';

type AlertRuleDrawerProps = {
  ruleName?: string;
  onClose: () => void;
  submitCallback?: () => void;
} & DrawerProps;
const { TextArea } = Input;
export default function RuleDrawerForm({
  ruleName,
  submitCallback,
  onClose,
  ...props
}: AlertRuleDrawerProps) {
  const [form] = Form.useForm();
  const { data: rulesRes } = useRequest(alert.listRules);
  const rules = rulesRes?.data;
  const isEdit = !!ruleName;
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
    if (!values.labels) values.labels = [];
    values.labels = values.labels.filter((label) => label.key && label.value);
    alert.createOrUpdateRule(values).then(({ successful }) => {
      if (successful) {
        message.success('操作成功！');
        onClose();
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
      title="告警规则配置"
      onClose={onClose}
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
        <Row gutter={[24, 0]}>
          <Col span={24}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请选择',
                },
              ]}
              name={'instanceType'}
              label="对象类型"
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
              rules={
                isEdit
                  ? [
                      {
                        required: true,
                        message: '请输入',
                      },
                    ]
                  : [
                      {
                        required: true,
                        message: '请输入',
                      },
                      {
                        validator: async (_, value) => {
                          if (rules) {
                            for (const rule of rules) {
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
                    ]
              }
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
              name={'severity'}
              label="告警级别"
            >
              <Select
                options={LEVER_OPTIONS_ALARM?.map((item) => ({
                  value: item.value,
                  label: (
                    <Tag color={SEVERITY_MAP[item?.value]?.color}>
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
              <InputTimeComp />
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
              validateDebounce={1500}
              rules={[
                {
                  validator: (_, value: CommonKVPair[]) => {
                    if (!validateLabelValues(value)) {
                      return Promise.reject('请检查标签是否完整输入');
                    }
                    return Promise.resolve();
                  },
                },
              ]}
              name="labels"
            >
              <InputLabelComp />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </AlertDrawer>
  );
}
