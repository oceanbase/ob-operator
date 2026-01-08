import { inspection } from '@/api';
import type {
  InspectionInspectionScenario,
  InspectionPolicy,
} from '@/api/generated/api';
import CustomTooltip from '@/components/CustomTooltip';
import { TIME_FORMAT_WITHOUT_SECOND } from '@/constants/datetime';
import { getColumnSearchProps } from '@/utils/component';
import { parseCronExpression } from '@/utils/cron';
import { formatTime } from '@/utils/datetime';
import { intl } from '@/utils/intl';
import { getTimezoneInfo } from '@/utils/timezone';
import { CheckCircleFilled } from '@ant-design/icons';
import { theme } from '@oceanbase/design';
import { Link } from '@umijs/max';
import { useRequest } from 'ahooks';

import { history } from '@umijs/max';
import {
  Button,
  Drawer,
  Form,
  InputRef,
  Modal,
  Space,
  Table,
  Tabs,
  Tag,
  TimePicker,
  message,
} from 'antd';
import dayjs from 'dayjs';
import { toNumber } from 'lodash';

import { useEffect, useRef, useState } from 'react';
import SchduleSelectFormItem from '../Tenant/Detail/NewBackup/SchduleSelectFormItem';

export default function InspectionList() {
  const { token } = theme.useToken();
  const [searchText, setSearchText] = useState('');
  const [searchedColumn, setSearchedColumn] = useState('');
  const searchInput = useRef<InputRef>(null);
  const [open, setOpen] = useState(false);
  const [inspectionPolicies, setInspectionPolicies] = useState<any>({});
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('basic');
  // 存储所有tab的表单数据
  const [allTabData, setAllTabData] = useState<Record<string, any>>({});

  // 当drawer打开且inspectionPolicies有数据时，设置初始值
  useEffect(() => {
    if (open && inspectionPolicies?.obCluster) {
      // 延迟设置初始值，确保表单已经渲染
      setTimeout(() => {
        const initialValues = getInitialValues(activeTab);
        form.setFieldsValue(initialValues);
      }, 100);
    }
  }, [open, inspectionPolicies, activeTab, form]);

  const {
    data: listInspectionPolicies,
    loading,
    refresh,
  } = useRequest(inspection.listInspectionPolicies, {
    defaultParams: [{} as any],
  });

  const { run: triggerInspection } = useRequest(inspection.triggerInspection, {
    manual: true,
    onSuccess: () => {
      message.success(
        intl.formatMessage({
          id: 'src.pages.Inspection.TriggerInspectionSuccess',
          defaultMessage: '发起巡检成功',
        }),
      );
      refresh();
    },
  });
  const { run: createOrUpdateInspectionPolicy, loading: saveLoading } =
    useRequest(inspection.createOrUpdateInspectionPolicy, {
      manual: true,
      onSuccess: () => {
        message.success(
          intl.formatMessage({
            id: 'src.pages.Inspection.SaveInspectionConfigSuccess',
            defaultMessage: '保存巡检配置成功',
          }),
        );
        setOpen(false);
        form.resetFields();
        setActiveTab('basic');
        setInspectionPolicies({});
        refresh();
      },
    });
  const { run: deleteInspectionPolicy } = useRequest(
    inspection.deleteInspectionPolicy,
    {
      manual: true,
      onSuccess: () => {
        message.success(
          intl.formatMessage({
            id: 'src.pages.Inspection.DeleteInspectionConfigSuccess',
            defaultMessage: '删除巡检配置成功',
          }),
        );
        // 保持抽屉开启状态，不清空表单
        // 更新inspectionPolicies，移除被删除的调度配置
        if (inspectionPolicies?.scheduleConfig) {
          const updatedScheduleConfig =
            inspectionPolicies.scheduleConfig.filter(
              (config: any) => config.scenario !== activeTab,
            );
          setInspectionPolicies({
            ...inspectionPolicies,
            scheduleConfig: updatedScheduleConfig,
          });
        }
        // 清空当前tab的保存数据
        setAllTabData((prev) => {
          const newData = { ...prev };
          delete newData[activeTab];
          return newData;
        });
        // 重置当前tab的表单
        form.resetFields();
        // 刷新列表数据
        refresh();
      },
    },
  );

  const dataSource = listInspectionPolicies?.data || [];

  // 手动增加clusterName和namespace，便于搜索
  const realData = dataSource?.map((item: any) => {
    return {
      ...item,
      clusterName: item?.obCluster?.clusterName || item?.obCluster?.name,
      namespace: `${item?.obCluster?.namespace}/${item?.obCluster?.name}`,
    };
  });

  const columns = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.ResourceName',
        defaultMessage: '资源名',
      }),
      dataIndex: 'namespace',
      ...getColumnSearchProps({
        dataIndex: 'namespace',
        searchInput: searchInput,
        setSearchText: setSearchText,
        setSearchedColumn: setSearchedColumn,
        searchText: searchText,
        searchedColumn: searchedColumn,
        arraySearch: false,
        symbol: '',
      }),

      render: (text: string, record: any) => {
        return (
          <Link to={`/cluster/${text}/${record?.obCluster?.clusterName}`}>
            <CustomTooltip text={text} width={120} />
          </Link>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.ClusterName',
        defaultMessage: '集群名',
      }),
      dataIndex: 'clusterName',
      ...getColumnSearchProps({
        dataIndex: 'clusterName',
        searchInput: searchInput,
        setSearchText: setSearchText,
        setSearchedColumn: setSearchedColumn,
        searchText: searchText,
        searchedColumn: searchedColumn,
        arraySearch: false,
        symbol: '',
      }),
      render: (text: string) => {
        return <CustomTooltip text={text} width={80} />;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.BasicInspection',
        defaultMessage: '基础巡检',
      }),
      dataIndex: 'basicInspection',
      sorter: (a: any, b: any) => {
        // 按基础巡检的最新时间排序
        const aRepo = a.latestReports?.find(
          (item: any) => item?.scenario === 'basic',
        );
        const bRepo = b.latestReports?.find(
          (item: any) => item?.scenario === 'basic',
        );
        const aTime = aRepo?.finishTime || 0;
        const bTime = bRepo?.finishTime || 0;
        return bTime - aTime; // 最新的在前
      },
      render: (text: any, record: any) => {
        const repo = record.latestReports?.find(
          (item: any) => item?.scenario === 'basic',
        );
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        const id = `${repo?.namespace}/${repo?.name}`;

        const showContent = record?.scheduleConfig?.find(
          (item: any) => item?.scenario === 'basic',
        );

        return (
          <div style={{ lineHeight: 1.8 }}>
            {showContent ? (
              <>
                {repo?.finishTime && (
                  <div
                    style={{ marginBottom: 8, color: 'rgba(0, 0, 0, 0.65)' }}
                  >
                    {intl.formatMessage(
                      {
                        id: 'src.pages.Inspection.InspectionTime',
                        defaultMessage: '巡检时间：{time}',
                      },
                      { time: formatTime(repo?.finishTime) },
                    )}
                  </div>
                )}
                <div style={{ marginBottom: 8 }}>
                  <Space size={8} wrap>
                    <span style={{ fontWeight: 500, fontSize: 13 }}>
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.InspectionResult',
                        defaultMessage: '巡检结果：',
                      })}
                    </span>
                    <span
                      style={{ color: 'rgba(166,29,36,1)', fontWeight: 500 }}
                    >
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Critical',
                          defaultMessage: '高{count}',
                        },
                        { count: criticalCount || 0 },
                      )}
                    </span>
                    <span
                      style={{ color: token.colorWarning, fontWeight: 500 }}
                    >
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Moderate',
                          defaultMessage: '中{count}',
                        },
                        { count: moderateCount || 0 },
                      )}
                    </span>
                    <span style={{ color: token.colorError, fontWeight: 500 }}>
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Failed',
                          defaultMessage: '失败{count}',
                        },
                        { count: failedCount || 0 },
                      )}
                    </span>
                  </Space>
                </div>
                <div>
                  <Space size={12}>
                    <a
                      style={{
                        pointerEvents: !repo ? 'none' : 'auto',
                        opacity: !repo ? 0.5 : 1,
                        fontSize: 13,
                      }}
                      onClick={() => {
                        if (repo) {
                          history.push(`/inspection/report/${id}`);
                        }
                      }}
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.ViewReport',
                        defaultMessage: '查看报告',
                      })}
                    </a>
                    <a
                      style={{ fontSize: 13 }}
                      onClick={() =>
                        Modal.confirm({
                          title: intl.formatMessage({
                            id: 'src.pages.Inspection.ConfirmTriggerBasicInspection',
                            defaultMessage: '确定要发起基础巡检吗？',
                          }),
                          onOk: () => {
                            triggerInspection(
                              record.obCluster.namespace,
                              record.obCluster.name,
                              'basic',
                            );
                          },
                        })
                      }
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.TriggerInspectionNow',
                        defaultMessage: '立即巡检',
                      })}
                    </a>
                  </Space>
                </div>
              </>
            ) : (
              <span style={{ color: 'rgba(0, 0, 0, 0.25)' }}>-</span>
            )}
          </div>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.PerformanceInspection',
        defaultMessage: '性能巡检',
      }),
      dataIndex: 'performanceInspection',
      sorter: (a: any, b: any) => {
        // 按性能巡检的最新时间排序
        const aRepo = a.latestReports?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const bRepo = b.latestReports?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const aTime = aRepo?.finishTime || 0;
        const bTime = bRepo?.finishTime || 0;
        return bTime - aTime; // 最新的在前
      },
      render: (text: any, record: any) => {
        const showContent = record?.scheduleConfig?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const repo = record.latestReports?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        const id = `${repo?.namespace}/${repo?.name}`;

        return (
          <div style={{ lineHeight: 1.8 }}>
            {showContent ? (
              <>
                {repo?.finishTime && (
                  <div
                    style={{ marginBottom: 8, color: 'rgba(0, 0, 0, 0.65)' }}
                  >
                    {intl.formatMessage(
                      {
                        id: 'src.pages.Inspection.InspectionTime',
                        defaultMessage: '巡检时间：{time}',
                      },
                      { time: formatTime(repo?.finishTime) },
                    )}
                  </div>
                )}
                <div style={{ marginBottom: 8 }}>
                  <Space size={8} wrap>
                    <span style={{ fontWeight: 500, fontSize: 13 }}>
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.InspectionResult',
                        defaultMessage: '巡检结果：',
                      })}
                    </span>
                    <span
                      style={{ color: 'rgba(166,29,36,1)', fontWeight: 500 }}
                    >
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Critical',
                          defaultMessage: '高{count}',
                        },
                        { count: criticalCount || 0 },
                      )}
                    </span>
                    <span
                      style={{ color: token.colorWarning, fontWeight: 500 }}
                    >
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Moderate',
                          defaultMessage: '中{count}',
                        },
                        { count: moderateCount || 0 },
                      )}
                    </span>
                    <span style={{ color: token.colorError, fontWeight: 500 }}>
                      {intl.formatMessage(
                        {
                          id: 'src.pages.Inspection.Failed',
                          defaultMessage: '失败{count}',
                        },
                        { count: failedCount || 0 },
                      )}
                    </span>
                  </Space>
                </div>
                <div>
                  <Space size={12}>
                    <a
                      style={{
                        pointerEvents: !repo ? 'none' : 'auto',
                        opacity: !repo ? 0.5 : 1,
                        fontSize: 13,
                      }}
                      onClick={() => {
                        if (repo) {
                          history.push(`/inspection/report/${id}`);
                        }
                      }}
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.ViewReport',
                        defaultMessage: '查看报告',
                      })}
                    </a>
                    <a
                      style={{ fontSize: 13 }}
                      onClick={() => {
                        Modal.confirm({
                          title: intl.formatMessage({
                            id: 'src.pages.Inspection.ConfirmTriggerPerformanceInspection',
                            defaultMessage: '确定要发起性能巡检吗？',
                          }),
                          onOk: () => {
                            triggerInspection(
                              record.obCluster.namespace,
                              record.obCluster.name,
                              'performance',
                            );
                          },
                        });
                      }}
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Inspection.TriggerInspectionNow',
                        defaultMessage: '立即巡检',
                      })}
                    </a>
                  </Space>
                </div>
              </>
            ) : (
              <span style={{ color: 'rgba(0, 0, 0, 0.25)' }}>-</span>
            )}
          </div>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.ScheduleStatus',
        defaultMessage: '调度状态',
      }),
      dataIndex: 'status',

      render: (text: string) => {
        const content =
          text === 'enabled'
            ? intl.formatMessage({
                id: 'src.pages.Inspection.Enabled',
                defaultMessage: '已启用',
              })
            : intl.formatMessage({
                id: 'src.pages.Inspection.Disabled',
                defaultMessage: '未启用',
              });
        const color = text === 'enabled' ? 'success' : 'default';
        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Inspection.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'opeation',
      render: (_: any, record: any) => {
        return (
          <a
            onClick={() => {
              setOpen(true);
              setInspectionPolicies(record);
              // 清空所有tab的保存数据，确保每次打开都是干净的状态
              setAllTabData({});
              // 重置表单状态
              form.resetFields();
              setActiveTab('basic');
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Inspection.ScheduleConfig',
              defaultMessage: '调度配置',
            })}
          </a>
        );
      },
    },
  ];

  // 获取指定tab的初始值
  const getInitialValues = (tabKey: string) => {
    const repo = inspectionPolicies?.scheduleConfig?.find(
      (item: any) => item?.scenario === tabKey,
    );

    // 尝试从不同的字段名获取cron表达式
    const cronExpression = repo?.crontab || repo?.schedule || '';

    const parseResult = parseCronExpression(cronExpression);

    // 如果解析失败，使用默认值
    const schedule = parseResult.success ? parseResult.data : null;

    const getScheduleMode = () => {
      if (schedule?.dayOfMonth && schedule.dayOfMonth !== '*') return 'Monthly';
      if (schedule?.dayOfWeek && schedule.dayOfWeek !== '*') return 'Weekly';
      return 'Dayly';
    };

    const getScheduleDays = () => {
      if (schedule?.dayOfMonth && schedule.dayOfMonth !== '*') {
        // 处理多个日期的情况，如 "1,15" 或 "1-5"
        if (schedule.dayOfMonth.includes(',')) {
          return schedule.dayOfMonth
            .split(',')
            .map((day: string) => toNumber(day));
        }
        if (schedule.dayOfMonth.includes('-')) {
          const [start] = schedule.dayOfMonth
            .split('-')
            .map((day: string) => toNumber(day));
          return [start]; // 只取第一个值作为示例
        }
        return [toNumber(schedule.dayOfMonth)];
      }
      if (schedule?.dayOfWeek && schedule.dayOfWeek !== '*') {
        // 处理多个星期的情况，如 "1,3,5" 或 "1-5"
        if (schedule.dayOfWeek.includes(',')) {
          return schedule.dayOfWeek.split(',').map((day: string) => {
            const dayNum = toNumber(day);
            // 将cron的0-6转换为前端的1-7格式
            // cron：0=周日, 1=周一, ..., 6=周六
            // 前端：1=周一, 2=周二, ..., 7=周日
            if (dayNum === 0) return 7; // 周日：0 -> 7
            return dayNum; // 其他天：1-6 -> 1-6
          });
        }
        if (schedule.dayOfWeek.includes('-')) {
          const [start] = schedule.dayOfWeek
            .split('-')
            .map((day: string) => toNumber(day));
          // 转换单个值
          const dayNum = start;
          if (dayNum === 0) return [7]; // 周日：0 -> 7
          return [dayNum]; // 其他天：1-6 -> 1-6
        }
        const dayNum = toNumber(schedule.dayOfWeek);
        // 转换单个值
        if (dayNum === 0) return [7]; // 周日：0 -> 7
        return [dayNum]; // 其他天：1-6 -> 1-6
      }
      return [];
    };

    const getScheduleTime = () => {
      if (schedule?.hour !== undefined && schedule?.minute !== undefined) {
        // 确保时间格式正确，补零
        const hour = String(schedule.hour).padStart(2, '0');
        const minute = String(schedule.minute).padStart(2, '0');
        const timeString = `${hour}:${minute}`;
        return dayjs(timeString, TIME_FORMAT_WITHOUT_SECOND);
      }

      // 如果没有现有配置，返回默认时间（凌晨2点）
      return null;
    };

    // 优先使用保存的数据，如果没有则使用解析的数据
    const savedData = allTabData[tabKey as keyof typeof allTabData];

    if (savedData) {
      return savedData;
    }

    // 如果没有现有配置，返回默认值（日调度，凌晨2点）
    if (!repo) {
      return {
        scheduleDates: {
          mode: 'Dayly',
          days: [],
        },
        scheduleTime: undefined,
      };
    }

    return {
      scheduleDates: {
        mode: getScheduleMode(),
        days: getScheduleDays(),
      },
      scheduleTime: getScheduleTime(),
    };
  };

  const onChange = (activeKey: string) => {
    // 保存当前tab的表单数据
    const currentFormData = form.getFieldsValue();
    if (currentFormData.scheduleTime || currentFormData.scheduleDates) {
      setAllTabData((prev) => ({
        ...prev,
        [activeTab]: currentFormData,
      }));
    }

    setActiveTab(activeKey);
    // 先重置表单，清除所有字段值
    form.resetFields();
    // 然后设置新tab的初始值
    const initialValues = getInitialValues(activeKey);
    form.setFieldsValue(initialValues);
  };

  // 从表单数据生成cron表达式
  const generateCronFromFormData = (scheduleTime: any, scheduleDates: any) => {
    if (!scheduleTime || !scheduleDates) {
      return null;
    }

    const hour = scheduleTime.hour();
    const minute = scheduleTime.minute();

    let dayOfMonth = '*';
    let dayOfWeek = '*';
    const month = '*';

    switch (scheduleDates.mode) {
      case 'Dayly':
        // 每天执行: 0 2 * * *
        break;
      case 'Weekly':
        // 每周执行: 0 2 * * 1 (1=周一)
        if (scheduleDates.days && scheduleDates.days.length > 0) {
          // 将前端的1-7转换为cron的0-6格式
          // 前端：1=周一, 2=周二, ..., 7=周日
          // cron：0=周日, 1=周一, ..., 6=周六
          const convertedDays = scheduleDates.days.map((day: number) => {
            if (day === 7) return 0; // 周日：7 -> 0
            return day; // 其他天：1-6 -> 1-6
          });
          dayOfWeek = convertedDays.join(',');
        }
        break;
      case 'Monthly':
        // 每月执行: 0 2 3 * * (3号执行)
        if (scheduleDates.days && scheduleDates.days.length > 0) {
          dayOfMonth = scheduleDates.days.join(',');
        }
        break;
      default:
        return null;
    }

    const cronExpression = `${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`;

    // 获取浏览器当前时区信息
    const timezoneInfo = getTimezoneInfo();
    return {
      cron: cronExpression,
      timeZone: timezoneInfo.name,
    };
  };

  // 处理删除巡检的函数
  const handleDeleteInspection = (repo: any, tabKey: string) => {
    const getInspectionTypeName = () => {
      return tabKey === 'basic'
        ? intl.formatMessage({
            id: 'src.pages.Inspection.Basic',
            defaultMessage: '基础',
          })
        : intl.formatMessage({
            id: 'src.pages.Inspection.Performance',
            defaultMessage: '性能',
          });
    };

    Modal.confirm({
      title: intl.formatMessage(
        {
          id: 'src.pages.Inspection.ConfirmDeleteInspection',
          defaultMessage: '确定要删除{type}巡检吗？',
        },
        { type: getInspectionTypeName() },
      ),
      onOk: () => {
        // 从 inspectionPolicies 中获取正确的 namespace 和 name
        const namespace = inspectionPolicies?.obCluster?.namespace;
        const name = inspectionPolicies?.obCluster?.name;
        deleteInspectionPolicy(namespace, name, repo?.scenario);
      },
    });
  };

  const scheduleValue = Form.useWatch(['scheduleDates'], form);
  const content = (tabKey: string) => {
    // 根据tabKey获取对应的调度配置
    const repo = inspectionPolicies?.scheduleConfig?.find(
      (item: any) => item?.scenario === tabKey,
    );

    return (
      <Form
        form={form}
        initialValues={getInitialValues(tabKey)}
        key={`form-${tabKey}`} // 添加key确保每个tab的表单是独立的
      >
        <SchduleSelectFormItem
          form={form}
          scheduleValue={scheduleValue}
          type="inspection"
        />
        <h4>
          {intl.formatMessage({
            id: 'src.pages.Inspection.ScheduleTime',
            defaultMessage: '调度时间',
          })}
        </h4>
        <Form.Item
          name={['scheduleTime']}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Inspection.PleaseSelectScheduleTime',
                defaultMessage: '请选择调度时间',
              }),
            },
          ]}
        >
          <TimePicker format={TIME_FORMAT_WITHOUT_SECOND} />
        </Form.Item>
        {repo && (
          <Form.Item>
            <Button
              style={{ color: 'red' }}
              onClick={() => handleDeleteInspection(repo, tabKey)}
            >
              {intl.formatMessage({
                id: 'src.pages.Inspection.Delete',
                defaultMessage: '删除',
              })}
            </Button>
          </Form.Item>
        )}
      </Form>
    );
  };

  // 巡检类型配置
  const inspectionTypes = [
    {
      key: 'basic',
      label: intl.formatMessage({
        id: 'src.pages.Inspection.BasicInspection',
        defaultMessage: '基础巡检',
      }),
    },
    {
      key: 'performance',
      label: intl.formatMessage({
        id: 'src.pages.Inspection.PerformanceInspection',
        defaultMessage: '性能巡检',
      }),
    },
  ];

  // 生成tab项
  const items = inspectionTypes.map(({ key, label }) => ({
    key,
    label: (
      <Space>
        <span>{label}</span>
        {inspectionPolicies?.scheduleConfig?.find(
          (item: any) => item.scenario === key,
        ) && <CheckCircleFilled style={{ color: token.colorSuccess }} />}
      </Space>
    ),
    children: content(key),
  }));

  return (
    <>
      <Table dataSource={realData} columns={columns} loading={loading} />
      <Drawer
        title={intl.formatMessage(
          {
            id: 'src.pages.Inspection.InspectionScheduleConfig',
            defaultMessage: '{namespace}/{name} 巡检调度配置',
          },
          {
            namespace: inspectionPolicies?.obCluster?.namespace || '',
            name: inspectionPolicies?.obCluster?.name || '',
          },
        )}
        onClose={() => {
          setOpen(false);
          form.resetFields();
          setActiveTab('basic');
          // 清空所有tab的保存数据
          setAllTabData({});
          // 不清空 inspectionPolicies，保持数据用于下次打开
        }}
        open={open}
        footer={
          <Space>
            <Button
              onClick={() => {
                setOpen(false);
                form.resetFields();
                setActiveTab('basic');
                setInspectionPolicies({});
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Inspection.Cancel',
                defaultMessage: '取消',
              })}
            </Button>
            <Button
              type="primary"
              loading={saveLoading}
              onClick={async () => {
                // 先验证当前表单
                try {
                  await form.validateFields();
                } catch (error) {
                  // 表单验证失败，显示错误信息
                  message.error(
                    intl.formatMessage({
                      id: 'src.pages.Inspection.PleaseCompleteScheduleConfig',
                      defaultMessage: '请完善调度配置信息',
                    }),
                  );
                  return;
                }

                // 收集所有tab的配置数据
                const allScheduleConfigs = [];

                // 保存当前tab的数据
                const currentFormData = form.getFieldsValue();
                if (
                  currentFormData.scheduleTime &&
                  currentFormData.scheduleDates
                ) {
                  const currentCronData = generateCronFromFormData(
                    currentFormData.scheduleTime,
                    currentFormData.scheduleDates,
                  );

                  if (currentCronData) {
                    allScheduleConfigs.push({
                      scenario: activeTab,
                      schedule: currentCronData.cron,
                      timeZone: currentCronData.timeZone,
                    });
                  }
                }

                // 添加其他已保存的tab数据
                Object.keys(allTabData).forEach((tabKey) => {
                  if (tabKey !== activeTab) {
                    const tabData = allTabData[tabKey];
                    if (tabData.scheduleTime && tabData.scheduleDates) {
                      const cronData = generateCronFromFormData(
                        tabData.scheduleTime,
                        tabData.scheduleDates,
                      );

                      if (cronData) {
                        allScheduleConfigs.push({
                          scenario: tabKey,
                          schedule: cronData.cron,
                          timeZone: cronData.timeZone,
                        });
                      }
                    }
                  }
                });

                // 添加其他已配置的tab（从inspectionPolicies中获取）
                if (inspectionPolicies?.scheduleConfig) {
                  inspectionPolicies.scheduleConfig.forEach((config: any) => {
                    // 只添加没有在当前表单和保存数据中的配置
                    const isInCurrentForm = activeTab === config.scenario;
                    const isInSavedData = allTabData[config.scenario];

                    if (!isInCurrentForm && !isInSavedData) {
                      // 获取时区信息
                      const timezoneInfo = getTimezoneInfo();
                      allScheduleConfigs.push({
                        scenario: config.scenario,
                        schedule: config.schedule,
                        timeZone: timezoneInfo.name,
                      });
                    }
                  });
                }

                // 检查是否有配置数据
                if (allScheduleConfigs.length === 0) {
                  message.error(
                    intl.formatMessage({
                      id: 'src.pages.Inspection.PleaseConfigureAtLeastOneInspectionType',
                      defaultMessage: '请至少配置一个巡检类型',
                    }),
                  );
                  return;
                }

                // 构建请求体
                const body: InspectionPolicy = {
                  obCluster: inspectionPolicies?.obCluster,
                  status: inspectionPolicies?.status || 'enabled',
                  scheduleConfig: allScheduleConfigs.map((config) => ({
                    scenario: config.scenario as InspectionInspectionScenario,
                    schedule: config.schedule,
                    timeZone: config.timeZone,
                  })),
                };
                createOrUpdateInspectionPolicy(body);
              }}
            >
              {saveLoading
                ? intl.formatMessage({
                    id: 'src.pages.Inspection.Saving',
                    defaultMessage: '保存中...',
                  })
                : intl.formatMessage({
                    id: 'src.pages.Inspection.Confirm',
                    defaultMessage: '确定',
                  })}
            </Button>
          </Space>
        }
      >
        <Tabs activeKey={activeTab} items={items} onChange={onChange} />
      </Drawer>
    </>
  );
}
