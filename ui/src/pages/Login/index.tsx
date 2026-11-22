import { intl } from '@/utils/intl';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { useModel, useNavigate } from '@umijs/max';
import { Button, Form, Input, message } from 'antd';
import React, { useState } from 'react';

import { user } from '@/api';
import logoSrc from '@/assets/oceanbase_logo.svg';
import { encryptText } from '@/hook/usePublicKey';
import { getAppInfo } from '@/services';
import styles from './index.less';

const Login: React.FC = () => {
  const { refresh } = useModel('@@initialState');
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: API.User) => {
    setLoading(true);
    try {
      const { data } = await getAppInfo();
      const password = data?.publicKey && encryptText(values.password, data.publicKey);
      if (!password) {
        message.error(intl.formatMessage({
          id: 'dashboard.pages.Login.EncryptionFailed',
          defaultMessage: '密码加密失败，请重试',
        }));
        return;
      }
      const res = await user.login({ ...values, password });
      if (res.successful) {
        // Navigation must wait for account information and permissions, even
        // when the network takes longer than a fixed delay.
        await refresh();
        navigate(res.data.needReset ? '/reset' : '/overview');
        localStorage.setItem('user', values.username);
      }
    } catch (e) {
      // Request errors are reported by the global request handler.
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.loginContainer}>
      <div>
        <img className={styles.logo} src={logoSrc} alt="" />
      </div>
      <Form
        name="normal_login"
        className={styles.loginForm}
        initialValues={{ remember: true }}
        onFinish={onFinish}
      >
        <Form.Item
          name="username"
          rules={[{ required: true, message: 'Please input your Username!' }]}
        >
          <Input
            style={{ width: 300 }}
            prefix={<UserOutlined />}
            placeholder="Username"
          />
        </Form.Item>
        <Form.Item
          name="password"
          rules={[{ required: true, message: 'Please input your Password!' }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            type="password"
            placeholder="Password"
          />
        </Form.Item>

        <Form.Item style={{ display: 'flex', justifyContent: 'center' }}>
          <Button
            loading={loading}
            style={{ width: 270 }}
            type="primary"
            htmlType="submit"
          >
            {intl.formatMessage({
              id: 'dashboard.pages.Login.Login',
              defaultMessage: '登录',
            })}
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
};

export default Login;
