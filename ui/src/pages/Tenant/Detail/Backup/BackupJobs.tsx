import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { getBackupJobs } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import type { TabsProps } from 'antd';
import { Card, Col, Tabs } from 'antd';
import { useState } from 'react';
import JobTable from './JobTable';

export default function BackupJobs() {
  const [curSelect, setCurSelect] = useState<API.JobType>('FULL');
  const [ns, name] = getNSName();
  const { data: backupJobsResponse } = useRequest(
    () => {
      return getBackupJobs({ ns, name, type: curSelect });
    },
    {
      refreshDeps: [curSelect],
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
      key: 'INCR',
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
