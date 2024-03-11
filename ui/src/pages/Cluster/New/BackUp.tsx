import CollapsibleCard from '@/components/CollapsibleCard';
import { intl } from '@/utils/intl';
import { Col, Form, Input, Row } from 'antd';

export default function BackUp() {
  return (
    <Col span={24}>
      <CollapsibleCard
        collapsible={true}
        title={intl.formatMessage({
          id: 'Dashboard.Cluster.New.BackUp.BackupAndRecovery',
          defaultMessage: '备份恢复',
        })}
        bordered={false}
      >
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
      </CollapsibleCard>
    </Col>
  );
}
