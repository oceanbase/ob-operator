import { alert } from '@/api';
import type { RouteRouteParam } from '@/api/generated';
import { AlarmMatcher } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import InputLabelComp from '@/components/InputLabelComp';
import InputTimeComp from '@/components/InputTimeComp';
import { LABELNAME_REG, VALIDATE_DEBOUNCE } from '@/constants';
import { LABEL_NAME_RULE } from '@/constants/rules';
import { intl } from '@/utils/intl';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Col, Form, Row, Select, message } from 'antd';
import { useEffect } from 'react';
import { validateLabelValues } from '../helper';
import styles from './index.less';

interface ShieldDrawerFormProps extends DrawerProps {
  id?: string;
  onClose: () => void;
  submitCallback?: () => void;
}

export default function SubscripDrawerForm({
  id,
  submitCallback,
  onClose,
  ...props
}: ShieldDrawerFormProps) {
  const isEdit = !!id;
  const [form] = Form.useForm<RouteRouteParam>();
  const initialValues = {
    matchers: [
      {
        name: '',
        value: '',
        isRegex: false,
      },
    ],
  };
  const { data: listReceiversRes } = useRequest(alert.listReceivers);
  const listReceivers = listReceiversRes?.data;
  const submit = async (values: RouteRouteParam) => {
    if (isEdit) values.id = id;
    if (!values.matchers) values.matchers = [];
    values.matchers = values.matchers.filter(
      (matcher) => matcher.name && matcher.value,
    );
    const { successful } = await alert.createOrUpdateRoute(values);
    if (successful) {
      message.success(
        intl.formatMessage(
          {
            id: 'src.pages.Alert.Subscriptions.BA84E413',
            defaultMessage: "{ConditionalExpression0}成功!",
          },
          {
            ConditionalExpression0: isEdit
              ? intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.8F1CA27C',
                  defaultMessage: '修改',
                })
              : intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.B4AB7DEF',
                  defaultMessage: '创建',
                }),
          },
        ),
      );
      submitCallback && submitCallback();
      onClose();
    }
  };
  useEffect(() => {
    if (isEdit) {
      alert.getRoute(id).then(({ successful, data }) => {
        if (successful) {
          form.setFieldsValue(data);
        }
      });
    }
  }, [id]);
  return (
    <AlertDrawer
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      onClose={() => {
        onClose();
      }}
      {...props}
    >
      <Form
        form={form}
        preserve={false}
        onFinish={submit}
        initialValues={initialValues}
        layout="vertical"
      >
        <Row>
          <Col span={12}>
            <p>
              {intl.formatMessage({
                id: 'src.pages.Alert.Subscriptions.F5FC734D',
                defaultMessage: '通道配置',
              })}
            </p>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.9FAEA249',
                    defaultMessage: '请选择',
                  }),
                },
              ]}
              label={intl.formatMessage({
                id: 'src.pages.Alert.Subscriptions.898A358F',
                defaultMessage: '告警通道',
              })}
              name={'receiver'}
            >
              <Select
                placeholder={intl.formatMessage({
                  id: 'src.pages.Alert.Subscriptions.7C1E1B2A',
                  defaultMessage: '请选择',
                })}
                options={listReceivers?.map((receiver) => ({
                  label: receiver.name,
                  value: receiver.name,
                }))}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <p>
              {intl.formatMessage({
                id: 'src.pages.Alert.Subscriptions.A797D56B',
                defaultMessage: '匹配配置',
              })}
            </p>
            <Form.Item
              name={'matchers'}
              validateDebounce={VALIDATE_DEBOUNCE}
              validateFirst
              rules={[
                {
                  validator: (_, value: AlarmMatcher[]) => {
                    if (!validateLabelValues(value)) {
                      return Promise.reject(
                        intl.formatMessage({
                          id: 'src.pages.Alert.Subscriptions.971D4038',
                          defaultMessage: '请检查标签是否完整输入',
                        }),
                      );
                    }
                    return Promise.resolve();
                  },
                },
                LABEL_NAME_RULE,
              ]}
              label={
                <IconTip
                  icon={
                    <span style={{ color: 'rgba(0, 0, 0, 0.45)' }}>
                      <QuestionCircleOutlined />
                      {intl.formatMessage({
                        id: 'src.pages.Alert.Subscriptions.7237B7FE',
                        defaultMessage: '（可选）',
                      })}
                    </span>
                  }
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.2071AD79',
                    defaultMessage:
                      '按照标签匹配条件推送告警，支持值匹配或者正则表达式，当所有条件都满足时告警才会被推送',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.DD88FE8C',
                    defaultMessage: '标签',
                  })}
                />
              }
            >
              <InputLabelComp regex={true} defaulLabelName="name" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              validateFirst
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.22115027',
                    defaultMessage: '请输入',
                  }),
                },
                {
                  validator: (_, value) => {
                    if (
                      value.some((item: string) => !LABELNAME_REG.test(item))
                    ) {
                      return Promise.reject(
                        intl.formatMessage({
                          id: 'src.pages.Alert.Subscriptions.FEDAA081',
                          defaultMessage:
                            '标签名需满足以字母或下划线开头，包含字母，数字，下划线',
                        }),
                      );
                    }
                    return Promise.resolve();
                  },
                },
              ]}
              name={'aggregateLabels'}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.E6C81EFB',
                    defaultMessage: '告警聚合使用的标签',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.A650A8D0',
                    defaultMessage: '聚合标签',
                  })}
                />
              }
            >
              <Select
                className={styles.aggregateLabels}
                mode="tags"
                tokenSeparators={[',']}
                dropdownStyle={{ display: 'none' }}
                style={{ width: '100%' }}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name={'repeatInterval'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.F8964AD2',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.7CCE86F7',
                    defaultMessage: '告警消息推送的重复周期',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.15BFC65C',
                    defaultMessage: '推送周期',
                  })}
                />
              }
            >
              <InputTimeComp />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.A4615AC5',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              name={'groupWait'}
              label={intl.formatMessage({
                id: 'src.pages.Alert.Subscriptions.E6F8DAA7',
                defaultMessage: '聚合等待时间',
              })}
            >
              <InputTimeComp />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name={'groupInterval'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.DEAF40DE',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.9CB5A5C8',
                    defaultMessage: '告警消息聚合的时间区间',
                  })}
                  content={intl.formatMessage({
                    id: 'src.pages.Alert.Subscriptions.CBA8D190',
                    defaultMessage: '聚合区间',
                  })}
                />
              }
            >
              <InputTimeComp />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </AlertDrawer>
  );
}
