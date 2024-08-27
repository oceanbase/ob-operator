import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { changeTenantPassword } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { useParams } from '@umijs/max';
import { Form, Input, message } from 'antd';

import CustomModal from '.';

type FieldType = {
  Password: string;
};

export default function ModifyPasswordModal({
  visible,
  setVisible,
  successCallback,
}: API.CommonModalType) {
  const [form] = Form.useForm();
  const { ns, name } = useParams();
  const publicKey = usePublicKey();

  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const onCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const res = await changeTenantPassword({
      ns,
      name,
      User: 'root',
      Password: encryptText(values.password, publicKey) as string,
    });
    if (res.successful) {
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyPasswordModal.OperationSucceeded',
            defaultMessage: '操作成功！',
          }),
      );
      if (successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ModifyPasswordModal.ModifyRootPassword',
        defaultMessage: '修改 root 密码',
      })}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
    >
      <Form
        form={form}
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyPasswordModal.EnterANewPassword',
            defaultMessage: '输入新密码',
          })}
          name="password"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
                defaultMessage: '请输入',
              }),
            },
          ]}
        >
          <Input.Password
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyPasswordModal.EnterAgain',
            defaultMessage: '再次输入',
          })}
          name="passwordAgain"
          validateTrigger="onBlur"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
                defaultMessage: '请输入',
              }),
            },
            () => ({
              validator(_: any, value: string) {
                if (
                  form.getFieldValue('password') &&
                  value !== form.getFieldValue('password')
                ) {
                  return Promise.reject(
                    new Error(
                      intl.formatMessage({
                        id: 'Dashboard.components.customModal.ModifyPasswordModal.TheTwoInputsAreInconsistent',
                        defaultMessage: '两次输入不一致',
                      }),
                    ),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
        >
          <Input.Password
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
