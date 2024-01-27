import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { patchTenantConfiguration } from '@/services/tenant';
import { Form, Input, message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  unitNum: string;
};

export default function ModifyUnitModal({
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
    const res = await patchTenantConfiguration({ namespace, name, ...values });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title="修改 Unit 数量"
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
          label="Unit 数量"
          name="unitNum"
          rules={[
            {
              required: true,
              message: '请输入 Unit 数量',
            },
          ]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
