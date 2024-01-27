import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { ReplayLogOfTenant } from '@/services/tenant';
import { Form,message,DatePicker,TimePicker } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  unitNumber: string;
};

export default function LogReplayModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const [form] = Form.useForm();

  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const [namespace, name] = getNSName();
    const res = await ReplayLogOfTenant({ namespace, name, ...values });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title="备租户日志回放"
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label="恢复日期"
          name="date"
          rules={[
            {
              required: true,
              message: '请选择恢复日期',
            },
          ]}
        >
          <DatePicker />
        </Form.Item>
        <Form.Item<FieldType>
          label="时分秒"
          name="time"
          rules={[
            {
              required: true,
              message: '请选择恢复时间',
            },
          ]}
        >
          <TimePicker />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
