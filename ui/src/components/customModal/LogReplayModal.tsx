import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { replayLogOfTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { DatePicker, Form, TimePicker, message } from 'antd';
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
    const res = await replayLogOfTenant({ namespace, name, ...values });
    if (res.successful) {
      message.success(res.message);
      if(successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.LogReplayModal.StandbyTenantLogPlayback',
        defaultMessage: '备租户日志回放',
      })}
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
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.LogReplayModal.RecoveryDate',
            defaultMessage: '恢复日期',
          })}
          name="date"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.LogReplayModal.SelectARecoveryDate',
                defaultMessage: '请选择恢复日期',
              }),
            },
          ]}
        >
          <DatePicker />
        </Form.Item>
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.LogReplayModal.MinutesAndSeconds',
            defaultMessage: '时分秒',
          })}
          name="time"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.LogReplayModal.SelectARecoveryTime',
                defaultMessage: '请选择恢复时间',
              }),
            },
          ]}
        >
          <TimePicker />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
