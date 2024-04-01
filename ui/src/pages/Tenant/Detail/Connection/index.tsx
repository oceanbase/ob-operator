import React, { useEffect, useState } from 'react'
import { PageContainer } from '@ant-design/pro-components'
import { intl } from '@/utils/intl'
import { OBTerminal } from '@/components/Terminal/terminal'
import { Button, Row, message } from 'antd'
import { request, useParams } from '@umijs/max'
import { useRequest } from 'ahooks'
import BasicInfo from '../Overview/BasicInfo'
import { getTenant } from '@/services/tenant'


const TenantConnection: React.FC = () => {
  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'dashboard.Tenant.Detail.Connection',
        defaultMessage: '连接租户',
      })
    }
  }

  const {ns, name} = useParams();

  const { data: tenantDetailResponse, run: getTenantDetail } = useRequest(getTenant, {
    manual: true,
  });

  const { runAsync } = useRequest(async (): Promise<{
    data: { terminalId: string }
  }> => {
    return request(`/api/v1/obtenants/${ns}/${name}/terminal`, {
      method: 'PUT'
    })
  }, {
    manual: true
  })

  useEffect(() => {
    getTenantDetail({ ns: ns!, name: name! });
  }, []);

  const [terminalId, setTerminalId] = useState<string>()

  const tenantDetail = tenantDetailResponse?.data;

  return (
    <PageContainer header={header()}>
      <link
        rel="stylesheet"
        href="https://cdn.jsdelivr.net/npm/xterm/css/xterm.css"
      />
      <Row gutter={[16, 16]}>
        {tenantDetail && (
          <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
        )}
        <div style={{margin: 12, width: '100%'}}>
          {terminalId ? (
            <OBTerminal terminalId={terminalId} onClose={() => {
              setTerminalId(undefined)
              message.info('连接已关闭')
            }} />
          ) : (
            <Button onClick={async () => {
              if (!tenantDetail || tenantDetail.info.status === 'failed') {
                message.error('租户未正常运行')
                return
              }
              const res = await runAsync()
              if (res?.data?.terminalId) {
                setTerminalId(res.data.terminalId)
              }
            }}>创建连接</Button>
          )}
        </div>
      </Row>
    </PageContainer>
  )
}

export default TenantConnection
