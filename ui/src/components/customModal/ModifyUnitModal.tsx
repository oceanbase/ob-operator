import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { patchTenantConfiguration } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { Form, InputNumber, message } from 'antd';
import { useEffect } from 'react';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  unitNum: string;
};

export default function ModifyUnitModal({
  visible,
  setVisible,
  defaultValue = 1,
  successCallback,
}: CommonModalType & { defaultValue: number }) {
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
    const res = await patchTenantConfiguration({
      ns: namespace,
      name,
      ...values,
    });
    if (res.successful) {
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyUnitModal.ModifiedSuccessfully',
            defaultMessage: '修改成功',
          }),
      );
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  useEffect(() => {
    if (defaultValue !== form.getFieldValue('unitNum')) {
      form.resetFields();
    }
  }, [defaultValue]);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ModifyUnitModal.ModifyTheNumberOfUnits',
        defaultMessage: '修改 Unit 数量',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        initialValues={{ unitNum: defaultValue }}
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyUnitModal.NumberOfUnits',
            defaultMessage: 'Unit 数量',
          })}
          name="unitNum"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyUnitModal.PleaseEnterTheNumberOf',
                defaultMessage: '请输入 Unit 数量',
              }),
            },
          ]}
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
