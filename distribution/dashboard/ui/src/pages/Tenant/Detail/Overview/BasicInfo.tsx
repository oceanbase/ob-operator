import { colorMap } from '@/constants';
import { ProCard } from '@ant-design/pro-components';
import { Col, Descriptions, Row, Tag } from 'antd';

export default function BasicInfo({ info, source }: API.TenantBasicInfo) {
  const { style } = info;
  const InfoConfig = {
    name: '资源名',
    namespace: '命名空间',
    tenantName: '租户名',
    clusterName: '集群名',
    tenantRole: '租户角色',
    unitNumber: 'unit 数量',
    status: '状态',
    locality: '优先级',
    charset: '字符集',
  };
  const SourceConfig = {
    primaryTenant: '主租户',
    restoreType: '恢复源',
    archiveSource: '存档来源',
    bakDataSource: '数据源',
    until: 'until',
  };

  return (
    <Row style={{ marginBottom: 24 }} gutter={[16, 16]}>
      <Col span={24}>
        <ProCard style={style}>
          <Descriptions column={5} title="基本信息">
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
            <Descriptions title="租户资源">
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
