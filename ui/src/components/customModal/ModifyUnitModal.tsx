import { useParams } from '@umijs/max';
import { modifyUnitNumReportWrap } from '@/services/reportRequest/tenantReportReq';
import { intl } from '@/utils/intl';
import { Form,InputNumber,message } from 'antd';
import { useEffect } from 'react';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  unitNum: string;
};

export default function ModifyUnitModal({
  visible,
  setVisible,
  params = {
    defaultUnitCount: 1,
  },
  successCallback,
}: CommonModalType & { params: { defaultUnitCount: number } }) {
  const [form] = Form.useForm();
  const { ns, name } = useParams();
  const { defaultUnitCount } = params;
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const res = await modifyUnitNumReportWrap({
      ns,
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
      if(successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  useEffect(() => {
    if (defaultUnitCount !== form.getFieldValue('unitNum')) {
      form.resetFields();
    }
  }, [defaultUnitCount]);

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
        initialValues={{ unitNum: defaultUnitCount }}
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
            min={1}
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
