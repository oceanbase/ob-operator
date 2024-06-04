import { alert } from '@/api';
import type { RouteRoute } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import InputLabel from '@/components/InputLabel';
import { QuestionCircleOutlined } from '@ant-design/icons';
import type { DrawerProps } from 'antd';
import { Col, Form, InputNumber, Row, Select, message } from 'antd';
import { filterLabel } from '../helper';
import { useEffect } from 'react';
import styles from './index.less';

interface ShieldDrawerFormProps extends DrawerProps {
  id?: string;
  recevierNames: string[];
  onClose: () => void;
  submitCallback?: () => void;
}

export default function SubscripDrawerForm({
  id,
  recevierNames,
  submitCallback,
  onClose,
  ...props
}: ShieldDrawerFormProps) {
  const isEdit = !!id;
  const recevierOptions = recevierNames.map((name) => ({
    label: name,
    value: name,
  }));
  const [form] = Form.useForm<RouteRoute>();
  const initialValues = {
    matchers: [
      {
        isRegex: false,
        name: '',
        value: '',
      },
    ],
  };
  const submit = async (values: RouteRoute) => {
    values.matchers = filterLabel(values.matchers);
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
              <Select placeholder="请选择" options={recevierOptions} />
            </Form.Item>
          </Col>
          <Col span={24}>
            <p>匹配配置</p>
            <Form.Item
              label={
                <div>
                  <span>标签</span>
                  <QuestionCircleOutlined className={styles.questionIcon} />
                  <span style={{ color: 'rgba(0,0,0,0.45)' }}>(可选)</span>
                </div>
              }
            >
              <InputLabel
                wrapFormName="matchers"
                labelFormName="name"
                valueFormName="value"
                regBoxFormName="isRegex"
                form={form}
              />
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
              label={
                <div>
                  聚合标签{' '}
                  <QuestionCircleOutlined className={styles.questionIcon} />
                </div>
              }
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
                <div>
                  推送周期{' '}
                  <QuestionCircleOutlined className={styles.questionIcon} />
                </div>
              }
            >
              <InputNumber min={1} addonAfter="分钟" />
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
              <InputNumber min={1} addonAfter="分钟" />
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
                <div>
                  聚合区间{' '}
                  <QuestionCircleOutlined className={styles.questionIcon} />
                </div>
              }
            >
              <InputNumber min={1} addonAfter="分钟" />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </AlertDrawer>
  );
}
