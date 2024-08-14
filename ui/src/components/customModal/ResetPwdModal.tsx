import { access } from '@/api';
import type { ParamResetPasswordParam } from '@/api/generated';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { Form, Input, message } from 'antd';
import { omit } from 'lodash';
import CustomModal from '.';

interface ResetPwdModalProps {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  successCallback?: () => void;
}

export default function ResetPwdModal({
  visible,
  setVisible,
  successCallback,
}: ResetPwdModalProps) {
  const [form] = Form.useForm();
  const publicKey = usePublicKey();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const onFinish = async (
    values: ParamResetPasswordParam & { confirmPassword: string },
  ) => {
    values.oldPassword = encryptText(values.oldPassword!, publicKey) as string;
    values.password = encryptText(values.password!, publicKey) as string;
    const res = await access.resetPassword(omit(values, ['confirmPassword']));
    if (res.successful) {
      message.success('操作成功！');
      if (successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title="修改密码"
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item
          labelCol={{ span: 4 }}
          rules={[
            {
              required: true,
              message: '请输入',
            },
          ]}
          label="原密码"
          name={'oldPassword'}
        >
          <Input type="password" placeholder="请输入" />
        </Form.Item>
        <Form.Item
          validateFirst
          labelCol={{ span: 4 }}
          rules={[
            {
              required: true,
              message: '请输入',
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                const oldPwd = getFieldValue('oldPassword');
                if (oldPwd === value) {
                  return Promise.reject('新密码不能和原密码相同');
                }
                return Promise.resolve();
              },
            }),
          ]}
          label="新密码"
          name={'password'}
        >
          <Input type="password" placeholder="请输入" />
        </Form.Item>
        <Form.Item
          validateFirst
          labelCol={{ span: 4 }}
          rules={[
            {
              required: true,
              message: '请输入',
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                const newPwd = getFieldValue('password');
                if (newPwd !== value) {
                  return Promise.reject('两次密码输入不一致');
                }
                return Promise.resolve();
              },
            }),
          ]}
          label="确认密码"
          name={'confirmPassword'}
        >
          <Input type="password" placeholder="请输入" />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
