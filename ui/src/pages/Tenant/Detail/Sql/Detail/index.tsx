import { DATE_TIME_FORMAT, DateSelectOption } from '@/constants/datetime';
import {
  listSqlMetrics,
  listSqlStats,
  queryPlanDetailInfo,
  querySqlDetailInfo,
  querySqlHistoryInfo,
} from '@/services/sql';
import { intl } from '@/utils/intl';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { ProCard, ProDescriptions, ProTable } from '@ant-design/pro-components';
import { Line } from '@antv/g2plot';
import {
  history,
  useLocation,
  useParams,
  useRequest,
  useSearchParams,
} from '@umijs/max';
import {
  Button,
  DatePicker,
  Drawer,
  Select,
  Space,
  Tag,
  Typography,
} from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import dayjs from 'dayjs';
import { useEffect, useRef, useState } from 'react';
import { getLocale } from 'umi';

const { RangePicker } = DatePicker;
const { Title } = Typography;

// --- Chart Component ---

interface SqlTrendChartProps {
  data: API.MetricData[];
  type: 'execution' | 'latency';
  height?: number;
}

const SqlTrendChart: React.FC<SqlTrendChartProps> = ({
  data,
  type,
  height = 300,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<Line>();

  useEffect(() => {
    if (!containerRef.current || !data || data.length === 0) return;

    // Transform data for G2Plot
    // API.MetricData: { metric: { name, labels }, values: { timestamp, value }[] }
    // G2Plot expects array of objects
    const plotData: any[] = [];
    data.forEach((metricData) => {
      const name = metricData?.metric?.name || type;
      metricData?.values?.forEach((v) => {
        plotData.push({
          time: v.timestamp * 1000, // Convert to ms
          value: v.value,
          category: name,
        });
      });
    });

    // sort by time
    plotData.sort((a, b) => a.time - b.time);

    if (chartRef.current) {
      chartRef.current.destroy();
      chartRef.current = undefined;
    }

    const chart = new Line(containerRef.current, {
      data: plotData,
      xField: 'time',
      yField: 'value',
      seriesField: 'category',
      height,
      xAxis: {
        type: 'time',
        mask: 'HH:mm',
      },
      yAxis: {
        label: {
          formatter: (v: string) => {
            return Number(v).toFixed(2);
          },
        },
      },
      tooltip: {
        formatter: (datum: any) => {
          return { name: datum.category, value: datum.value.toFixed(2) };
        },
      },
      legend: {
        position: 'top',
      },
    });

    chart.render();
    chartRef.current = chart;

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
        chartRef.current = undefined;
      }
    };
  }, [data, type, height]);

  return <div ref={containerRef} style={{ height }} />;
};

// --- Main Page Component ---

const SqlDetail: React.FC = () => {
  const { ns, name, sqlId } = useParams<{
    ns: string;
    name: string;
    sqlId: string;
  }>();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const stateSqlMeta = location.state as API.SqlInfo | undefined;

  const dbName = searchParams.get('dbName') || '';
  const urlStartTime = searchParams.get('startTime');
  const urlEndTime = searchParams.get('endTime');

  // Time range for history trend
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    urlStartTime
      ? dayjs.unix(Number(urlStartTime))
      : dayjs().subtract(30, 'minute'),
    urlEndTime ? dayjs.unix(Number(urlEndTime)) : dayjs(),
  ]);

  const [latencyMetricsMeta, setLatencyMetricsMeta] = useState<
    API.SqlMetricMeta[]
  >([]);
  const [selectedLatencyMetrics, setSelectedLatencyMetrics] = useState<
    string[]
  >(['elapsed_time']);

  // Fetch metrics meta to populate latency selector
  useRequest(
    () =>
      listSqlMetrics({ language: getLocale() === 'zh-CN' ? 'zh_CN' : 'en_US' }),
    {
      onSuccess: (data) => {
        const list = Array.isArray(data) ? data : (data as any)?.data || [];
        const latencyCat = list.find((c: any) => c.category === 'latency');
        if (latencyCat && latencyCat.metrics) {
          setLatencyMetricsMeta(latencyCat.metrics);
          // Set default selected metrics based on displayByDefault
          const defaults = latencyCat.metrics
            .filter((m: any) => m.displayByDefault)
            .map((m: any) => m.key);
          if (defaults.length > 0) {
            setSelectedLatencyMetrics(defaults);
          } else {
            // Fallback to first if no defaults
            if (latencyCat.metrics.length > 0) {
              setSelectedLatencyMetrics([latencyCat.metrics[0].key]);
            }
          }
        }
      },
    },
  );

  // 1. Fetch Static Detail Info (Plans, Indexes, etc.)
  // Use a wider range or the initial range to find the SQL text
  const { data: detailData, loading: detailLoading } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      // We use the initial time range from URL or a default wide window to ensure we find the SQL text/plans
      // For simplicity, we can use the current timeRange, but ideally this shouldn't reload when user zooms in trend.
      // However, if we want to isolate it, we can use url params.
      const start = urlStartTime
        ? Number(urlStartTime)
        : dayjs().subtract(24, 'hour').unix();
      const end = urlEndTime ? Number(urlEndTime) : dayjs().unix();

      return querySqlDetailInfo({
        namespace: ns,
        obtenant: name,
        sqlId,
        database: dbName,
        startTime: start,
        endTime: end,
        interval: 60,
      } as any);
    },
    {
      refreshDeps: [ns, name, sqlId, dbName], // Only reload if identity changes
    },
  );

  // 2. Fetch History Trend Info
  const { data: historyData, loading: historyLoading } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      const start = timeRange[0].unix();
      const end = timeRange[1].unix();
      // Calculate interval
      let interval = Math.floor((end - start) / 60);
      if (interval < 1) interval = 1;

      return querySqlHistoryInfo({
        namespace: ns,
        obtenant: name,
        sqlId,
        database: dbName,
        startTime: start,
        endTime: end,
        interval,
        outputColumns: selectedLatencyMetrics,
      });
    },
    {
      refreshDeps: [timeRange, selectedLatencyMetrics, ns, name, sqlId, dbName],
    },
  );

  // Time Range Helpers
  const range = (start: number, end: number) => {
    const result = [];
    for (let i = start; i < end; i++) {
      result.push(i);
    }
    return result;
  };

  const disabledDateTime: RangePickerProps['disabledTime'] = (_) => {
    const isToday = _?.date() === dayjs().date();
    if (!isToday)
      return {
        disabledHours: () => [],
        disabledMinutes: () => [],
        disabledSeconds: () => [],
      };
    return {
      disabledHours: () => range(0, 24).splice(dayjs().hour() + 1, 24),
      disabledMinutes: (hour) => {
        if (hour === dayjs().hour()) {
          return range(0, 60).splice(dayjs().minute() + 1, 60);
        }
        return [];
      },
      disabledSeconds: (hour, minute) => {
        if (hour === dayjs().hour() && minute === dayjs().minute()) {
          return range(0, 60).splice(dayjs().second(), 60);
        }
        return [];
      },
    };
  };

  const disabledDate: RangePickerProps['disabledDate'] = (current) => {
    return current && current > dayjs().endOf('day');
  };

  const handleTimeChange = (dates: any) => {
    if (dates) {
      setTimeRange([dates[0], dates[1]]);
    }
  };

  const { data: sqlMetaData } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      return listSqlStats({
        // Reuse existing list API to get meta
        namespace: ns,
        obtenant: name,
        startTime: timeRange[0].unix(),
        endTime: timeRange[1].unix(),
        keyword: sqlId, // Filter by SQL ID
        outputColumns: ['query_sql', 'user_name', 'db_name', 'sql_id'], // metrics
        pageSize: 1,
        pageNum: 1,
      } as any);
    },
    {
      refreshDeps: [ns, name, sqlId, timeRange],
    },
  );

  const sqlMeta = stateSqlMeta || sqlMetaData?.data?.items?.[0];

  const [planDrawerOpen, setPlanDrawerOpen] = useState(false);

  const [currentPlanId, setCurrentPlanId] = useState<number>();

  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  const [planDataSource, setPlanDataSource] = useState<any[]>([]);

  const {
    loading: planDetailLoading,

    run: fetchPlanDetail,
  } = useRequest(
    async (record: API.PlanStatistic) => {
      if (!ns || !name) return;

      const res = await queryPlanDetailInfo({
        namespace: ns,

        obtenant: name,

        tenantID: record.tenantID,

        svrIP: record.svrIP,

        svrPort: record.svrPort,

        planID: record.planID,
      });

      return res;
    },

    {
      manual: true,

      onSuccess: (res) => {
        const data = res?.data || (res as any);

        if (data?.planDetail) {
          const keys: React.Key[] = [];

          // Deep clone to avoid mutating the cached data

          const root = JSON.parse(JSON.stringify(data.planDetail));

          const traverse = (node: any, key: string) => {
            node.key = key;

            keys.push(key);

            if (node.childOperators && node.childOperators.length > 0) {
              node.childOperators.forEach((child: any, index: number) => {
                traverse(child, `${key}-${index}`);
              });
            }
          };

          traverse(root, '0');

          setPlanDataSource([root]);

          setExpandedKeys(keys);
        } else {
          setPlanDataSource([]);

          setExpandedKeys([]);
        }
      },
    },
  );

  const handlePlanClick = (record: API.PlanStatistic) => {
    setCurrentPlanId(record.planID);

    setPlanDataSource([]);

    setExpandedKeys([]);

    setPlanDrawerOpen(true);

    fetchPlanDetail(record);
  };

  // Handle both wrapped (response.data) and unwrapped (response IS data) cases

  const sqlInfo = detailData?.data || (detailData as any);
  const historyInfo = historyData?.data || (historyData as any);

  return (
    <div
      style={{
        backgroundColor: '#F5F7FA',

        minHeight: '100vh',

        padding: 24,
      }}
    >
      {/* Header & Basic Info */}

      <ProCard ghost gutter={[16, 16]} direction="column">
        <ProCard loading={detailLoading}>
          <Space
            style={{
              marginBottom: 16,

              justifyContent: 'space-between',

              width: '100%',
            }}
          >
            <Space>
              <Button
                icon={<ArrowLeftOutlined />}
                onClick={() => history.back()}
                type="text"
              >
                {intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.Back',
                  defaultMessage: '返回',
                })}
              </Button>

              <Title level={4} style={{ margin: 0 }}>
                {intl.formatMessage(
                  {
                    id: 'src.pages.Tenant.Detail.Sql.Detail.SqlDetail',
                    defaultMessage: 'SQL 详情：{sqlId}',
                  },
                  { sqlId },
                )}
              </Title>
            </Space>

            {/* Time Picker Removed from here */}
          </Space>

          <ProDescriptions column={3} bordered>
            <ProDescriptions.Item
              label={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.SqlText',
                defaultMessage: 'SQL 文本',
              })}
              span={3}
            >
              <Typography.Paragraph
                ellipsis={{ rows: 2, expandable: true, symbol: 'more' }}
                copyable
              >
                {sqlMeta?.querySql || sqlInfo?.querySql || '-'}
              </Typography.Paragraph>
            </ProDescriptions.Item>

            <ProDescriptions.Item
              label={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.SqlId',
                defaultMessage: 'SQL ID',
              })}
            >
              {sqlId}
            </ProDescriptions.Item>

            <ProDescriptions.Item
              label={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.Database',
                defaultMessage: '数据库',
              })}
            >
              {dbName || sqlMeta?.dbName || '-'}
            </ProDescriptions.Item>

            <ProDescriptions.Item
              label={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.User',
                defaultMessage: '用户',
              })}
            >
              {sqlMeta?.userName || '-'}
            </ProDescriptions.Item>
          </ProDescriptions>
        </ProCard>

        {/* Charts */}

        <ProCard
          title={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Sql.Detail.HistoryRequestInfo',
            defaultMessage: '历史请求信息',
          })}
          headerBordered
          loading={historyLoading}
          extra={
            <RangePicker
              showTime
              value={timeRange}
              onChange={handleTimeChange}
              format={DATE_TIME_FORMAT}
              disabledDate={disabledDate}
              disabledTime={disabledDateTime}
              presets={DateSelectOption.filter((o) => o.value !== 'custom').map(
                (o) => ({
                  label: o.label,

                  value: [dayjs().subtract(o.value as number, 'ms'), dayjs()],
                }),
              )}
            />
          }
        >
          <ProCard split="vertical">
            <ProCard
              title={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.Executions',
                defaultMessage: '执行次数',
              })}
              colSpan={12}
            >
              <SqlTrendChart
                data={historyInfo?.executionTrend || []}
                type="execution"
              />
            </ProCard>

            <ProCard
              title={
                <Space>
                  <span>
                    {intl.formatMessage({
                      id: 'src.pages.Tenant.Detail.Sql.Detail.Latency',
                      defaultMessage: '延迟',
                    })}
                  </span>

                  <Select
                    mode="multiple"
                    maxTagCount="responsive"
                    value={selectedLatencyMetrics}
                    onChange={setSelectedLatencyMetrics}
                    style={{ width: 400 }}
                    options={latencyMetricsMeta.map((m) => ({
                      label: m.name,

                      value: m.key,
                    }))}
                  />
                </Space>
              }
              colSpan={12}
            >
              <SqlTrendChart
                data={historyInfo?.latencyTrend || []}
                type="latency"
              />
            </ProCard>
          </ProCard>
        </ProCard>

        {/* Diagnosis */}

        <ProCard
          title={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Sql.Detail.DiagnosisAndAdvice',
            defaultMessage: '诊断与建议',
          })}
          headerBordered
          loading={detailLoading}
        >
          {sqlInfo?.diagnoseInfo && sqlInfo.diagnoseInfo.length > 0 ? (
            <ProTable<any>
              rowKey={(record, index) => `${record.ruleName}-${index}`}
              dataSource={sqlInfo.diagnoseInfo}
              search={false}
              options={false}
              toolBarRender={false}
              pagination={false}
              columns={[
                {
                  title: intl.formatMessage({
                    id: 'src.pages.Tenant.Detail.Sql.Detail.Level',
                    defaultMessage: '级别',
                  }),
                  dataIndex: 'level',
                  width: 100,
                  render: (_: any, record: any) => {
                    const level = record.level;
                    let color = 'blue';
                    if (level === 'CRITICAL') color = 'red';
                    if (level === 'WARN') color = 'orange';
                    if (level === 'NOTICE') color = 'cyan';
                    return <Tag color={color}>{level}</Tag>;
                  },
                },
                {
                  title: intl.formatMessage({
                    id: 'src.pages.Tenant.Detail.Sql.Detail.RuleName',
                    defaultMessage: '规则名称',
                  }),
                  dataIndex: 'ruleName',
                  width: 200,
                },
                {
                  title: intl.formatMessage({
                    id: 'src.pages.Tenant.Detail.Sql.Detail.Reason',
                    defaultMessage: '原因',
                  }),
                  dataIndex: 'reason',
                },
                {
                  title: intl.formatMessage({
                    id: 'src.pages.Tenant.Detail.Sql.Detail.Suggestion',
                    defaultMessage: '建议',
                  }),
                  dataIndex: 'suggestion',
                },
              ]}
            />
          ) : (
            <Tag color="green">
              {intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.NoDiagnosisIssuesFound',
                defaultMessage: '未发现诊断问题',
              })}
            </Tag>
          )}
        </ProCard>

        {/* Plan Statistics */}

        <ProCard
          title={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Sql.Detail.PlanStatistics',
            defaultMessage: '执行计划统计',
          })}
          headerBordered
          loading={detailLoading}
        >
          <ProTable<API.PlanStatistic>
            rowKey={(record) =>
              `${record.svrIP}-${record.svrPort}-${record.planID}`
            }
            dataSource={sqlInfo?.plans || []}
            columns={[
              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.PlanId',
                  defaultMessage: '执行计划 ID',
                }),

                dataIndex: 'planID',

                render: (text, record) => (
                  <a onClick={() => handlePlanClick(record)}>{text}</a>
                ),
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.SvrIp',
                  defaultMessage: '服务器 IP',
                }),
                dataIndex: 'svrIP',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.SvrPort',
                  defaultMessage: '服务器端口',
                }),
                dataIndex: 'svrPort',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.PlanHash',
                  defaultMessage: '执行计划哈希',
                }),
                dataIndex: 'planHash',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.Cost',
                  defaultMessage: '成本',
                }),

                dataIndex: 'cost',

                sorter: (a, b) => a.cost - b.cost,
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.CpuCost',
                  defaultMessage: 'CPU 成本',
                }),

                dataIndex: 'cpuCost',

                sorter: (a, b) => a.cpuCost - b.cpuCost,
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.IoCost',
                  defaultMessage: 'IO 成本',
                }),

                dataIndex: 'ioCost',

                sorter: (a, b) => a.ioCost - b.ioCost,
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.GeneratedTime',
                  defaultMessage: '生成时间',
                }),

                dataIndex: 'generatedTime',

                render: (_, record) =>
                  dayjs.unix(record.generatedTime).format(DATE_TIME_FORMAT),

                sorter: (a, b) => a.generatedTime - b.generatedTime,
              },
            ]}
            search={false}
            toolBarRender={false}
            pagination={{ pageSize: 5 }}
          />
        </ProCard>

        {/* Index Info */}

        <ProCard
          title={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Sql.Detail.IndexInfo',
            defaultMessage: '索引信息',
          })}
          headerBordered
          loading={detailLoading}
        >
          <ProTable<API.IndexInfo>
            dataSource={sqlInfo?.indexies || []}
            columns={[
              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.TableName',
                  defaultMessage: '表名',
                }),
                dataIndex: 'tableName',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.IndexName',
                  defaultMessage: '索引名',
                }),
                dataIndex: 'indexName',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.IndexType',
                  defaultMessage: '索引类型',
                }),
                dataIndex: 'indexType',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.Uniqueness',
                  defaultMessage: '唯一性',
                }),
                dataIndex: 'uniqueness',
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.Columns',
                  defaultMessage: '列',
                }),

                dataIndex: 'columns',

                render: (cols) =>
                  Array.isArray(cols) ? cols.join(', ') : cols,
              },

              {
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.Detail.Status',
                  defaultMessage: '状态',
                }),
                dataIndex: 'status',
              },
            ]}
            search={false}
            toolBarRender={false}
            pagination={{ pageSize: 5 }}
          />
        </ProCard>
      </ProCard>

      <Drawer
        title={intl.formatMessage(
          {
            id: 'src.pages.Tenant.Detail.Sql.Detail.PlanDetail',
            defaultMessage: '执行计划详情（执行计划 ID：{planId}）',
          },
          { planId: currentPlanId },
        )}
        width={1000}
        open={planDrawerOpen}
        onClose={() => setPlanDrawerOpen(false)}
        bodyStyle={{ padding: 0 }}
      >
        <ProTable
          columns={[
            {
              title: intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.Operator',
                defaultMessage: '操作符',
              }),
              dataIndex: 'operator',
              key: 'operator',
            },

            {
              title: intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.Name',
                defaultMessage: '名称',
              }),
              dataIndex: 'name',
              key: 'name',
            },

            {
              title: intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.EstRows',
                defaultMessage: '预估行数',
              }),

              dataIndex: 'estimatedRows',

              key: 'estimatedRows',
            },

            {
              title: intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.Cost',
                defaultMessage: '成本',
              }),
              dataIndex: 'cost',
              key: 'cost',
            },

            {
              title: intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.OutputFilter',
                defaultMessage: '输出/过滤',
              }),

              dataIndex: 'outputOrFilter',

              key: 'outputOrFilter',

              ellipsis: true,
            },
          ]}
          dataSource={planDataSource}
          rowKey="key"
          loading={planDetailLoading}
          pagination={false}
          search={false}
          options={false}
          expandable={{
            expandedRowKeys: expandedKeys,

            onExpandedRowsChange: (keys) =>
              setExpandedKeys(keys as React.Key[]),

            childrenColumnName: 'childOperators',
          }}
        />
      </Drawer>
    </div>
  );
};

export default SqlDetail;
