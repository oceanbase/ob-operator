import { intl } from '@/utils/intl';
import { Card, Col, Form, Input, InputNumber, Row, Space, Switch } from 'antd';
import { useState } from 'react';
import styles from './index.less';

interface AdvancedConfigurationProps {
  open?: boolean;
  disable?: boolean;
}

const { Password } = Input;

export default function AdvancedConfiguration({
  open = false,
  disable = false,
}: AdvancedConfigurationProps) {
  const [isExpand, setIsExpand] = useState(open);
  const [passInputExpand, setPassInputExpand] = useState(open);
  const [jobInputExpand, setJobInputExpand] = useState(open);
  const [recoveryInputExpand, setRecoveryInputExpand] = useState(open);
  return (
    <Card
      title={
        <Space>
          <span>
            {intl.formatMessage({
              id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.AdvancedConfiguration',
              defaultMessage: '高级配置',
            })}
          </span>
          <Switch onChange={() => setIsExpand(!isExpand)} checked={isExpand} />
        </Space>
      }
      bordered={false}
    >
      {isExpand && (
        <Row>
          <Col className={styles.column} span={24}>
            <label className={styles.labelText}>
              {intl.formatMessage({
                id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.EncryptedBackup',
                defaultMessage: '加密备份：',
              })}
            </label>
            <Switch
              className={styles.switch}
              onChange={() => setPassInputExpand(!passInputExpand)}
              checked={passInputExpand}
            />

            {passInputExpand && (
              <Form.Item
                label={intl.formatMessage({
                  id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.EncryptedPassword',
                  defaultMessage: '加密密码',
                })}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.EnterAnEncryptionPassword',
                      defaultMessage: '请输入加密密码',
                    }),
                  },
                ]}
                name={['bakEncryptionPassword']}
              >
                <Password
                  disabled={disable}
                  style={{ width: 216 }}
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </Form.Item>
            )}
          </Col>
          <Col className={styles.column} span={24}>
            <label className={styles.labelText}>
              {intl.formatMessage({
                id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.BackupTaskRetention',
                defaultMessage: '备份任务保留：',
              })}
            </label>
            <Switch
              className={styles.switch}
              onChange={() => setJobInputExpand(!jobInputExpand)}
              checked={jobInputExpand}
            />

            {jobInputExpand && (
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.PleaseEnterTheRetentionDays',
                      defaultMessage: '请输入备份任务保留天数',
                    }),
                  },
                ]}
                name={['jobKeepDays']}
              >
                <InputNumber min={1} disabled={disable} />
              </Form.Item>
            )}
          </Col>
          <Col className={styles.column} span={24}>
            <label className={styles.labelText}>
              {intl.formatMessage({
                id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.DataRecoveryWindow',
                defaultMessage: '数据恢复窗口：',
              })}
            </label>
            <Switch
              className={styles.switch}
              onChange={() => setRecoveryInputExpand(!recoveryInputExpand)}
              checked={recoveryInputExpand}
            />

            {recoveryInputExpand && (
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.EnterTheDataRecoveryWindow',
                      defaultMessage: '请输入数据恢复窗口：',
                    }),
                  },
                ]}
                name={['recoveryDays']}
              >
                <InputNumber min={1} disabled={disable} />
              </Form.Item>
            )}
          </Col>
          <Form.Item
            name={['pieceIntervalDays']}
            label={intl.formatMessage({
              id: 'Dashboard.Detail.NewBackup.AdvancedConfiguration.ArchiveSliceInterval',
              defaultMessage: '归档切片间隔',
            })}
          >
            <InputNumber disabled={disable} min={1} max={7} />
          </Form.Item>
        </Row>
      )}
    </Card>
  );
}
