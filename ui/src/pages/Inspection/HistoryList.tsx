import { inspection } from '@/api';
import CustomTooltip from '@/components/CustomTooltip';
import { formatTime } from '@/utils/datetime';
import { intl } from '@/utils/intl';
import { theme } from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import { Link } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Table, Tag } from 'antd';

export default function HistoryList() {
  const { token } = theme.useToken();

  const { data: listInspectionReports, loading } = useRequest(
    inspection.listInspectionReports,
    {
      defaultParams: [{} as any],
    },
  );

  const dataSource = (listInspectionReports?.data || []).sort((a, b) => {
    // 按照 startTime 时间由近到远排列（最新的在前）
    return (b.startTime || 0) - (a.startTime || 0);
  });

  const statusList = [
    {
      label: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Success',
        defaultMessage: '成功',
      }),
      color: 'success',
      value: 'successful',
    },
    {
      label: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Failed',
        defaultMessage: '失败',
      }),
      color: 'error',
      value: 'failed',
    },
    {
      label: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Running',
        defaultMessage: '运行中',
      }),
      color: 'processing',
      value: 'running',
    },
    {
      label: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Pending',
        defaultMessage: '等待中',
      }),
      color: 'warning',
      value: 'pending',
    },
  ];

  const columns = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Task',
        defaultMessage: '任务',
      }),
      dataIndex: 'namespace',
      render: (text, record) => {
        return <CustomTooltip text={`${text}/${record?.name}`} width={100} />;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.ResourceName',
        defaultMessage: '资源名',
      }),
      dataIndex: 'obCluster',
      render: (text) => {
        return <span>{`${text?.namespace}/${text?.name}`}</span>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.ClusterName',
        defaultMessage: '集群名',
      }),
      dataIndex: ['obCluster', 'clusterName'],
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.InspectionScenario',
        defaultMessage: '巡检场景',
      }),
      dataIndex: 'scenario',
      render: (text) => {
        const color = text === 'basic' ? 'success' : 'processing';
        const content =
          text === 'basic'
            ? intl.formatMessage({
                id: 'src.pages.Inspection.BasicInspection',
                defaultMessage: '基础巡检',
              })
            : intl.formatMessage({
                id: 'src.pages.Inspection.PerformanceInspection',
                defaultMessage: '性能巡检',
              });

        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.StartTime',
        defaultMessage: '开始时间',
      }),
      dataIndex: 'startTime',
      sorter: (a: any, b: any) => (a.startTime || 0) - (b.startTime || 0),
      render: (text) => {
        return formatTime(text);
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.EndTime',
        defaultMessage: '结束时间',
      }),
      dataIndex: 'finishTime',
      sorter: (a: any, b: any) => (a.finishTime || 0) - (b.finishTime || 0),
      render: (text) => {
        return formatTime(text);
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.TaskStatus',
        defaultMessage: '任务状态',
      }),
      dataIndex: 'status',
      width: 100,
      render: (text) => {
        const content = findByValue(statusList, text);
        return <Tag color={content.color}>{content.label}</Tag>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.InspectionResult',
        defaultMessage: '巡检结果',
      }),
      dataIndex: 'resultStatistics',
      width: 120,
      render: (text) => {
        const { failedCount, criticalCount, moderateCount } = text || {};
        return (
          <div>
            <div style={{ color: token.colorError }}>
              {intl.formatMessage(
                {
                  id: 'src.pages.Inspection.HistoryList.FailedCount',
                  defaultMessage: '失败:{count}',
                },
                { count: failedCount || 0 },
              )}
            </div>
            <div style={{ color: 'rgba(166,29,36,1)' }}>
              {intl.formatMessage(
                {
                  id: 'src.pages.Inspection.HistoryList.HighRiskCount',
                  defaultMessage: '高风险:{count}',
                },
                { count: criticalCount || 0 },
              )}
            </div>
            <div style={{ color: 'orange' }}>
              {intl.formatMessage(
                {
                  id: 'src.pages.Inspection.HistoryList.ModerateRiskCount',
                  defaultMessage: '中风险:{count}',
                },
                { count: moderateCount || 0 },
              )}
            </div>
          </div>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.HistoryList.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'opeation',
      width: 100,
      render: (_: any, record: any) => {
        const id = `${record?.namespace}/${record?.name}`;
        return (
          <Link
            disabled={record?.status === 'running'}
            to={`/inspection/report/${id}`}
          >
            {intl.formatMessage({
              id: 'src.pages.Inspection.HistoryList.ViewReport',
              defaultMessage: '查看报告',
            })}
          </Link>
        );
      },
    },
  ];

  return <Table dataSource={dataSource} columns={columns} loading={loading} />;
}
