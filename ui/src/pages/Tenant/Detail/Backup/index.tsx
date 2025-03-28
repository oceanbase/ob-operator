import { BACKUP_RESULT_STATUS, REFRESH_TENANT_TIME } from '@/constants';
import { getBackupPolicy, getTenant } from '@/services/tenant';
import { history, useAccess, useParams } from '@umijs/max';
import { Button, Card, Col, Form, Row, message } from 'antd';
import { useEffect, useRef, useState } from 'react';

import EmptyImg from '@/assets/empty.svg';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useRequest } from 'ahooks';
import { formatNewTenantForm } from '../../helper';
import BasicInfo from '../Overview/BasicInfo';
import BackupConfiguration from './BackupConfiguration';
import BackupJobs from './BackupJobs';

import { usePublicKey } from '@/hook/usePublicKey';
import { createTenantReportWrap } from '@/services/reportRequest/tenantReportReq';
import { strTrim } from '@/utils/helper';
import RecoverFormItem from '../NewBackup/RecoverFormItem';

export default function Backup() {
  const { ns, name, tenantName } = useParams();
  const access = useAccess();
  const [backupPolicy, setBackupPolicy] = useState<API.BackupPolicy>();
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const [isEditing, setIsEditing] = useState<boolean>(false);

  const [activeTabKey, setActiveTabKey] = useState<string>('backup');
  const [form] = Form.useForm<FormData>();
  const publicKey = usePublicKey();
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>([]);
  const [selectClusterId, setSelectClusterId] = useState<string>();
  const { name: clusterName, namespace } =
    clusterList.filter((cluster) => cluster.id === selectClusterId)[0] || {};

  const { refresh: backupPolicyRefresh, loading } = useRequest(
    getBackupPolicy,
    {
      defaultParams: [{ ns: ns!, name: name! }],
      pollingInterval: REFRESH_TENANT_TIME,
      ready:
        !backupPolicy ||
        (!isEditing && !BACKUP_RESULT_STATUS.includes(backupPolicy.status)),
      onSuccess: ({ successful, data }) => {
        if (successful) {
          setBackupPolicy(data);
        }
      },
    },
  );
  const { data: tenantDetailResponse } = useRequest(getTenant, {
    defaultParams: [{ ns: ns!, name: name! }],
  });
  const tenantDetail = tenantDetailResponse?.data;

  const { charset, deletionProtection, rootCredential, tenantRole, scenario } =
    tenantDetail?.info || {};

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, []);

  const tabList = [
    {
      key: 'backup',
      label: '备份',
    },
    {
      key: 'recover',
      label: '恢复',
    },
  ];

  const onTabChange = (key: string) => {
    setActiveTabKey(key);
    form.resetFields();
  };

  const {
    destType: type,
    archivePath: archiveSource,
    bakDataPath: bakDataSource,
    ossAccessSecret,
    bakEncryptionSecret,
  } = backupPolicy || {};
  const onFinish = () => {
    form.validateFields().then(async (values) => {
      const { source } = values;
      const obj = {
        charset,
        deletionProtection,
        rootCredential,
        namespace,
        tenantRole,
        scenario,
      };

      const restore = {
        restore: {
          ...source.restore,
          type,
          archiveSource,
          bakDataSource,
          ossAccessSecret,
          bakEncryptionSecret,
        },
      };
      if (source) {
        values.source = restore;
      }
      const reqData = formatNewTenantForm(
        strTrim({ ...values, ...obj }),
        clusterName,
        publicKey,
      );

      if (!reqData.pools?.length) {
        message.warning(
          intl.formatMessage({
            id: 'Dashboard.Tenant.New.SelectAtLeastOneZone',
            defaultMessage: '至少选择一个Zone',
          }),
        );
        return;
      }
      const res = await createTenantReportWrap({
        ...reqData,
      });
      if (res?.successful) {
        message.success('创建恢复租户成功', 3);
        form.resetFields();
        history.replace('/tenant');
      }
    });
  };

  const contentList: Record<string, React.ReactNode> = {
    backup: (
      <BackupConfiguration
        loading={loading}
        onDelete={() => {
          if (timerRef.current) {
            clearInterval(timerRef.current);
            timerRef.current = null;
          }
          setBackupPolicy((prev) => {
            if (!prev) {
              return prev;
            }
            return { ...prev, status: 'DELETING' };
          });
        }}
        isEditing={isEditing}
        setIsEditing={setIsEditing}
        backupPolicy={backupPolicy}
        setBackupPolicy={setBackupPolicy}
        backupPolicyRefresh={backupPolicyRefresh}
      />
    ),
    recover: (
      <Card title="创建恢复">
        <Form form={form}>
          <RecoverFormItem
            form={form}
            type="detail"
            clusterList={clusterList}
            setSelectClusterId={setSelectClusterId}
            setClusterList={setClusterList}
            selectClusterId={selectClusterId}
            onFinish={onFinish}
          />
        </Form>
      </Card>
    ),
  };

  return (
    <PageContainer>
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
            <p style={{ color: '#8592ad' }}>
              {intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Backup.5F7EA7F3',
                defaultMessage: '该租户尚未创建备份策略',
              })}
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
                ns={ns}
                name={name}
              />
            </Col>
          )}
          <Card
            style={{ marginTop: 24, marginBottom: 24 }}
            tabList={tabList}
            activeTabKey={activeTabKey}
            onTabChange={onTabChange}
            tabProps={{
              size: 'middle',
            }}
          >
            {contentList[activeTabKey]}
          </Card>
          {activeTabKey === 'backup' && <BackupJobs />}
        </Row>
      )}
    </PageContainer>
  );
}
