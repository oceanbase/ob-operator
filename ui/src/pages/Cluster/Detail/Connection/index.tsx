import { info, terminal } from '@/api';
import { OBTerminal } from '@/components/Terminal/terminal';
import { getClusterDetailReq } from '@/services';
import { intl } from '@/utils/intl';
import { LinkOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Row, Space, message } from 'antd';
import React, { useEffect, useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';

const ClusterConnection: React.FC = () => {
  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Cluster.Detail.Connection',
        defaultMessage: '集群连接',
      }),
    };
  };
  const { ns, name } = useParams();
  const access = useAccess();

  const { data: dashboardInfo } = useRequest(info.getProcessInfo);

  const { data: clusterDetail } = useRequest(getClusterDetailReq, {
    defaultParams: [{ name: name!, ns: ns! }],
  });

  const { runAsync } = useRequest(terminal.createOBClusterConnection, {
    manual: true,
  });

  const [terminalId, setTerminalId] = useState<string>();

  // 组件卸载时关闭连接
  useEffect(() => {
    return () => {
      if (terminalId) {
        // 这里可以调用关闭连接的API 通过设置terminalId为undefined来触发关闭
        setTerminalId(undefined);
      }
    };
  }, [terminalId, ns, name]);

  // 监听路由变化，关闭连接
  useEffect(() => {
    const handleRouteChange = () => {
      if (terminalId) {
        setTerminalId(undefined);
      }
    };

    // 监听 popstate 事件（浏览器前进后退）
    window.addEventListener('popstate', handleRouteChange);

    return () => {
      window.removeEventListener('popstate', handleRouteChange);
    };
  }, [terminalId]);

  return (
    <PageContainer header={header()}>
      <link
        rel="stylesheet"
        href="https://cdn.jsdelivr.net/npm/xterm/css/xterm.css"
      />

      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
          </Col>
        )}
        {access.obclusterwrite ? (
          <div style={{ margin: 12, width: '100%' }}>
            {terminalId ? (
              <OBTerminal
                terminalId={terminalId}
                onClose={() => {
                  setTerminalId(undefined);
                  if (terminalId) {
                    message.info(
                      intl.formatMessage({
                        id: 'Dashboard.Cluster.Detail.CloseConnection',
                        defaultMessage: '连接已关闭',
                      }),
                    );
                  }
                }}
              />
            ) : (
              <Space>
                <Button
                  onClick={async () => {
                    if (clusterDetail.info.status !== 'running') {
                      message.error(
                        intl.formatMessage({
                          id: 'Dashboard.Cluster.Detail.NotRunning',
                          defaultMessage: '集群未运行',
                        }),
                      );
                      return;
                    }
                    const res = await runAsync(ns!, name!);
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
                      if (clusterDetail.info.status !== 'running') {
                        message.error(
                          intl.formatMessage({
                            id: 'Dashboard.Cluster.Detail.NotRunning',
                            defaultMessage: '集群未运行',
                          }),
                        );
                        return;
                      }
                      const res = await terminal.createOBClusterConnection(
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
                      id: 'src.pages.Cluster.Detail.Connection.F8A13BAA',
                      defaultMessage: '通过 ODC 连接',
                    })}

                    <LinkOutlined />
                  </Button>
                )}
              </Space>
            )}
          </div>
        ) : null}
      </Row>
    </PageContainer>
  );
};

export default ClusterConnection;
