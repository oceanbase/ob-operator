import CollapsibleCard from '@/components/CollapsibleCard';
import { intl } from '@/utils/intl';
import { Col, Form, Input, Row } from 'antd';

export default function BackUp() {
  return (
    <Col span={24}>
      <CollapsibleCard
        collapsible={true}
        title={intl.formatMessage({
          id: 'src.pages.Cluster.New.1E85AB3D',
          defaultMessage: '挂载 NFS 备份卷',
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
                  id: 'OBDashboard.Cluster.New.BackUp.PleaseEnter.A',
                  defaultMessage: '例如 172.17.x.x',
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
                  id: 'OBDashboard.Cluster.New.BackUp.PleaseEnter.B',
                  defaultMessage: '例如 /opt/nfs',
                })}
              />
            </Form.Item>
          </Col>
        </Row>
      </CollapsibleCard>
    </Col>
  );
}
