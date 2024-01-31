import { SUFFIX_UNIT } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { patchTenantConfiguration } from '@/services/tenant';

import { Form, InputNumber, message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';
export default function ModifyUnitDetailModal({
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
      title="调整 Unit 规格"
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
        <Form.Item
          label="CPU"
          name="cpuCount"
          rules={[
            {
              required: true,
              message: '请输入 CPU 核数',
            },
          ]}
        >
          <InputNumber addonAfter='核' placeholder="请输入" />
        </Form.Item>
        <Form.Item
          label="Memory"
          name="memorySize"
          rules={[
            {
              required: true,
              message: '请输入 Memory',
            },
          ]}
        >
          <InputNumber addonAfter={SUFFIX_UNIT} placeholder="请输入" />
        </Form.Item>
        <Form.Item label="LogDiskSize" name="logDiskSize">
          <InputNumber addonAfter={SUFFIX_UNIT} placeholder="请输入" />
        </Form.Item>
        <Form.Item label="min iops" name="minIops">
          <InputNumber placeholder="请输入" />
        </Form.Item>
        <Form.Item label="max iops" name="maxIops">
          <InputNumber placeholder="请输入" />
        </Form.Item>
        <Form.Item label="iops权重" name="iopsWeight">
          <InputNumber placeholder="请输入" />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
