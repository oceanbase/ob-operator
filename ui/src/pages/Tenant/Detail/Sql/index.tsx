import { obtenant } from '@/api';
import EmptyImg from '@/assets/empty.svg';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { DATE_TIME_FORMAT, DateSelectOption } from '@/constants/datetime';
import { listSqlMetrics, listSqlStats } from '@/services/sql';
import { getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { SettingOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Spin } from '@oceanbase/design';
import {
  Link,
  history,
  useLocation,
  useParams,
  useRequest,
  useSearchParams,
} from '@umijs/max';
import { Button, Card, Checkbox, Tooltip, message } from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import dayjs from 'dayjs';
import { useEffect, useMemo, useRef, useState } from 'react';
import { getLocale } from 'umi';
import ColumnSelectionDrawer from './ColumnSelectionDrawer';

const SQL_LIST_STORAGE_KEY = 'sql_list_params';

export default function SqlList() {
  const { ns, name, tenantName } = useParams<{
    ns: string;
    name: string;
    tenantName: string;
  }>();
  const location = useLocation();
  const actionRef = useRef<ActionType>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedMetricKeys, setSelectedMetricKeys] = useState<string[]>([]);
  const [maxElapsedTime, setMaxElapsedTime] = useState<number>(0);
  const [searchParams, setSearchParams] = useSearchParams();
  const [activeTab, setActiveTab] = useState<string>(
    searchParams.get('activeTab') || 'sql_analysis',
  );
  const [currentParams, setCurrentParams] = useState<{
    startTime?: number;
    endTime?: number;
  }>({});

  // 检测是否从 SQL 详情页返回，如果是则恢复保存的参数
  useEffect(() => {
    const savedParams = sessionStorage.getItem(SQL_LIST_STORAGE_KEY);
    if (savedParams) {
      try {
        const params = JSON.parse(savedParams);
        // 检查是否是从详情页返回（通过 location.state 或 URL 参数判断）
        const isFromDetail = (location.state as any)?.fromDetail || false;

        if (isFromDetail && params) {
          // 恢复参数到 URL
          const newSearchParams = new URLSearchParams();
          if (params.activeTab) {
            newSearchParams.set('activeTab', params.activeTab);
            setActiveTab(params.activeTab);
          }
          if (params.startTime) {
            newSearchParams.set('startTime', params.startTime.toString());
          }
          if (params.endTime) {
            newSearchParams.set('endTime', params.endTime.toString());
          }
          if (params.keyword) {
            newSearchParams.set('keyword', params.keyword);
          }
          if (params.user) {
            newSearchParams.set('user', params.user);
          }
          if (params.database) {
            newSearchParams.set('database', params.database);
          }
          if (params.includeInnerSql !== undefined) {
            newSearchParams.set(
              'includeInnerSql',
              params.includeInnerSql.toString(),
            );
          }

          setSearchParams(newSearchParams);
          // 清除保存的参数
          sessionStorage.removeItem(SQL_LIST_STORAGE_KEY);

          // 延迟刷新表格，确保参数已更新
          setTimeout(() => {
            actionRef.current?.reload();
          }, 100);
        }
      } catch (e) {
        console.error('Failed to restore SQL list params:', e);
        sessionStorage.removeItem(SQL_LIST_STORAGE_KEY);
      }
    }
  }, [location.state, setSearchParams]);

  // 监听路由变化，当离开 SQL 页面时清除 SQL 相关参数
  useEffect(() => {
    const sqlRelatedParams = [
      'startTime',
      'endTime',
      'includeInnerSql',
      'keyword',
      'user',
      'database',
      'activeTab',
      'current',
      'pageSize',
    ];

    const unlisten = history.listen(({ location: newLocation }) => {
      const isSqlPage = newLocation.pathname.includes('/sql');

      if (!isSqlPage) {
        // 如果不在 SQL 页面，只清除 SQL 相关的 URL 参数，保留路径和其他参数
        const currentParams = new URLSearchParams(newLocation.search);
        let hasChanges = false;

        // 只删除 SQL 相关的参数，保留其他所有参数
        sqlRelatedParams.forEach((param) => {
          if (currentParams.has(param)) {
            currentParams.delete(param);
            hasChanges = true;
          }
        });

        if (hasChanges) {
          // 保留原始路径，只更新查询参数
          const newSearch = currentParams.toString();
          const newUrl = `${newLocation.pathname}${
            newSearch ? `?${newSearch}` : ''
          }${newLocation.hash || ''}`;

          // 使用 replace 而不是 push，避免在历史记录中留下带参数的 URL
          // 延迟执行，确保路由已经完成跳转
          setTimeout(() => {
            history.replace(newUrl);
          }, 0);
        }
      }
    });

    // 组件卸载时也清除参数（如果还在当前页面）
    return () => {
      unlisten();
      // 检查当前路径是否还是 SQL 页面
      const currentPath = window.location.pathname;
      if (!currentPath.includes('/sql')) {
        const currentParams = new URLSearchParams(window.location.search);
        let hasChanges = false;

        // 只删除 SQL 相关的参数，保留其他所有参数
        sqlRelatedParams.forEach((param) => {
          if (currentParams.has(param)) {
            currentParams.delete(param);
            hasChanges = true;
          }
        });

        if (hasChanges) {
          const newSearch = currentParams.toString();
          const newUrl = `${currentPath}${newSearch ? `?${newSearch}` : ''}${
            window.location.hash || ''
          }`;
          window.history.replaceState({ ...window.history.state }, '', newUrl);
        }
      }
    };
  }, []);

  // Helper to robustly extract metrics array regardless of response format
  const getMetricsList = (data: any): API.SqlMetricMetaCategory[] => {
    if (!data) return [];
    if (Array.isArray(data)) return data;
    if (data.data && Array.isArray(data.data)) return data.data;
    return [];
  };

  // Fetch metric metadata to know available columns and defaults
  const { data: metricsData } = useRequest(
    () =>
      listSqlMetrics({ language: getLocale() === 'zh-CN' ? 'zh_CN' : 'en_US' }),
    {
      onSuccess: (data) => {
        const list = getMetricsList(data);
        const defaults: string[] = [];
        list.forEach((category) => {
          category.metrics.forEach((metric) => {
            if (metric.displayByDefault || metric.immutable) {
              defaults.push(metric.key);
            }
          });
        });
        setSelectedMetricKeys(defaults);
      },
    },
  );

  const {
    data: tenantDetailResponse,
    run: getTenantDetail,
    loading: tenantDetailLoading,
  } = useRequest(getTenant, {
    defaultParams: [{ ns: ns!, name: name! }],
  });

  const defaultSqlAnalyzer = tenantDetailResponse?.info?.sqlAnalyzerEnabled;

  const { run: createSQLAnalyzer } = useRequest(obtenant.createSQLAnalyzer, {
    manual: true,
    onSuccess: () => {
      message.success(
        intl.formatMessage({
          id: 'src.pages.Tenant.Detail.Sql.SqlDiagnosisEnabled',
          defaultMessage: 'SQL诊断已开启',
        }),
      );
    },
  });

  const initialTimeRange: [dayjs.Dayjs, dayjs.Dayjs] = useMemo(
    () => [dayjs().subtract(30, 'minute'), dayjs()],
    [],
  );

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

  const metaFieldMap: Record<string, keyof API.SqlMetaInfo> = {
    query_sql: 'querySql',
    db_name: 'dbName',
    user_name: 'userName',
    sql_id: 'sqlId',
    svr_ip: 'svrIp',
    svr_port: 'svrPort',
    client_ip: 'clientIp',
  };

  const METRIC_COLORS: Record<string, string> = {
    execute_time: '#4096FF', // blue
    queue_time: '#95DE54', // green
    get_plan_time: '#FFD666', // orange
  };

  // ... existing imports ...

  // ... inside the component ...
  // Generate dynamic columns based on selected keys and metadata
  const dynamicColumns: ProColumns<API.SqlInfo>[] = useMemo(() => {
    const list = getMetricsList(metricsData);
    if (list.length === 0 || selectedMetricKeys.length === 0) return [];

    const cols: ProColumns<API.SqlInfo>[] = [];
    const allMetrics: API.SqlMetricMeta[] = [];
    list.forEach((cat) => {
      allMetrics.push(...cat.metrics);
    });

    allMetrics.forEach((metric) => {
      if (selectedMetricKeys.includes(metric.key)) {
        const title = metric.unit
          ? `${metric.name} (${metric.unit})`
          : metric.name;
        const colConfig: ProColumns<API.SqlInfo> = {
          title: title,
          dataIndex: metric.key,
          search: false,
          width: 120,
        };

        if (metaFieldMap[metric.key]) {
          colConfig.dataIndex = metaFieldMap[metric.key];
          if (metric.key === 'sql_id') {
            colConfig.width = 120;
            colConfig.copyable = true;
            colConfig.ellipsis = true;
          } else if (metric.key === 'query_sql') {
            colConfig.fixed = 'left';
            colConfig.width = 150;
            colConfig.ellipsis = true;
            colConfig.copyable = true;
            colConfig.render = (dom, record) => {
              const handleClick = () => {
                // 保存当前表单参数到 sessionStorage
                // 从 URL 参数和当前状态获取表单值
                const paramsToSave = {
                  activeTab,
                  startTime: currentParams.startTime,
                  endTime: currentParams.endTime,
                  keyword: searchParams.get('keyword') || undefined,
                  user: searchParams.get('user') || undefined,
                  database: searchParams.get('database') || undefined,
                  includeInnerSql:
                    searchParams.get('includeInnerSql') === 'true',
                };
                sessionStorage.setItem(
                  SQL_LIST_STORAGE_KEY,
                  JSON.stringify(paramsToSave),
                );
              };

              const params = new URLSearchParams();
              params.append('dbName', record.dbName);
              if (currentParams.startTime) {
                params.append('startTime', currentParams.startTime.toString());
              }
              if (currentParams.endTime) {
                params.append('endTime', currentParams.endTime.toString());
              }
              return (
                <Link
                  to={{
                    pathname: `/tenant/${ns}/${name}/${tenantName}/sql/${record.sqlId}`,
                    search: params.toString(),
                  }}
                  state={record}
                  onClick={handleClick}
                >
                  {dom}
                </Link>
              );
            };
          } else if (metric.key === 'user_name') {
            colConfig.width = 100;
          }
        } else if (metric.key === 'elapsed_time') {
          colConfig.width = 250;
          colConfig.defaultSortOrder = 'descend';
          colConfig.render = (_, record) => {
            const elapsedStat = record.latencyStatistics?.find(
              (s) => s.name === 'elapsed_time',
            );
            if (!elapsedStat) return '-';
            const total = elapsedStat.value;
            if (total <= 0) return '0.00';

            // Find component metrics that are currently selected and have a color defined
            const components =
              record.latencyStatistics?.filter(
                (s) =>
                  s.name !== 'elapsed_time' &&
                  selectedMetricKeys.includes(s.name) &&
                  METRIC_COLORS[s.name],
              ) || [];

            // Sort components based on the order of keys in METRIC_COLORS to ensure consistent display order
            const orderedKeys = Object.keys(METRIC_COLORS);
            components.sort((a, b) => {
              return orderedKeys.indexOf(a.name) - orderedKeys.indexOf(b.name);
            });

            // Calculate width for each component relative to the total of this row
            const segments = components.map((comp) => {
              const width = (comp.value / total) * 100;
              return {
                name: comp.name,
                value: comp.value,
                width,
                color: METRIC_COLORS[comp.name],
              };
            });

            // Calculate the width of the bar relative to the max elapsed time on the page
            const MAX_BAR_WIDTH = 150;
            const barWidth =
              maxElapsedTime > 0 ? (total / maxElapsedTime) * MAX_BAR_WIDTH : 0;

            return (
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <Tooltip
                  title={
                    <div>
                      <div>
                        {intl.formatMessage(
                          {
                            id: 'src.pages.Tenant.Detail.Sql.Total',
                            defaultMessage: '总计：{total} ms',
                          },
                          { total: total.toFixed(2) },
                        )}
                      </div>
                      {segments.map((seg) => (
                        <div
                          key={seg.name}
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 8,
                          }}
                        >
                          <span
                            style={{
                              width: 8,
                              height: 8,
                              backgroundColor: seg.color,
                              borderRadius: '50%',
                            }}
                          ></span>
                          <span>
                            {intl.formatMessage(
                              {
                                id: 'src.pages.Tenant.Detail.Sql.MetricValue',
                                defaultMessage: '{name}：{value} ms',
                              },
                              { name: seg.name, value: seg.value.toFixed(2) },
                            )}
                          </span>
                        </div>
                      ))}
                    </div>
                  }
                >
                  <div
                    style={{
                      width: barWidth,
                      height: 12,
                      backgroundColor: '#f0f0f0', // Background represents total
                      borderRadius: 2,
                      overflow: 'hidden',
                      display: 'flex',
                      marginRight: 8,
                      position: 'relative',
                    }}
                  >
                    {segments.map((seg, idx) => (
                      <div
                        key={idx}
                        style={{
                          width: `${seg.width}%`,
                          height: '100%',
                          backgroundColor: seg.color,
                        }}
                      />
                    ))}
                  </div>
                </Tooltip>
                <span>{total.toFixed(2)}</span>
              </div>
            );
          };
          colConfig.sorter = true;
        } else {
          colConfig.render = (_, record) => {
            const stat =
              record.executionStatistics?.find((s) => s.name === metric.key) ||
              record.latencyStatistics?.find((s) => s.name === metric.key);
            if (!stat) return '-';
            return Number.isInteger(stat.value)
              ? stat.value
              : stat.value.toFixed(2);
          };
          colConfig.sorter = true;
        }
        cols.push(colConfig);
      }
    });
    return cols;
  }, [
    metricsData,
    selectedMetricKeys,
    ns,
    name,
    tenantName,
    maxElapsedTime,
    currentParams,
  ]);

  const columns: ProColumns<API.SqlInfo>[] = [
    ...dynamicColumns,
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.UserName',
        defaultMessage: '用户名',
      }),
      dataIndex: 'user',
      hideInTable: true,
      order: 100,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.DatabaseName',
        defaultMessage: '数据库',
      }),
      dataIndex: 'database',
      hideInTable: true,
      order: 99,
    },
    {
      title: '',
      dataIndex: 'includeInnerSql',
      hideInTable: true,
      order: 98,
      formItemProps: {
        valuePropName: 'checked',
      },
      renderFormItem: (_item, _config, form) => {
        return (
          <Checkbox
            onChange={(e) => {
              form.setFieldValue('includeInnerSql', e.target.checked);
              form.submit();
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.IncludeInnerSqls',
              defaultMessage: '包含内部 SQL',
            })}
          </Checkbox>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.TimeRange',
        defaultMessage: '时间范围',
      }),
      dataIndex: 'timeRange',
      valueType: 'dateTimeRange',
      hideInTable: true,
      order: 97,
      fieldProps: {
        format: DATE_TIME_FORMAT,
        disabledDate: disabledDate,
        disabledTime: disabledDateTime,
        presets: DateSelectOption.filter((o) => o.value !== 'custom').map(
          (o) => ({
            label: o.label,
            value: [dayjs().subtract(o.value as number, 'ms'), dayjs()],
          }),
        ),
      },
      search: {
        transform: (value: [string, string]) => {
          return {
            startTime: dayjs(value[0]).unix(),
            endTime: dayjs(value[1]).unix(),
          };
        },
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.Keyword',
        defaultMessage: '关键字',
      }),
      dataIndex: 'keyword',
      hideInTable: true,
      order: 96,
    },
  ];

  const handSqlAnalyzer = async () => {
    createSQLAnalyzer(ns, name).then(() => {
      if (ns && name) {
        getTenantDetail({ ns, name });
      }
    });
  };

  return (
    <div
      style={{
        backgroundColor: 'transparent',
        minHeight: '100vh',
        padding: 24,
      }}
    >
      {tenantDetailLoading ? (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: 'calc(100vh - 98px)',
          }}
        >
          <Spin />
        </div>
      ) : defaultSqlAnalyzer ? (
        <>
          <ProTable<API.SqlInfo>
            headerTitle={intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.SqlAnalysis',
              defaultMessage: 'SQL 分析',
            })}
            loading={tenantDetailLoading}
            actionRef={actionRef}
            rowKey={(record) =>
              `${record.sqlId}_${record.svrIp}_${record.svrPort}_${record.planId}_${record.userName}_${record.dbName}`
            }
            params={{ outputColumns: selectedMetricKeys, activeTab }}
            toolbar={{
              menu: {
                type: 'tab',
                activeKey: activeTab,
                items: [
                  {
                    label: intl.formatMessage({
                      id: 'src.pages.Tenant.Detail.Sql.SqlAnalysis',
                      defaultMessage: 'SQL 分析',
                    }),
                    key: 'sql_analysis',
                  },
                  {
                    label: intl.formatMessage({
                      id: 'src.pages.Tenant.Detail.Sql.SlowSql',
                      defaultMessage: '慢 SQL',
                    }),
                    key: 'slow_sql',
                  },
                ],
                onChange: (key) => {
                  const k = key as string;
                  setActiveTab(k);
                  const newParams = new URLSearchParams(searchParams);
                  newParams.set('activeTab', k);
                  setSearchParams(newParams);
                },
              },
            }}
            form={{
              syncToUrl: (values, type) => {
                if (type === 'get') {
                  const { startTime, endTime, ...rest } = values;
                  return {
                    ...rest,
                    timeRange:
                      startTime && endTime
                        ? [
                            dayjs.unix(Number(startTime)),
                            dayjs.unix(Number(endTime)),
                          ]
                        : initialTimeRange,
                  };
                }
                const { timeRange, ...rest } = values;
                if (timeRange && timeRange[0] && timeRange[1]) {
                  return {
                    ...rest,
                    startTime: dayjs(timeRange[0]).unix(),
                    endTime: dayjs(timeRange[1]).unix(),
                  };
                }
                return rest;
              },
            }}
            search={{
              collapsed: false,
              collapseRender: false,
              labelWidth: 'auto',
              span: 8,
            }}
            options={false}
            toolBarRender={() => [
              <Button
                key="column-selection"
                icon={<SettingOutlined />}
                onClick={() => setDrawerOpen(true)}
              >
                {intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.ColumnSelection',
                  defaultMessage: '列选择',
                })}
              </Button>,
            ]}
            scroll={{ x: 1500 }}
            request={async (params, sort) => {
              if (!ns || !name || !tenantName) {
                return { data: [], success: false };
              }

              const { startTime, endTime, ...restParams } = params;

              // Ensure startTime and endTime are present, defaulting to initialTimeRange if not
              const effectiveStartTime =
                startTime ?? initialTimeRange[0].unix();
              const effectiveEndTime = endTime ?? initialTimeRange[1].unix();

              setCurrentParams({
                startTime: effectiveStartTime,
                endTime: effectiveEndTime,
              });

              // 如果没有排序参数，默认使用 elapsed_time 降序
              const sortKeys = Object.keys(sort);
              const defaultSortColumn =
                sortKeys.length > 0 ? Object.keys(sort)[0] : 'elapsed_time';
              const defaultSortOrder =
                sortKeys.length > 0
                  ? Object.values(sort)[0] === 'ascend'
                    ? 'asc'
                    : 'desc'
                  : 'desc';

              const msg = await listSqlStats({
                namespace: ns,
                obtenant: name,
                sortColumn: defaultSortColumn,
                sortOrder: defaultSortOrder,
                pageNum: restParams.current,
                pageSize: restParams.pageSize,
                keyword: restParams.keyword as string,
                user: restParams.user as string,
                database: restParams.database as string,
                includeInnerSql: restParams.includeInnerSql as boolean,
                suspiciousOnly: activeTab === 'slow_sql',
                startTime: effectiveStartTime,
                endTime: effectiveEndTime,
                outputColumns: selectedMetricKeys,
              });

              const items = msg.data?.items || [];
              let maxTime = 0;
              items.forEach((item) => {
                const elapsed =
                  item.latencyStatistics?.find((s) => s.name === 'elapsed_time')
                    ?.value || 0;
                if (elapsed > maxTime) maxTime = elapsed;
              });
              setMaxElapsedTime(maxTime);

              return {
                data: items,
                success: msg.successful,
                total: msg.data?.totalCount || 0,
                page: params.current,
              };
            }}
            columns={columns}
            pagination={{
              defaultPageSize: 20,
              showSizeChanger: true,
            }}
          />
          <ColumnSelectionDrawer
            open={drawerOpen}
            onClose={() => setDrawerOpen(false)}
            selectedKeys={selectedMetricKeys}
            onSelectionChange={(keys) => {
              setSelectedMetricKeys(keys);
              actionRef.current?.reload();
            }}
            metrics={getMetricsList(metricsData)}
          />
        </>
      ) : (
        <Card
          style={{
            height: 'calc(100vh - 98px)',
          }}
          bodyStyle={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100%',
            padding: '40px 24px',
          }}
        >
          <img
            src={EmptyImg}
            alt="empty"
            style={{ marginBottom: 24, height: 100, width: 110 }}
          />
          <p
            style={{
              color: '#8592ad',
              marginBottom: 24,
              textAlign: 'center',
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.TenantHasNotEnabledSqlDiagnosis',
              defaultMessage: '该租户尚未开启 SQL 分析，是否立即开启？',
            })}
          </p>
          <Button
            type="primary"
            onClick={() => {
              showDeleteConfirm({
                onOk: handSqlAnalyzer,
                title: intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.Sql.ConfirmEnableSqlDiagnosis',
                  defaultMessage: '确认要开启 SQL 分析吗？',
                }),
              });
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Sql.EnableNow',
              defaultMessage: '立即开启',
            })}
          </Button>
        </Card>
      )}
    </div>
  );
}
