import CollapsibleCard from '@/components/CollapsibleCard';
import { SUFFIX_UNIT } from '@/constants';
import { MIRROR_MONITOR } from '@/constants/doc';
import { intl } from '@/utils/intl';
import {
Col,
Form,
Input,
InputNumber,
Tooltip,
} from 'antd';

const monitorTooltipText = intl.formatMessage({
  id: 'OBDashboard.Cluster.New.Monitor.TheImageShouldBeFully',
  defaultMessage:
    '镜像应写全 registry/image:tag，例如 oceanbase/obagent:4.2.0-100000062023080210',
});

export default function Monitor() {
  return (
    <Col span={24}>
      <CollapsibleCard title="监控" collapsible={true} bordered={false}>
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
                {intl.formatMessage({
                  id: 'Dashboard.Cluster.New.Monitor.Image',
                  defaultMessage: '镜像',
                })}{' '}
                <a href={MIRROR_MONITOR} rel="noreferrer" target="_blank">
                  {intl.formatMessage({
                    id: 'Dashboard.Cluster.New.Monitor.ImageList',
                    defaultMessage: '（镜像列表）',
                  })}
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
      </CollapsibleCard>
    </Col>
  );
}
