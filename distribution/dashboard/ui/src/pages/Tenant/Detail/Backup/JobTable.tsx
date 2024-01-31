import type { TableProps } from 'antd';
import { Table } from 'antd';

interface JobTableProps {
  dataSource: API.BackupJob[];
}

const columns: TableProps<API.BackupJob>['columns'] = [
  {
    title: 'name',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: 'path',
    dataIndex: 'path',
    key: 'path',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: '数据库状态',
    dataIndex: 'statusInDatabase',
    key: 'statusInDatabase',
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
  },
  {
    title: '加密密码',
    dataIndex: 'encryptionSecret',
    key: 'encryptionSecret',
  },
  {
    title: '开始时间',
    dataIndex: 'startTime',
    key: 'startTime',
  },
];

export default function JobTable({ dataSource }: JobTableProps) {
  return <Table columns={columns} dataSource={dataSource} />;
}
