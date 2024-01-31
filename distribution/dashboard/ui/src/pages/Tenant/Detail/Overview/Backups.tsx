import { ProCard } from '@ant-design/pro-components';
import { Col, Descriptions, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface BackupsProps {
  backupPolicy: API.BackupPolicy;
  backupJobs: API.BackupJob[];
}

export default function Backups({ backupPolicy, backupJobs }: BackupsProps) {
  const PolicyConfig = {
    destType: 'destType',
    archivePath: '档案路径',
    bakDataPath: 'bakDataPath',
    scheduleType: '计划类型',
    scheduleDates: '计划日期',
    scheduleTime: '计划时间',
  };

  const columns: ColumnsType<API.BackupJob> = [
    {
      title: 'name',
      dataIndex: 'name',
      key: 'name',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
    },
    {
      title: '数据库中状态',
      dataIndex: 'statusInDatabase',
      key: 'statusInDatabase',
      width: 120,
    },
    {
      title: '类型',
      dataIndex: 'string',
      key: 'string',
      width: 120,
    },
    {
      title: 'path',
      dataIndex: 'path',
      key: 'path',
      width: 120,
    },
    {
      title: 'encryptionSecret',
      dataIndex: 'encryptionSecret',
      key: 'encryptionSecret',
      width: 120,
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 120,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      key: 'endTime',
      width: 120,
    },
  ];

  return (
    <Col span={24}>
      <ProCard title={<h2>备份</h2>} collapsible>
        <Descriptions column={5} title="备份策略">
          {Object.keys(PolicyConfig).map((key, index) => (
            <Descriptions.Item label={PolicyConfig[key]} key={index}>
              {key !== 'scheduleDates' ? (
                <span>{backupPolicy[key]}</span>
              ) : (
                <div>
                  {backupPolicy[key].map((item, index) => (
                    <span key={index}>
                      {item.backupType},{item.day}
                      {index !== backupPolicy[key].length}{' '}
                    </span>
                  ))}
                </div>
              )}
            </Descriptions.Item>
          ))}
        </Descriptions>
        <Table
          dataSource={backupJobs}
          rowKey="name"
          pagination={{ simple: true }}
          columns={columns}
        />
      </ProCard>
    </Col>
  );
}
