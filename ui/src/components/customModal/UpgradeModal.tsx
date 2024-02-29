import { intl } from '@/utils/intl';
import { Form, Input, message } from 'antd';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { upgradeObcluster } from '@/services';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  image: string;
};
export default function UpgradeModal({
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
  const onFinish = async ({ image }: any) => {
    const [ns, name] = getNSName();
    const res = await upgradeObcluster({ ns, name, image });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.UpgradeModal.VersionUpgrade',
        defaultMessage: '版本升级',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'OBDashboard.components.customModal.UpgradeModal.Image',
            defaultMessage: '镜像',
          })}
          name="image"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.UpgradeModal.PleaseEnterAnImage',
                defaultMessage: '请输入镜像!',
              }),
            },
          ]}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.UpgradeModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
