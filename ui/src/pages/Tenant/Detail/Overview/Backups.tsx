import CollapsibleCard from '@/components/CollapsibleCard';
import { WEEK_TEXT_MAP } from '@/constants/schedule';
import { intl } from '@/utils/intl';
import { Col, Descriptions, Table, Tooltip, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

interface BackupsProps {
  backupPolicy?: API.BackupPolicy;
  backupJobs?: API.BackupJob[];
  loading?: boolean;
}

const { Text } = Typography;

interface policyDisplay {
  key: keyof API.BackupPolicy;
  text: string;
  render?: (value: any) => React.ReactNode;
  span?: number;
}

export default function Backups({
  backupPolicy,
  backupJobs,
  loading,
}: BackupsProps) {
  const basics: policyDisplay[] = [
    {
      key: 'status',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.Status',
        defaultMessage: '备份状态',
      }),
      span: 3,
    },
    {
      key: 'destType',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.BackupMediaType',
        defaultMessage: '备份介质类型',
      }),
    },
    {
      key: 'archivePath',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.FilePath',
        defaultMessage: '日志归档路径',
      }),
      render: (value: string) => (
        <Text style={{ width: 300 }} ellipsis={{ tooltip: value }}>
          {value}
        </Text>
      ),
    },
    {
      key: 'bakDataPath',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.DataBackupPath',
        defaultMessage: '数据备份路径',
      }),
      render: (value: string) => (
        <Text style={{ width: 300 }} ellipsis={{ tooltip: value }}>
          {value}
        </Text>
      ),
    },
    {
      key: 'scheduleType',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.PlanType',
        defaultMessage: '计划类型',
      }),
      render: (value: string) => (
        <span>
          {value === 'Monthly'
            ? intl.formatMessage({
                id: 'Dashboard.Detail.Overview.Backups.MonthlyCycle',
                defaultMessage: '按月为周期',
              })
            : intl.formatMessage({
                id: 'Dashboard.Detail.Overview.Backups.CycleByWeek',
                defaultMessage: '按周为周期',
              })}
        </span>
      ),
    },
    {
      key: 'scheduleDates',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.PlannedDate',
        defaultMessage: '计划日期',
      }),
      render: (value: API.ScheduleDatesType) => {
        const fullArr = value.filter((item) => item.backupType === 'Full');
        const incrementalArr = value.filter(
          (item) => item.backupType === 'Incremental',
        );
        return (
          <div>
            {fullArr.length ? (
              <p>
                {backupPolicy?.scheduleType === 'Monthly' ? (
                  <>
                    {' '}
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Backups.FullBackupTheFirstOf',
                      defaultMessage: '全量备份：每个月第',
                    })}
                    {fullArr.map((item) => item.day).join(',')}
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Backups.Days',
                      defaultMessage: '天',
                    })}
                  </>
                ) : (
                  <>
                    {intl.formatMessage({
                      id: 'src.pages.Tenant.Detail.Overview.56424605',
                      defaultMessage: '全量备份：每周',
                    })}

                    {fullArr
                      .map((item) => WEEK_TEXT_MAP.get(item.day))
                      .join(',')}
                  </>
                )}
              </p>
            ) : null}

            {incrementalArr.length ? (
              <p>
                {backupPolicy?.scheduleType === 'Monthly' ? (
                  <>
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Backups.IncrementalBackupTheFirstOf',
                      defaultMessage: '增量备份：每个月第',
                    })}

                    {incrementalArr.map((item) => item.day).join(',')}
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Backups.Days',
                      defaultMessage: '天',
                    })}
                  </>
                ) : (
                  <>
                    {intl.formatMessage({
                      id: 'src.pages.Tenant.Detail.Overview.A2F7D655',
                      defaultMessage: '增量备份：每周',
                    })}

                    {incrementalArr
                      .map((item) => WEEK_TEXT_MAP.get(item.day))
                      .join(',')}
                  </>
                )}
              </p>
            ) : null}
          </div>
        );
      },
    },
    {
      key: 'scheduleTime',
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.ScheduledTime',
        defaultMessage: '计划时间',
      }),
    },
  ];

  const columns: ColumnsType<API.BackupJob> = [
    {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.Name',
        defaultMessage: '名称',
      }),
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
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Backups.Path',
        defaultMessage: '路径',
      }),
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
      <CollapsibleCard
        loading={loading}
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Backups.Backup',
              defaultMessage: '备份',
            })}
          </h2>
        }
        collapsible={true}
        defaultExpand={true}
      >
        {backupPolicy && (
          <Descriptions
            column={3}
            title={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Backups.BackupPolicy',
              defaultMessage: '备份策略',
            })}
          >
            {basics.map((item, index) => {
              return (
                <Descriptions.Item
                  label={item.text}
                  key={index}
                  span={item?.span ?? 1}
                >
                  {item.render ? (
                    item.render(backupPolicy[item.key])
                  ) : (
                    <span>{backupPolicy[item.key] as string}</span>
                  )}
                </Descriptions.Item>
              );
            })}
          </Descriptions>
        )}

        <Table
          dataSource={backupJobs}
          rowKey="name"
          pagination={{ simple: true }}
          columns={columns}
        />
      </CollapsibleCard>
    </Col>
  );
}
