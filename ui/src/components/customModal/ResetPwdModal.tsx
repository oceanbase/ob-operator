import { access } from '@/api';
import type { ParamResetPasswordParam } from '@/api/generated';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { intl } from '@/utils/intl';
import { Form, Input, message } from 'antd';
import { omit } from 'lodash';
import CustomModal from '.';

interface ResetPwdModalProps {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  showOriginPwd?: boolean;
  targetUser?: string;
  successCallback?: () => void;
}

export default function ResetPwdModal({
  visible,
  setVisible,
  showOriginPwd = true,
  targetUser,
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
    let res;
    if (showOriginPwd) {
      res = await access.resetPassword(omit(values, ['confirmPassword']));
    } else {
      // Administrator changes password
      res = await access.patchAccount(targetUser!, {
        password: values.password,
      });
    }
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'src.components.customModal.0D1428CF',
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
        id: 'src.components.customModal.BC49A0BD',
        defaultMessage: '修改密码',
      })}
      open={visible}
      onOk={handleSubmit}
      onCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
    >
      <Form form={form} onFinish={onFinish}>
        {showOriginPwd && (
          <Form.Item
            labelCol={{ span: 4 }}
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.components.customModal.D33CD1F7',
                  defaultMessage: '请输入',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'src.components.customModal.ADFC0106',
              defaultMessage: '原密码',
            })}
            name={'oldPassword'}
          >
            <Input.Password
              type="password"
              placeholder={intl.formatMessage({
                id: 'src.components.customModal.E97DEF21',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        )}

        <Form.Item
          validateFirst
          labelCol={{ span: 4 }}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.components.customModal.DF89BC3F',
                defaultMessage: '请输入',
              }),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                const oldPwd = getFieldValue('oldPassword');
                if (oldPwd === value) {
                  return Promise.reject(
                    intl.formatMessage({
                      id: 'src.components.customModal.FC52C4E2',
                      defaultMessage: '新密码不能和原密码相同',
                    }),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
          label={intl.formatMessage({
            id: 'src.components.customModal.7F950CE6',
            defaultMessage: '新密码',
          })}
          name={'password'}
        >
          <Input.Password
            type="password"
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.8BE441F0',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item
          validateFirst
          labelCol={{ span: 4 }}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.components.customModal.D7A47B92',
                defaultMessage: '请输入',
              }),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                const newPwd = getFieldValue('password');
                if (newPwd !== value) {
                  return Promise.reject(
                    intl.formatMessage({
                      id: 'src.components.customModal.B04BBA7E',
                      defaultMessage: '两次密码输入不一致',
                    }),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
          label={intl.formatMessage({
            id: 'src.components.customModal.B69BEEA1',
            defaultMessage: '确认密码',
          })}
          name={'confirmPassword'}
        >
          <Input.Password
            type="password"
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.69F2300D',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
