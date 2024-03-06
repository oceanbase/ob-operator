import { intl } from '@/utils/intl';
import { Col, Descriptions, Table, Tooltip, Card } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface BackupsProps {
  backupPolicy: API.BackupPolicy;
  backupJobs: API.BackupJob[];
}

export default function Backups({ backupPolicy, backupJobs }: BackupsProps) {
  const PolicyConfig = {
    destType: 'destType',
    archivePath: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Backups.FilePath',
      defaultMessage: '档案路径',
    }),
    bakDataPath: 'bakDataPath',
    scheduleType: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Backups.PlanType',
      defaultMessage: '计划类型',
    }),
    scheduleDates: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Backups.PlannedDate',
      defaultMessage: '计划日期',
    }),
    scheduleTime: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Backups.ScheduledTime',
      defaultMessage: '计划时间',
    }),
  };

  const columns: ColumnsType<API.BackupJob> = [
    {
      title: 'name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.Status',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      key: 'status',
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.StatusInTheDatabase',
        defaultMessage: '数据库中状态',
      }),
      dataIndex: 'statusInDatabase',
      key: 'statusInDatabase',
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.Type',
        defaultMessage: '类型',
      }),
      dataIndex: 'string',
      key: 'string',
    },
    {
      title: 'path',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (val) => (
        <Tooltip placement="topLeft" title={val}>
          <span>{val}</span>
        </Tooltip>
      ),
    },
    {
      title: 'encryptionSecret',
      dataIndex: 'encryptionSecret',
      key: 'encryptionSecret',
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.StartTime',
        defaultMessage: '开始时间',
      }),
      dataIndex: 'startTime',
      key: 'startTime',
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.EndTime',
        defaultMessage: '结束时间',
      }),
      dataIndex: 'endTime',
      key: 'endTime',
    },
  ];

  return (
    <Col span={24}>
      <Card
        title={
          <h2>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Backups.Backup',
              defaultMessage: '备份',
            })}
          </h2>
        }
        collapsible
      >
        <Descriptions
          column={5}
          title={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.Backups.BackupPolicy',
            defaultMessage: '备份策略',
          })}
        >
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
      </Card>
    </Col>
  );
}
