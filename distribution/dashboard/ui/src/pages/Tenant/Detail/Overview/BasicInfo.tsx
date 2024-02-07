import { colorMap } from '@/constants';
import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { Col, Descriptions, Row, Tag } from 'antd';

export default function BasicInfo({
  info,
  source,
  style,
}: API.TenantBasicInfo & { style?: React.CSSProperties }) {
  const InfoConfig = {
    name: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ResourceName',
      defaultMessage: '资源名',
    }),
    namespace: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Namespace',
      defaultMessage: '命名空间',
    }),
    tenantName: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.TenantName',
      defaultMessage: '租户名',
    }),
    clusterName: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ClusterName',
      defaultMessage: '集群名',
    }),
    tenantRole: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.TenantRole',
      defaultMessage: '租户角色',
    }),
    unitNumber: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.NumberOfUnits',
      defaultMessage: 'unit 数量',
    }),
    status: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Status',
      defaultMessage: '状态',
    }),
    locality: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Priority',
      defaultMessage: '优先级',
    }),
    charset: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.CharacterSet',
      defaultMessage: '字符集',
    }),
  };
  const SourceConfig = {
    primaryTenant: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.MasterTenant',
      defaultMessage: '主租户',
    }),
    restoreType: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.RecoverySource',
      defaultMessage: '恢复源',
    }),
    archiveSource: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ArchiveSource',
      defaultMessage: '存档来源',
    }),
    bakDataSource: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.DataSource',
      defaultMessage: '数据源',
    }),
    until: 'until',
  };

  return (
    <Row style={{ marginBottom: 24 }} gutter={[16, 16]}>
      <Col span={24}>
        <ProCard style={style}>
          <Descriptions
            column={5}
            title={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.BasicInformation',
              defaultMessage: '基本信息',
            })}
          >
            {Object.keys(InfoConfig).map(
              (key: keyof typeof InfoConfig, index) => {
                return (
                  <Descriptions.Item key={index} label={InfoConfig[key]}>
                    {key !== 'status' ? (
                      info[key]
                    ) : (
                      <Tag color={colorMap.get(info[key])}>{info[key]}</Tag>
                    )}
                  </Descriptions.Item>
                );
              },
            )}
          </Descriptions>
          {source && (
            <Descriptions
              title={intl.formatMessage({
                id: 'Dashboard.Detail.Overview.BasicInfo.TenantResources',
                defaultMessage: '租户资源',
              })}
            >
              {Object.keys(SourceConfig).map((key, index) => (
                <Descriptions.Item label={SourceConfig[key]} key={index}>
                  {source[key]}
                </Descriptions.Item>
              ))}
            </Descriptions>
          )}
        </ProCard>
      </Col>
    </Row>
  );
}
