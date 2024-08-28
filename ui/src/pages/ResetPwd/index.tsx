import { access } from '@/api';
import logoSrc from '@/assets/oceanbase_logo.svg';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { intl } from '@/utils/intl';
import { useModel, useNavigate } from '@umijs/max';
import { Alert, Button, Form, Input } from 'antd';

import styles from './index.less';

interface FormData {
  password: string;
  passwordConfirm: string;
}

export default function ResetPwd() {
  const publicKey = usePublicKey();
  const navigate = useNavigate();
  const { setInitialState, initialState, refresh } = useModel('@@initialState');
  const onFinish = async (values: FormData) => {
    values.password = encryptText(values.password, publicKey) as string;
    const res = await access.resetPassword({ password: values.password });
    if (res.successful) {
      setInitialState({
        ...initialState,
        accountInfo: { ...initialState!.accountInfo, needReset: false },
      });
      await refresh();
      navigate('/overview');
    }
  };
  return (
    <div className={styles.loginContainer}>
      <div>
        <img className={styles.logo} src={logoSrc} alt="" />
      </div>
      <Alert
        className={styles.alertContent}
        message={intl.formatMessage({
          id: 'src.pages.ResetPwd.2768C822',
          defaultMessage:
            '您之前没有登录过，当前密码为默认密码，长期使用不安全。请先修改密码再继续使用 oceanbase dashboard！',
        })}
        type="warning"
      />

      <Form
        name="normal_login"
        className={styles.loginForm}
        initialValues={{ remember: true }}
        onFinish={onFinish}
      >
        <Form.Item
          name="password"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.ResetPwd.1BCDA661',
                defaultMessage: '请输入新密码',
              }),
            },
          ]}
        >
          <Input.Password
            type="password"
            placeholder={intl.formatMessage({
              id: 'src.pages.ResetPwd.EB8B735A',
              defaultMessage: '新密码',
            })}
          />
        </Form.Item>
        <Form.Item
          name="passwordConfirm"
          validateFirst
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.ResetPwd.D18DE694',
                defaultMessage: '请确认密码',
              }),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                const newPassword = getFieldValue('password');
                if (newPassword !== value) {
                  return Promise.reject(
                    intl.formatMessage({
                      id: 'src.pages.ResetPwd.C17C7281',
                      defaultMessage: '两次密码输入不一致',
                    }),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
        >
          <Input.Password
            type="password"
            placeholder={intl.formatMessage({
              id: 'src.pages.ResetPwd.0D23DF17',
              defaultMessage: '确认密码',
            })}
          />
        </Form.Item>

        <Form.Item style={{ display: 'flex', justifyContent: 'center' }}>
          <Button style={{ width: 270 }} type="primary" htmlType="submit">
            {intl.formatMessage({
              id: 'src.pages.ResetPwd.471AB584',
              defaultMessage: '确定',
            })}
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
}
