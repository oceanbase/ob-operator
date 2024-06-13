import { alert } from '@/api';
import type { CommonKVPair, RuleRule } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import InputLabelComp from '@/components/InputLabelComp';
import InputTimeComp from '@/components/InputTimeComp';
import { LEVER_OPTIONS_ALARM, SEVERITY_MAP } from '@/constants';
import { intl } from '@/utils/intl';
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
        message.success(
          intl.formatMessage({
            id: 'src.pages.Alert.Rules.5D79276F',
            defaultMessage: '操作成功！',
          }),
        );
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
      title={intl.formatMessage({
        id: 'src.pages.Alert.Rules.72206E63',
        defaultMessage: '告警规则配置',
      })}
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
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.7F6B182B',
                    defaultMessage: '请选择',
                  }),
                },
              ]}
              name={'instanceType'}
              label={intl.formatMessage({
                id: 'src.pages.Alert.Rules.6B2322AF',
                defaultMessage: '对象类型',
              })}
            >
              <Radio.Group>
                <Radio value="obcluster">
                  {intl.formatMessage({
                    id: 'src.pages.Alert.Rules.60487F0F',
                    defaultMessage: '集群',
                  })}
                </Radio>
                <Radio value="obtenant">
                  {intl.formatMessage({
                    id: 'src.pages.Alert.Rules.C7EBBB92',
                    defaultMessage: '租户',
                  })}
                </Radio>
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
                        message: intl.formatMessage({
                          id: 'src.pages.Alert.Rules.B7B764AE',
                          defaultMessage: '请输入',
                        }),
                      },
                    ]
                  : [
                      {
                        required: true,
                        message: intl.formatMessage({
                          id: 'src.pages.Alert.Rules.50003344',
                          defaultMessage: '请输入',
                        }),
                      },
                      {
                        validator: async (_, value) => {
                          if (rules) {
                            for (const rule of rules) {
                              if (rule.name === value) {
                                return Promise.reject(
                                  new Error(
                                    intl.formatMessage({
                                      id: 'src.pages.Alert.Rules.B46056EE',
                                      defaultMessage:
                                        '告警规则已存在，请重新输入',
                                    }),
                                  ),
                                );
                              }
                            }
                          }
                          return Promise.resolve();
                        },
                      },
                    ]
              }
              label={intl.formatMessage({
                id: 'src.pages.Alert.Rules.14235DA8',
                defaultMessage: '告警规则名',
              })}
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Rules.63C9E8E6',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.9CB4B6C4',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              name={'severity'}
              label={intl.formatMessage({
                id: 'src.pages.Alert.Rules.821C3FA7',
                defaultMessage: '告警级别',
              })}
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
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Rules.528ED58D',
                  defaultMessage: '请选择',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={16}>
            <Form.Item
              name={'query'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.412594B9',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.C01B1EFD',
                    defaultMessage: '判定告警的 PromQL 表达式',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.9A2D5103',
                    defaultMessage: '指标计算表达式',
                  })}
                />
              }
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Rules.179A6ACF',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.0FF3431E',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={intl.formatMessage({
                id: 'src.pages.Alert.Rules.D7E8AEDB',
                defaultMessage: '持续时间',
              })}
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
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.A2A41881',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.8DE8BA49',
                    defaultMessage:
                      '告警事件的摘要信息模版，可以使用 {{ }} 来标记需要替换的值',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.363B4BDE',
                    defaultMessage: 'summary 信息',
                  })}
                />
              }
            >
              <TextArea
                rows={4}
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Rules.9EDCE4CA',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              name={'description'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Rules.E04B7BC2',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.5FB853B5',
                    defaultMessage:
                      '告警事件的详情信息模版，可以使用 {{ }} 来标记需要替换的值',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.B11DCDB0',
                    defaultMessage: '告警详情信息',
                  })}
                />
              }
            >
              <TextArea
                rows={4}
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Rules.E336DF4E',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.1E26B90F',
                    defaultMessage: '添加到告警事件的标签',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Rules.66144CF9',
                    defaultMessage: '标签',
                  })}
                />
              }
              validateDebounce={1500}
              rules={[
                {
                  validator: (_, value: CommonKVPair[]) => {
                    if (!validateLabelValues(value)) {
                      return Promise.reject(
                        intl.formatMessage({
                          id: 'src.pages.Alert.Rules.0EAD0426',
                          defaultMessage: '请检查标签是否完整输入',
                        }),
                      );
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
