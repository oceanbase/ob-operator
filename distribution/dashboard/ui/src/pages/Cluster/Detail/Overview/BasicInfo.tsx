import { colorMap } from '@/constants';
import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { Col, Descriptions, Tag } from 'antd';

interface BasicInfoProps {
  name: string;
  namespace: string;
  status: string;
  image: string;
  style?:any
}

export default function BasicInfo({
  name,
  namespace,
  status,
  image,
  style
}: BasicInfoProps) {
  return (
    <Col span={24}>
      <ProCard style={style}>
        <Descriptions
          column={5}
          title={intl.formatMessage({
            id: 'dashboard.Detail.Overview.BasicInformation',
            defaultMessage: '基本信息',
          })}
        >
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterName',
              defaultMessage: '集群名：',
            })}
          >
            {name}
          </Descriptions.Item>
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'OBDashboard.Detail.Overview.BasicInfo.Namespace',
              defaultMessage: '命名空间：',
            })}
          >
            {namespace}
          </Descriptions.Item>
          <Descriptions.Item
            span={2}
            label={intl.formatMessage({
              id: 'OBDashboard.Detail.Overview.BasicInfo.Image',
              defaultMessage: '镜像：',
            })}
          >
            {image}
          </Descriptions.Item>
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterStatus',
              defaultMessage: '集群状态',
            })}
          >
            <Tag color={colorMap.get(status)}>{status}</Tag>
          </Descriptions.Item>
        </Descriptions>
      </ProCard>
    </Col>
  );
}
