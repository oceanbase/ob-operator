import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { Form, Input, message } from 'antd';
import type { CommonModalType } from '.';
import { changeTenantPassword } from '@/services/tenant';
import CustomModal from '.';

type FieldType = {
  Password: string;
};

export default function ModifyPasswordModal({
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
    //User?
    const res = await changeTenantPassword({ namespace, name, ...values })
      if (res.successful) {
        message.success(res.message);
        successCallback();
        form.resetFields();
        setVisible(false);
      }
  };
  return (
    <CustomModal
      title="修改 root 密码"
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
          label="输入新密码"
          name="Password"
          rules={[
            {
              required: true,
              message: '请输入',
            },
          ]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
        <Form.Item<FieldType>
          label="再次输入"
          name="Password"
          rules={[
            {
              required: true,
              message: '请输入',
            },
          ]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
