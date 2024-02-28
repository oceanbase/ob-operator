import { intl } from '@/utils/intl';
import {
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Space,
  Switch,
  Tooltip,
} from 'antd';
import { useEffect, useState } from 'react';
import { MIRROR_MONITOR } from '@/constants/doc';
import { SUFFIX_UNIT } from '@/constants';

const monitorTooltipText = intl.formatMessage({
  id: 'OBDashboard.Cluster.New.Monitor.TheImageShouldBeFully',
  defaultMessage:
    '镜像应写全 registry/image:tag，例如 oceanbase/obagent:4.2.0-100000062023080210',
});

export default function Monitor() {
  const [isExpand, setIsExpand] = useState(false);
  useEffect(() => {
    let [cardBody] = document.querySelectorAll(
      '#monitor-card .ant-card-body',
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
                id: 'dashboard.Cluster.New.Monitor.Monitoring',
                defaultMessage: '监控',
              })}
            </span>
            <Switch
              onChange={() => setIsExpand(!isExpand)}
              checked={isExpand}
            />
          </Space>
        }
        bordered={false}
        id="monitor-card"
      >
        {isExpand && (
          <>
            <Tooltip title={monitorTooltipText}>
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.Monitor.EnterAnImage',
                      defaultMessage: '请输入镜像',
                    }),
                  },
                ]}
                style={{ width: '50%' }}
                label={
                  <>
                    镜像{' '}
                    <a
                      href={MIRROR_MONITOR}
                      rel="noreferrer"
                      target="_blank"
                    >
                      （镜像列表）
                    </a>
                  </>
                }
                name={['monitor', 'image']}
              >
                <Input
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Monitor.EnterAnImage',
                    defaultMessage: '请输入镜像',
                  })}
                />
              </Form.Item>
            </Tooltip>

            <h1>
              {intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Monitor.Resources',
                defaultMessage: '资源',
              })}
            </h1>
            <div style={{ display: 'flex' }}>
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.Monitor.EnterTheNumberOfCpus',
                      defaultMessage: '请输入cpu数量',
                    }),
                  },
                ]}
                style={{ marginRight: 16 }}
                label={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.Monitor.NumberOfCpus',
                  defaultMessage: 'cpu数量',
                })}
                name={['monitor', 'resource', 'cpu']}
              >
                <InputNumber
                  min={1}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Monitor.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </Form.Item>
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.Monitor.EnterMemory',
                      defaultMessage: '请输入内存',
                    }),
                  },
                ]}
                label={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.Monitor.Memory',
                  defaultMessage: '内存',
                })}
                name={['monitor', 'resource', 'memory']}
              >
                <InputNumber
                  min={1}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Monitor.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </Form.Item>
            </div>
          </>
        )}
      </Card>
    </Col>
  );
}
