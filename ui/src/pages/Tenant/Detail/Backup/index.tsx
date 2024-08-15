import EmptyImg from '@/assets/empty.svg';
import { BACKUP_RESULT_STATUS, REFRESH_TENANT_TIME } from '@/constants';
import { getBackupPolicy, getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { history, useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Col, Row } from 'antd';
import { useEffect, useRef, useState } from 'react';
import BasicInfo from '../Overview/BasicInfo';
import BackupConfiguration from './BackupConfiguration';
import BackupJobs from './BackupJobs';

export default function Backup() {
  const { ns, name, tenantName } = useParams();
  const access = useAccess();
  const [backupPolicy, setBackupPolicy] = useState<API.BackupPolicy>();
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const { refresh: backupPolicyRefresh, loading } = useRequest(
    getBackupPolicy,
    {
      defaultParams: [{ ns: ns!, name: name! }],
      onSuccess: ({ successful, data }) => {
        if (successful) {
          setBackupPolicy(data);
          if (!BACKUP_RESULT_STATUS.includes(data.status)) {
            if (!timerRef.current) {
              timerRef.current = setInterval(() => {
                backupPolicyRefresh();
              }, REFRESH_TENANT_TIME);
            }
          } else if (timerRef.current) {
            clearInterval(timerRef.current);
            timerRef.current = null;
          }
        }
      },
    },
  );
  const { data: tenantDetailResponse } = useRequest(getTenant, {
    defaultParams: [{ ns: ns!, name: name! }],
  });
  const tenantDetail = tenantDetailResponse?.data;

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, []);

  return (
    <PageContainer loading={loading}>
      {!backupPolicy ? (
        <Card
          style={{
            height: 'calc(100vh - 98px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          bodyStyle={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <img
            src={EmptyImg}
            alt="empty"
            style={{ marginBottom: 24, height: 100, width: 110 }}
          />
          {access.obclusterwrite ? (
            <>
              <p style={{ color: '#8592ad', marginBottom: 24 }}>
                {intl.formatMessage({
                  id: 'Dashboard.Detail.Backup.TheTenantHasNotCreated',
                  defaultMessage: '该租户尚未创建备份策略，是否立即创建？',
                })}
              </p>
              <Button
                type="primary"
                onClick={() =>
                  history.push(`/tenant/${ns}/${name}/${tenantName}/backup/new`)
                }
              >
                {intl.formatMessage({
                  id: 'Dashboard.Detail.Backup.CreateNow',
                  defaultMessage: '立即创建',
                })}
              </Button>
            </>
          ) : (
            <p style={{ color: '#8592ad'}}>
              该租户尚未创建备份策略
            </p>
          )}
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          {tenantDetail && (
            <Col span={24}>
              <BasicInfo
                info={tenantDetail.info}
                source={tenantDetail.source}
              />
            </Col>
          )}
          <Col span={24}>
            <BackupConfiguration
              backupPolicy={backupPolicy}
              setBackupPolicy={setBackupPolicy}
              backupPolicyRefresh={backupPolicyRefresh}
            />
          </Col>
          <BackupJobs />
        </Row>
      )}
    </PageContainer>
  );
}
