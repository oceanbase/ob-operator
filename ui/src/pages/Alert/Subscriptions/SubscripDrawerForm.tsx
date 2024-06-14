import { alert } from '@/api';
import type { RouteRoute } from '@/api/generated';
import { AlarmMatcher } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import InputLabelComp from '@/components/InputLabelComp';
import InputTimeComp from '@/components/InputTimeComp';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Col, Form, Row, Select, message } from 'antd';
import { useEffect } from 'react';
import { validateLabelValues } from '../helper';

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
  const [form] = Form.useForm<RouteRoute>();
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
  const submit = async (values: RouteRoute) => {
    if (isEdit) values.id = id;
    if (!values.matchers) values.matchers = [];
    values.matchers = values.matchers.filter(
      (matcher) => matcher.name && matcher.value,
    );
    const { successful } = await alert.createOrUpdateRoute(values);
    if (successful) {
      message.success(`${isEdit ? '修改' : '创建'}成功!`);
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
            <p>通道配置</p>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请选择',
                },
              ]}
              label="告警通道"
              name={'receiver'}
            >
              <Select
                placeholder="请选择"
                options={listReceivers?.map((receiver) => ({
                  label: receiver.name,
                  value: receiver.name,
                }))}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <p>匹配配置</p>
            <Form.Item
              name={'matchers'}
              validateDebounce={1500}
              rules={[
                {
                  validator: (_, value: AlarmMatcher[]) => {
                    if (!validateLabelValues(value)) {
                      return Promise.reject('请检查标签是否完整输入');
                    }
                    return Promise.resolve();
                  },
                },
              ]}
              label={
                <IconTip
                  icon={
                    <span style={{ color: 'rgba(0, 0, 0, 0.45)' }}>
                      <QuestionCircleOutlined />
                      （可选）
                    </span>
                  }
                  tip="按照标签匹配条件推送告警，支持值匹配或者正则表达式，当所有条件都满足时告警才会被推送"
                  content="标签"
                />
              }
            >
              <InputLabelComp regex={true} defaulLabelName="name" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请输入',
                },
              ]}
              name={'aggregateLabels'}
              label={<IconTip tip="告警聚合使用的标签" content="聚合标签" />}
            >
              <Select
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
                  message: '请输入',
                },
              ]}
              label={
                <IconTip tip="告警消息推送的重复周期" content="推送周期" />
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
                  message: '请输入',
                },
              ]}
              name={'groupWait'}
              label="聚合等待时间"
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
                  message: '请输入',
                },
              ]}
              label={
                <IconTip tip="告警消息聚合的时间区间" content="聚合区间" />
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
