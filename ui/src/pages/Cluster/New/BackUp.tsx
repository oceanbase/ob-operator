import { intl } from '@/utils/intl';
import { Card, Col, Form, Input, Row, Space, Switch } from 'antd';
import { useEffect, useState } from 'react';

export default function BackUp() {
  const [isExpand, setIsExpand] = useState(false);

  useEffect(() => {
    let [cardBody] = document.querySelectorAll(
      '#backup-card .ant-card-body',
    ) as NodeListOf<HTMLElement>;
    if (!isExpand && cardBody) {
      cardBody.style.padding = '0px';
    } else {
      cardBody.style.padding = '24px';
    }
  }, [isExpand]);

  return (
    <Col span={24}>
      <Card
        title={
          <Space>
            <span>
              {intl.formatMessage({
                id: 'dashboard.Cluster.New.BackUp.BackupAndRecovery',
                defaultMessage: '备份恢复',
              })}
            </span>
            <Switch
              onChange={() => setIsExpand(!isExpand)}
              checked={isExpand}
            />
          </Space>
        }
        bordered={false}
        id="backup-card"
      >
        {isExpand && (
          <>
            <Row gutter={8}>
              <Col span={12}>
                <Form.Item
                  label={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BackUp.Address',
                    defaultMessage: '地址',
                  })}
                  name={['backupVolume', 'address']}
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'OBDashboard.Cluster.New.BackUp.PleaseEnterTheAddress',
                        defaultMessage: '请输入地址!',
                      }),
                    },
                  ]}
                >
                  <Input
                    placeholder={intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BackUp.PleaseEnter',
                      defaultMessage: '请输入',
                    })}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.BackUp.Path',
                    defaultMessage: '路径',
                  })}
                  name={['backupVolume', 'path']}
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'OBDashboard.Cluster.New.BackUp.EnterAPath',
                        defaultMessage: '请输入路径!',
                      }),
                    },
                  ]}
                >
                  <Input
                    placeholder={intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.BackUp.PleaseEnter',
                      defaultMessage: '请输入',
                    })}
                  />
                </Form.Item>
              </Col>
            </Row>
          </>
        )}
      </Card>
    </Col>
  );
}
