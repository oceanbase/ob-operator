import { DATE_TIME_FORMAT, DateSelectOption } from '@/constants/datetime';
import {
  listSqlMetrics,
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
  Checkbox,
  DatePicker,
  Drawer,
  Select,
  Space,
  Tag,
  Typography,
} from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import dayjs from 'dayjs';
import { useEffect, useMemo, useRef, useState } from 'react';
import { getLocale } from 'umi';

const { RangePicker } = DatePicker;
const { Title } = Typography;

// --- Chart Component ---

interface SqlTrendChartProps {
  data: API.MetricData[];
  type: 'execution' | 'latency';
  height?: number;
  loading?: boolean;
}

const SqlTrendChart: React.FC<SqlTrendChartProps> = ({
  data,
  type,
  loading,
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

  return <div ref={containerRef} style={{ height }} loading={loading} />;
};

// --- Main Page Component ---

const SqlDetail: React.FC = () => {
  const { ns, name, tenantName, sqlId } = useParams<{
    ns: string;
    name: string;
    tenantName: string;
    sqlId: string;
  }>();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const stateSqlMeta = location.state as API.SqlInfo | undefined;

  // 优先使用从列表页传递的数据，避免重复加载
  // 从列表页传递的数据包含：querySql, dbName, userName, sqlId 等完整信息
  // 使用 useMemo 确保这些值在 stateSqlMeta 变化时更新
  const dbName = useMemo(
    () => stateSqlMeta?.dbName || searchParams.get('dbName') || '',
    [stateSqlMeta?.dbName, searchParams],
  );
  const userName = useMemo(
    () => stateSqlMeta?.userName || searchParams.get('userName') || '',
    [stateSqlMeta?.userName, searchParams],
  );
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
  // 临时存储选中的指标，等下拉框关闭后再更新到正式状态
  const [tempSelectedLatencyMetrics, setTempSelectedLatencyMetrics] = useState<
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
            setTempSelectedLatencyMetrics(defaults);
          } else {
            // Fallback to first if no defaults
            if (latencyCat.metrics.length > 0) {
              setSelectedLatencyMetrics([latencyCat.metrics[0].key]);
              setTempSelectedLatencyMetrics([latencyCat.metrics[0].key]);
            }
          }
        }
      },
    },
  );

  // 1. Fetch Static Detail Info (Plans, Indexes, etc.)
  // Use a wider range or the initial range to find the SQL text
  // 注意：SQL 文本、数据库、用户等基本信息优先使用从列表页传递的 stateSqlMeta，避免重复加载
  // 如果从列表页传递了完整数据，API 主要用于获取 Plans 和 Indexes 等详细信息
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

      // 优先使用从列表页传递的数据，确保 API 调用时使用正确的参数
      // 这样后端可以更精确地查询，避免重复加载基本信息
      const effectiveDbName = stateSqlMeta?.dbName || dbName;
      const effectiveUserName = stateSqlMeta?.userName || userName;

      return querySqlDetailInfo({
        namespace: ns,
        obtenant: name,
        sqlId,
        database: effectiveDbName,
        user: effectiveUserName,
        startTime: start,
        endTime: end,
      } as any);
    },
    {
      // 添加 stateSqlMeta 到依赖项，确保当传递的数据变化时重新调用
      refreshDeps: [
        ns,
        name,
        sqlId,
        dbName,
        stateSqlMeta?.dbName,
        stateSqlMeta?.userName,
      ],
    },
  );

  // 2. Fetch History Trend Info
  const {
    data: historyData,
    loading: historyLoading,
    run: refreshHistoryData,
  } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      const start = timeRange[0].unix();
      const end = timeRange[1].unix();
      // Calculate interval
      let interval = Math.floor((end - start) / 60);
      if (interval < 1) interval = 1;

      // 优先使用从列表页传递的数据，确保 API 调用时使用正确的参数
      const effectiveDbName = stateSqlMeta?.dbName || dbName;
      const effectiveUserName = stateSqlMeta?.userName || userName;

      return querySqlHistoryInfo({
        namespace: ns,
        obtenant: name,
        sqlId,
        database: effectiveDbName,
        user: effectiveUserName,
        startTime: start,
        endTime: end,
        interval,
        outputColumns: selectedLatencyMetrics,
      });
    },
    {
      // 添加 stateSqlMeta 到依赖项，确保当传递的数据变化时重新调用
      refreshDeps: [
        timeRange,
        selectedLatencyMetrics,
        ns,
        name,
        sqlId,
        dbName,
        stateSqlMeta?.dbName,
        stateSqlMeta?.userName,
      ],
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

  const [planDrawerOpen, setPlanDrawerOpen] = useState(false);

  const [currentPlanId, setCurrentPlanId] = useState<number>();

  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  const [planDataSource, setPlanDataSource] = useState<any[]>([]);

  const { loading: planDetailLoading, run: fetchPlanDetail } = useRequest(
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
        {/* 基本信息卡片：如果有传递的数据，不显示加载状态 */}
        <ProCard loading={!stateSqlMeta && detailLoading}>
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
                onClick={() => {
                  // 返回到列表页，并带上标记表示是从详情页返回的
                  history.push({
                    pathname: `/tenant/${ns}/${name}/${tenantName}/sql`,
                    state: { fromDetail: true },
                  });
                }}
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
                {/* 只使用从列表页传递的数据，不等待 API 返回 */}
                {stateSqlMeta?.querySql || '-'}
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
              {/* 只使用从列表页传递的数据，不等待 API 返回 */}
              {stateSqlMeta?.dbName || '-'}
            </ProDescriptions.Item>

            <ProDescriptions.Item
              label={intl.formatMessage({
                id: 'src.pages.Tenant.Detail.Sql.Detail.User',
                defaultMessage: '用户',
              })}
            >
              {/* 只使用从列表页传递的数据，不等待 API 返回 */}
              {stateSqlMeta?.userName || '-'}
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
                loading={historyLoading}
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
                    value={tempSelectedLatencyMetrics}
                    onChange={setTempSelectedLatencyMetrics}
                    onDropdownVisibleChange={(open) => {
                      // 当下拉框关闭时，更新正式状态并刷新数据
                      if (!open) {
                        setSelectedLatencyMetrics(tempSelectedLatencyMetrics);
                        // 延迟刷新，确保状态已更新
                        setTimeout(() => {
                          refreshHistoryData();
                        }, 0);
                      }
                    }}
                    style={{ width: 400 }}
                    options={latencyMetricsMeta.map((m) => ({
                      label: m.name,
                      value: m.key,
                    }))}
                    dropdownRender={(menu) => (
                      <div>
                        <div
                          style={{
                            padding: '4px 8px',
                            borderBottom: '1px solid #f0f0f0',
                          }}
                        >
                          <Checkbox
                            indeterminate={
                              tempSelectedLatencyMetrics.length > 0 &&
                              tempSelectedLatencyMetrics.length <
                                latencyMetricsMeta.length
                            }
                            checked={
                              latencyMetricsMeta.length > 0 &&
                              tempSelectedLatencyMetrics.length ===
                                latencyMetricsMeta.length
                            }
                            onChange={(e) => {
                              if (e.target.checked) {
                                setTempSelectedLatencyMetrics(
                                  latencyMetricsMeta.map((m) => m.key),
                                );
                              } else {
                                setTempSelectedLatencyMetrics([]);
                              }
                            }}
                          >
                            {intl.formatMessage({
                              id: 'src.pages.Tenant.Detail.Sql.Detail.SelectAll',
                              defaultMessage: '全选',
                            })}
                          </Checkbox>
                        </div>
                        {menu}
                      </div>
                    )}
                  />
                </Space>
              }
              colSpan={12}
            >
              <SqlTrendChart
                data={historyInfo?.latencyTrend || []}
                type="latency"
                loading={historyLoading}
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
        >
          {sqlInfo?.diagnoseInfo && sqlInfo.diagnoseInfo.length > 0 ? (
            <ProTable<any>
              loading={detailLoading}
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
        >
          <ProTable<API.PlanStatistic>
            loading={detailLoading}
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
        >
          <ProTable<API.IndexInfo>
            loading={detailLoading}
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
