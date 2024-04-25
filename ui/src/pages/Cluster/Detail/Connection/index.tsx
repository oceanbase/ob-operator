import { OBTerminal } from '@/components/Terminal/terminal';
import { getClusterDetailReq } from '@/services';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { request, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Row, message } from 'antd';
import React, { useState } from 'react';
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

  const { data: clusterDetail } = useRequest(getClusterDetailReq, {
    defaultParams: [{ name: name!, ns: ns! }],
  });

  const { runAsync } = useRequest(
    async (): Promise<{
      data: { terminalId: string };
    }> => {
      return request(
        `/api/v1/obclusters/namespace/${ns}/name/${name}/terminal`,
        {
          method: 'PUT',
        },
      );
    },
    {
      manual: true,
    },
  );

  const [terminalId, setTerminalId] = useState<string>();

  return (
    <PageContainer header={header()}>
      <link
        rel="stylesheet"
        href="https://cdn.jsdelivr.net/npm/xterm/css/xterm.css"
      />
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
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
            <Button
              onClick={async () => {
                if (
                  (clusterDetail.info as API.ClusterInfo).status !== 'running'
                ) {
                  message.error(
                    intl.formatMessage({
                      id: 'Dashboard.Cluster.Detail.NotRunning',
                      defaultMessage: '集群未运行',
                    }),
                  );
                  return;
                }
                const res = await runAsync();
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
          )}
        </div>
      </Row>
    </PageContainer>
  );
};

export default ClusterConnection;
