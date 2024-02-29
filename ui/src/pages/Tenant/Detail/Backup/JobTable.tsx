import { intl } from '@/utils/intl';
import type { TableProps } from 'antd';
import { Table, Tooltip } from 'antd';

interface JobTableProps {
  dataSource: API.BackupJob[];
}

const columns: TableProps<API.BackupJob>['columns'] = [
  {
    title: 'name',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
    render: (value) => {
      return (
        <Tooltip placement='topLeft' title={value}>
          <span>{value}</span>
        </Tooltip>
      );
    },
  },
  {
    title: 'path',
    dataIndex: 'path',
    key: 'path',
    ellipsis: true,
    render: (value) => {
      return (
        <Tooltip placement='topLeft' title={value}>
          <span>{value}</span>
        </Tooltip>
      );
    },
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.DatabaseStatus',
      defaultMessage: '数据库状态',
    }),
    dataIndex: 'statusInDatabase',
    key: 'statusInDatabase',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.Type',
      defaultMessage: '类型',
    }),
    dataIndex: 'type',
    key: 'type',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.EncryptedPassword',
      defaultMessage: '加密密码',
    }),
    dataIndex: 'encryptionSecret',
    key: 'encryptionSecret',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Detail.Backup.JobTable.StartTime',
      defaultMessage: '开始时间',
    }),
    dataIndex: 'startTime',
    key: 'startTime',
  },
];

export default function JobTable({ dataSource }: JobTableProps) {
  return <Table columns={columns} dataSource={dataSource} />;
}
