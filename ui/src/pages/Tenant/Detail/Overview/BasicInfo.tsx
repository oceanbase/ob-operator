import { COLOR_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Card, Col, Descriptions, Tag } from 'antd';

export default function BasicInfo({
  info = {},
  source = {},
  loading,
  style,
}: API.TenantBasicInfo & { style?: React.CSSProperties; loading: boolean }) {
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
    clusterResourceName: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ClusterName',
      defaultMessage: '集群名',
    }),
    tenantRole: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.TenantRole',
      defaultMessage: '租户角色',
    }),
    charset: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.CharacterSet',
      defaultMessage: '字符集',
    }),
    status: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Status',
      defaultMessage: '状态',
    }),
    unitNumber: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.NumberOfUnits',
      defaultMessage: 'unit 数量',
    }),
    locality: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ReplicaDistribution',
      defaultMessage: '副本分布',
    }),
    primaryZone: 'PrimaryZone',
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

  const checkSource = (source: any) => {
    Object.keys(source).forEach((key) => {
      if (source[key]) return true;
    });
    return false;
  };

  return (
    <Col span={24}>
      <Card
        loading={loading}
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.TenantBasicInformation',
              defaultMessage: '租户基本信息',
            })}
          </h2>
        }
        style={style}
      >
        <Descriptions column={5}>
          {Object.keys(InfoConfig).map(
            (key: keyof typeof InfoConfig, index) => {
              return (
                <Descriptions.Item key={index} label={InfoConfig[key]}>
                  {key !== 'status' ? (
                    info[key]
                  ) : (
                    <Tag color={COLOR_MAP.get(info[key])}>{info[key]}</Tag>
                  )}
                </Descriptions.Item>
              );
            },
          )}
        </Descriptions>
        {checkSource(source) && (
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
      </Card>
    </Col>
  );
}
