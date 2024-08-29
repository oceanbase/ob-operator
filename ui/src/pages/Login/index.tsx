import { intl } from '@/utils/intl';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { useModel, useNavigate } from '@umijs/max';
import { Button, Form, Input } from 'antd';
import React, { useState } from 'react';

import { user } from '@/api';
import logoSrc from '@/assets/oceanbase_logo.svg';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import styles from './index.less';

const Login: React.FC = () => {
  const { refresh } = useModel('@@initialState');
  const navigate = useNavigate();
  const publicKey = usePublicKey();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: API.User) => {
    setLoading(true);
    values.password = encryptText(values.password, publicKey) as string;
    try {
      const res = await user.login(values);
      if (res.successful) {
        // Set a timer to wait for permissions to update
        setTimeout(() => {
          if (res.data.needReset) {
            navigate('/reset');
          } else {
            navigate('/overview');
          }
          setLoading(false);
        }, 500);
        refresh();
        localStorage.setItem('user', values.username);
      }
    } catch (e) {
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
