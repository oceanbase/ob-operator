import { inspection } from '@/api';
import { formatTime } from '@/utils/datetime';
import { theme } from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import { Link } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Table, Tag } from 'antd';

export default function HistoryList() {
  const { token } = theme.useToken();

  const { data: listInspectionReports } = useRequest(
    inspection.listInspectionReports,
    {
      defaultParams: [{}],
    },
  );

  console.log('listInspectionReports', listInspectionReports);

  const dataSource = listInspectionReports?.data || [];

  const statusList = [
    {
      label: '成功',
      color: 'success',
      value: 'successful',
    },
    {
      label: '失败',
      color: 'error',
      value: 'failed',
    },
    {
      label: '运行中',
      color: 'processing',
      value: 'running',
    },
    {
      label: '等待中',
      color: 'warning',
      value: 'pending',
    },
  ];

  const columns = [
    {
      title: '任务 ID',
      dataIndex: 'id',
    },
    {
      title: '资源名',
      dataIndex: 'obCluster',
      render: (text) => {
        return <span>{`${text?.namespace}/${text?.name}`}</span>;
      },
    },
    {
      title: '集群名',
      dataIndex: ['obCluster', 'clusterName'],
    },
    {
      title: '巡检场景',
      dataIndex: 'scenario',
      render: (text) => {
        const color = text === 'basic' ? 'success' : 'processing';
        const content = text === 'basic' ? '基础巡检' : '性能巡检';

        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      sorter: true,
      render: (text) => {
        return formatTime(text);
      },
    },
    {
      title: '结束时间',
      dataIndex: 'finishTime',
      sorter: true,
      render: (text) => {
        return formatTime(text);
      },
    },
    {
      title: '任务状态',
      dataIndex: 'status',
      render: (text) => {
        const content = findByValue(statusList, text);
        return <Tag color={content.color}>{content.label}</Tag>;
      },
    },
    {
      title: '巡检结果',
      dataIndex: 'resultStatistics',
      render: (text) => {
        const { failedCount, criticalCount, moderateCount } = text || {};
        return (
          <div>
            <div style={{ color: token.colorError }}>{`失败:${
              failedCount || 0
            }`}</div>
            <div style={{ color: 'purple' }}>{`高风险:${
              criticalCount || 0
            }`}</div>
            <div style={{ color: 'orange' }}>{`中风险:${
              moderateCount || 0
            }`}</div>
          </div>
        );
      },
    },
    {
      title: '操作',
      dataIndex: 'opeation',
      render: (text, record) => {
        console.log('record?.id', record);
        return (
          <Link to={`/inspection/report/${record?.id}`} target="_blank">
            查看报告
          </Link>
        );
      },
    },
  ];

  return <Table dataSource={dataSource} columns={columns} />;
}
