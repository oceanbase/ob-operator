import { Card, Col, Tabs } from 'antd';

import { getBackupJobs } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { TabsProps } from 'antd';
import { useState } from 'react';
import JobTable from './JobTable';

export default function BackupJobs() {
  const { ns, name } = useParams();
  const [curSelect, setCurSelect] = useState<API.JobType>('FULL');
  const { data: backupJobsResponse } = useRequest(
    () => {
      return getBackupJobs({ ns: ns!, name: name!, type: curSelect });
    },
    {
      refreshDeps: [curSelect],
      pollingInterval: 10000,
    },
  );

  const backupJobs = backupJobsResponse?.data || [];

  const items: TabsProps['items'] = [
    {
      key: 'FULL',
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupJobs.FullBackup',
        defaultMessage: '全量备份',
      }),
      children: <JobTable curSelect={curSelect} dataSource={backupJobs} />,
    },
    {
      key: 'INC',
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupJobs.IncrementalBackup',
        defaultMessage: '增量备份',
      }),
      children: <JobTable curSelect={curSelect} dataSource={backupJobs} />,
    },
    {
      key: 'ARCHIVE',
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupJobs.LogArchiving',
        defaultMessage: '日志归档',
      }),
      children: <JobTable curSelect={curSelect} dataSource={backupJobs} />,
    },
    {
      key: 'CLEAN',
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Backup.BackupJobs.DataCleansing',
        defaultMessage: '数据清理',
      }),
      children: <JobTable curSelect={curSelect} dataSource={backupJobs} />,
    },
  ];

  return (
    <Col span={24}>
      <Card>
        <Tabs
          defaultActiveKey="FULL"
          items={items}
          onChange={(key: string) => setCurSelect(key)}
        />
      </Card>
    </Col>
  );
}
