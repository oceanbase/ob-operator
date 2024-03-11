import { intl } from '@/utils/intl';
import { Form, InputNumber, message } from 'antd';
import CustomModal from '.';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { scaleObserver } from '@/services';
import { useEffect } from 'react';
import type { CommonModalType } from '.';

interface ScaleModalProps {
  zoneName?: string;
  defaultValue?: number;
}

export default function ScaleModal({
  visible,
  setVisible,
  successCallback,
  zoneName,
  defaultValue = 1,
}: ScaleModalProps & CommonModalType) {
  const [form] = Form.useForm();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const onFinish = async (val: any) => {
    if (!zoneName) throw new Error('zoneName is not defined');
    const [namespace, name] = getNSName();
    const res = await scaleObserver({
      namespace,
      name,
      zoneName,
      replicas: val.replicas,
    });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  useEffect(() => {
    form.setFieldValue('replicas', defaultValue);
  }, [defaultValue]);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.ScaleModal.ExpandZone',
        defaultMessage: '扩容Zone',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={() => {
        setVisible(false);
        form.resetFields();
      }}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item
          label="servers"
          name="replicas"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.ScaleModal.PleaseEnterTheNumberOf',
                defaultMessage: '请输入扩缩容数目!',
              }),
            },
          ]}
        >
          <InputNumber
            defaultValue={defaultValue}
            min={1}
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.ScaleModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
