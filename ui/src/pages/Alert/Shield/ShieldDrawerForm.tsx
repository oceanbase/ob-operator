import { alert } from '@/api';
import type { SilenceSilencerParam } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import InputLabel from '@/components/InputLabel';
import { Alert } from '@/type/alert';
import { QuestionCircleOutlined } from '@ant-design/icons';
import type { DrawerProps } from 'antd';
import {
  Button,
  Col,
  DatePicker,
  Form,
  Input,
  Radio,
  Row,
  message,
} from 'antd';
import dayjs from 'dayjs';
import { useEffect } from 'react';
import ShieldObjInput from './ShieldObjInput';
import styles from './index.less';

interface ShieldDrawerProps extends DrawerProps {
  id?: string;
  initialValues?: Alert.ShieldDrawerInitialValues;
  onClose: () => void;
}

const { TextArea } = Input;

export default function ShieldDrawerForm({
  id,
  onClose,
  initialValues,
  ...props
}: ShieldDrawerProps) {
  const [form] = Form.useForm<SilenceSilencerParam>();
  const shieldObjType = Form.useWatch(['instance', 'type'], form);
  const isEdit = !!id;
  const newInitialValues = {
    matchers: [
      {
        name: '',
        value: '',
        isRegex: false,
      },
    ],
    instance:{
      type:'obcluster'
    },
    ...initialValues,
  };

  const fieldEndTimeChange = (time: number | Date) => {
    if (typeof time === 'number') {
      form.setFieldValue('endsAt', dayjs(new Date().valueOf() + time));
    } else {
      form.setFieldValue('endsAt', dayjs(time));
    }
  };

  const submit = (values: SilenceSilencerParam) => {
    alert.createOrUpdateSilencer(values).then(({ successful }) => {
      if (successful) {
        message.success('操作成功!');
        onClose();
      }
    });
  };

  useEffect(() => {
    if (isEdit) {
      alert.getSilencer(id).then(({ successful, data }) => {
        if (successful) {
          form.setFieldsValue(data);
        }
      });
    }
  }, [id]);

  return (
    <AlertDrawer
      onClose={() => {
        onClose();
      }}
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      title="屏蔽条件"
      {...props}
    >
      <Form
        form={form}
        preserve={false}
        onFinish={submit}
        layout="vertical"
        initialValues={newInitialValues}
      >
        <Form.Item name={['instance', 'type']} label="屏蔽对象类型">
          <Radio.Group>
            <Radio value="obcluster"> 集群 </Radio>
            <Radio value="obtenant"> 租户 </Radio>
            <Radio value="observer"> OBServer </Radio>
          </Radio.Group>
        </Form.Item>
        <Form.Item label='屏蔽对象'>
          <ShieldObjInput shieldObjType={shieldObjType} form={form} />
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
            regBoxFormName="isRegex"
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
