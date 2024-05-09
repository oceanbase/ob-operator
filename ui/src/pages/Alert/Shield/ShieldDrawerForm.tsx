import { alert } from '@/api';
import AlertDrawer from '@/components/AlertDrawer';
import InputLabel from '@/components/InputLabel';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useUpdateEffect } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Button, Col, DatePicker, Form, Input, Radio, Row } from 'antd';
import dayjs from 'dayjs';
import { useEffect } from 'react';
import styles from './index.less';

type ShieldDrawerProps = {
  id?: string;
} & DrawerProps;

const { TextArea } = Input;

export default function ShieldDrawerForm({ id, ...props }: ShieldDrawerProps) {
  const [form] = Form.useForm();
  const instanceType = Form.useWatch(['instance', 'type'], form);
  const initialValues = {
    matchers: [
      {
        name: '',
        value: '',
        isRegex: false,
      },
    ],
  };

  const fieldEndTimeChange = (time: number | Date) => {
    if (typeof time === 'number') {
      form.setFieldValue('endsAt', dayjs(new Date().valueOf() + time));
    } else {
      form.setFieldValue('endsAt', dayjs(time));
    }
  };

  useEffect(() => {
    if (id) {
      alert.getSilencer(id).then(() => {
        // Something to do
      });
    }
  }, []);

  useUpdateEffect(() => {
    form.setFieldValue(['instance', instanceType], 'aa');
  }, [instanceType]);
  return (
    <AlertDrawer onSubmit={() => form.submit()} {...props}>
      <Form form={form} layout="vertical" initialValues={initialValues}>
        <Form.Item name={['instance', 'type']} label="屏蔽对象类型">
          <Radio.Group>
            <Radio value="obcluster"> 集群 </Radio>
            <Radio value="obtenant"> 租户 </Radio>
            <Radio value="observer"> OBServer </Radio>
          </Radio.Group>
        </Form.Item>
        {/* <Form.Item label='屏蔽告警规则'>
            
        </Form.Item> */}
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
            showRegBox={true}
          />
        </Form.Item>
        <Row style={{ alignItems: 'center' }}>
          <Col>
            <Form.Item name="endsAt" label="屏蔽结束时间">
              <DatePicker showTime format="YYYY-MM-DD HH:mm:ss" />
            </Form.Item>
          </Col>
          <Col>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(6 * 3600 * 1000)}
            >
              6小时
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(12 * 3600 * 1000)}
            >
              12小时
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(24 * 3600 * 1000)}
            >
              1天
            </Button>
            <Button
              onClick={() =>
                fieldEndTimeChange(new Date('2099-12-31 23:59:59'))
              }
              type="link"
            >
              永久
            </Button>
          </Col>
        </Row>
        <Form.Item name={'comment'} label="备注信息">
          <TextArea rows={4} placeholder="请输入" />
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
