import {
  generateRandomPassword,
  passwordRules,
} from '@/pages/Cluster/New/helper';
import { intl } from '@/utils/intl';
import { useUpdateEffect } from 'ahooks';
import { Button, Col, Form, Input, Row, Tooltip, message } from 'antd';
import type { FormInstance } from 'antd/lib/form';
import copy from 'copy-to-clipboard';
import { useState } from 'react';
import styles from './index.less';
interface PasswordInputProps {
  value: string;
  onChange: (val: string) => void;
  name: string;
  form: FormInstance<any>;
}
/**
 * 消费端：
 * 新增集群页 密码输入框
 * 新泽租户页 密码输入框
 */
export default function PasswordInput({
  value,
  onChange,
  name,
  form,
}: PasswordInputProps) {
  const { setFieldValue } = form;
  const [textVisile, setTextVisible] = useState<boolean>(false);
  const listenPasswordChange = async () => {
    try {
      await form.validateFields([name]);
      setTextVisible(true);
    } catch (err: any) {
      const { errorFields } = err;
      if (errorFields[0]?.errors.length) setTextVisible(false);
    }
  };
  const genaretaPassword = () => {
    const password = generateRandomPassword();
    onChange(password);
    setFieldValue('rootPassword', password);
    form.validateFields(['rootPassword']);
  };
  const passwordCopy = () => {
    if (value) {
      copy(value);
      message.success(
        intl.formatMessage({
          id: 'Dashboard.components.PasswordInput.CopiedSuccessfully',
          defaultMessage: '复制成功',
        }),
      );
    }
  };

  useUpdateEffect(() => {
    listenPasswordChange();
  }, [value]);
  return (
    <Tooltip
      color="#fff"
      overlayInnerStyle={{ color: 'rgba(0,0,0,.85)' }}
      overlayClassName={styles.toolTipContent}
      placement="bottomLeft"
      title={
        <ul>
          <li>
            {intl.formatMessage({
              id: 'Dashboard.components.PasswordInput.ToCharactersInLength',
              defaultMessage: '长度为 8~32 个字符',
            })}
          </li>
          <li>
            {intl.formatMessage({
              id: 'Dashboard.components.PasswordInput.CanOnlyContainLettersNumbers',
              defaultMessage: '只能包含字母、数字和特殊字符（~!@#%^&*_-+=|()',
            })}
            {}[]:;,.?/）
          </li>
          <li>
            {intl.formatMessage({
              id: 'Dashboard.components.PasswordInput.AtLeastUppercaseAndLowercase',
              defaultMessage: '大小写字母、数字和特殊字符都至少包含 2 个',
            })}
          </li>
        </ul>
      }
    >
      <Form.Item
        label={intl.formatMessage({
          id: 'Dashboard.components.PasswordInput.Password',
          defaultMessage: '密码',
        })}
        name="rootPassword"
        rules={passwordRules}
        className={styles.passwordFormItem}
        validateFirst
      >
        <Row gutter={8}>
          <Col style={{ flex: 1 }}>
            <Input.Password
              value={value}
              onChange={(val) => onChange(val.target.value)}
              placeholder={intl.formatMessage({
                id: 'Dashboard.components.PasswordInput.PleaseEnterOrGenerateRandomly',
                defaultMessage: '请输入或随机生成',
              })}
            />

            {textVisile && (
              <p style={{ color: 'rgba(0, 0, 0, 0.45)' }}>
                {intl.formatMessage({
                  id: 'Dashboard.components.PasswordInput.PleaseRememberThePasswordOr',
                  defaultMessage: '请牢记密码，也可',
                })}
                {' '}
                <a onClick={passwordCopy}>
                  {intl.formatMessage({
                    id: 'Dashboard.components.PasswordInput.CopyPassword',
                    defaultMessage: '复制密码',
                  })}
                </a>
                {' '}
                {intl.formatMessage({
                  id: 'Dashboard.components.PasswordInput.AndKeepItProperly',
                  defaultMessage: '并妥善保存',
                })}
              </p>
            )}
          </Col>
          <Col>
            <Button onClick={genaretaPassword}>
              {intl.formatMessage({
                id: 'Dashboard.components.PasswordInput.RandomlyGenerated',
                defaultMessage: '随机生成',
              })}
            </Button>
          </Col>
        </Row>
      </Form.Item>
    </Tooltip>
  );
}
