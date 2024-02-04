import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { getBackupJobs } from '@/services/tenant';
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
      label: '全量备份',
      children: <JobTable dataSource={backupJobs} />,
    },
    {
      key: 'INCR',
      label: '增量备份',
      children: <JobTable dataSource={backupJobs} />,
    },
    {
      key: 'ARCHIVE',
      label: '日志归档',
      children: <JobTable dataSource={backupJobs} />,
    },
    {
      key: 'CLEAN',
      label: '数据清理',
      children: <JobTable dataSource={backupJobs} />,
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
