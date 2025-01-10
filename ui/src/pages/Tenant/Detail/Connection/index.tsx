import { info, terminal } from '@/api';
import { OBTerminal } from '@/components/Terminal/terminal';
import { getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { LinkOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Row, Space, message } from 'antd';
import React, { useEffect, useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';

const TenantConnection: React.FC = () => {
  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Connection',
        defaultMessage: '连接租户',
      }),
    };
  };

  const { ns, name } = useParams();

  const {
    data: tenantDetailResponse,
    run: getTenantDetail,
    loading,
  } = useRequest(getTenant, {
    manual: true,
  });

  const { data: dashboardInfo } = useRequest(info.getProcessInfo);

  useEffect(() => {
    getTenantDetail({ ns: ns!, name: name! });
  }, []);

  const [terminalId, setTerminalId] = useState<string>();

  const tenantDetail = tenantDetailResponse?.data;

  return (
    <PageContainer header={header()}>
      <link
        rel="stylesheet"
        href="https://cdn.jsdelivr.net/npm/xterm/css/xterm.css"
      />

      <Row gutter={[16, 16]}>
        {tenantDetail && (
          <Col span={24}>
            <BasicInfo
              info={tenantDetail.info}
              source={tenantDetail.source}
              loading={loading}
              ns={ns}
              name={name}
            />
          </Col>
        )}

        <div style={{ margin: 12, width: '100%' }}>
          {terminalId ? (
            <OBTerminal
              terminalId={terminalId}
              onClose={() => {
                setTerminalId(undefined);
                message.info(
                  intl.formatMessage({
                    id: 'Dashboard.Cluster.Detail.CloseConnection',
                    defaultMessage: '连接已关闭',
                  }),
                );
              }}
            />
          ) : (
            <Space>
              <Button
                onClick={async () => {
                  if (!tenantDetail || tenantDetail.info.status !== 'running') {
                    message.error(
                      intl.formatMessage({
                        id: 'Dashboard.Cluster.Detail.AbnormalOperation',
                        defaultMessage: '租户未正常运行',
                      }),
                    );
                    return;
                  }
                  const res = await terminal.createOBTenantConnection(
                    ns!,
                    name!,
                  );
                  if (res?.data?.terminalId) {
                    setTerminalId(res.data.terminalId);
                  }
                }}
              >
                {intl.formatMessage({
                  id: 'Dashboard.Cluster.Detail.CreateConnection',
                  defaultMessage: '创建连接',
                })}
              </Button>
              {dashboardInfo?.data.configurableInfo.odcURL && (
                <Button
                  onClick={async () => {
                    if (
                      !tenantDetail ||
                      tenantDetail.info.status !== 'running'
                    ) {
                      message.error(
                        intl.formatMessage({
                          id: 'Dashboard.Cluster.Detail.AbnormalOperation',
                          defaultMessage: '租户未正常运行',
                        }),
                      );
                      return;
                    }
                    const res = await terminal.createOBTenantConnection(
                      ns!,
                      name!,
                      'ODC',
                    );
                    if (res?.data?.odcConnectionURL) {
                      window.open(res.data.odcConnectionURL);
                    }
                  }}
                >
                  {intl.formatMessage({
                    id: 'src.pages.Tenant.Detail.Connection.E5D2E652',
                    defaultMessage: '通过 ODC 连接',
                  })}

                  <LinkOutlined />
                </Button>
              )}
            </Space>
          )}
        </div>
      </Row>
    </PageContainer>
  );
};

export default TenantConnection;
