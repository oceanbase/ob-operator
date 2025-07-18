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
    defaultParams: [{}],
  });

  const { run: triggerInspection } = useRequest(inspection.triggerInspection, {
    manual: true,
    onSuccess: () => {
      message.success('发起巡检成功');
      refresh();
    },
  });
  const { run: createOrUpdateInspectionPolicy, loading: saveLoading } =
    useRequest(inspection.createOrUpdateInspectionPolicy, {
      manual: true,
      onSuccess: () => {
        message.success('保存巡检配置成功');
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
        message.success('删除巡检配置成功');
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
      title: '资源名',
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
            <CustomTooltip text={text} width={100} />
          </Link>
        );
      },
    },
    {
      title: '集群名',
      dataIndex: 'clusterName',
      width: 80,
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
        return <CustomTooltip text={text} width={60} />;
      },
    },
    {
      title: '基础巡检',
      dataIndex: 'latestReports',
      sorter: true,
      render: (text: any, record: any) => {
        const repo = text?.find((item: any) => item?.scenario === 'basic');
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        const id = `${repo?.namespace}/${repo?.name}`;

        const showContent = record?.scheduleConfig?.find(
          (item: any) => item?.scenario === 'basic',
        );
        return (
          <div>
            {showContent ? (
              <>
                <div>{`巡检时间：${formatTime(repo?.finishTime)}`}</div>
                <Space size={6}>
                  <span>巡检结果：</span>
                  <span style={{ color: 'rgba(166,29,36,1)' }}>{`高${
                    criticalCount || 0
                  }`}</span>
                  <span style={{ color: token.colorWarning }}>{`中${
                    moderateCount || 0
                  }`}</span>
                  <span style={{ color: token.colorError }}>{`失败${
                    failedCount || 0
                  }`}</span>
                  <a
                    style={{
                      pointerEvents: !repo ? 'none' : 'auto',
                      opacity: !repo ? 0.5 : 1,
                    }}
                    onClick={() => {
                      if (repo) {
                        history.push(`/inspection/report/${id}`);
                      }
                    }}
                  >
                    查看报告
                  </a>
                  <a
                    onClick={() =>
                      Modal.confirm({
                        title: '确定要发起基础巡检吗？',
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
                    立即巡检
                  </a>
                </Space>
              </>
            ) : (
              <> - </>
            )}
          </div>
        );
      },
    },
    {
      title: '性能巡检',
      dataIndex: 'latestReports',
      sorter: true,
      render: (text: any, record: any) => {
        const showContent = record?.scheduleConfig?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const repo = text?.find(
          (item: any) => item?.scenario === 'performance',
        );
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        const id = `${repo?.namespace}/${repo?.name}`;
        return (
          <div>
            {showContent ? (
              <>
                <div>{`巡检时间：${formatTime(repo?.finishTime)}`}</div>
                <Space size={6}>
                  <span>巡检结果：</span>
                  <span style={{ color: 'red' }}>{`高${
                    criticalCount || 0
                  }`}</span>
                  <span style={{ color: 'orange' }}>{`中${
                    moderateCount || 0
                  }`}</span>
                  <span style={{ color: token.colorError }}>{`失败${
                    failedCount || 0
                  }`}</span>
                  <a
                    style={{
                      pointerEvents: !repo ? 'none' : 'auto',
                      opacity: !repo ? 0.5 : 1,
                    }}
                    onClick={() => {
                      if (repo) {
                        history.push(`/inspection/report/${id}`);
                      }
                    }}
                  >
                    查看报告
                  </a>
                  <a
                    onClick={() => {
                      Modal.confirm({
                        title: '确定要发起性能巡检吗？',
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
                    立即巡检
                  </a>
                </Space>
              </>
            ) : (
              <>-</>
            )}
          </div>
        );
      },
    },
    {
      title: '调度状态',
      dataIndex: 'status',

      render: (text: string) => {
        const content = text === 'enabled' ? '已启用' : '未启用';
        const color = text === 'enabled' ? 'success' : 'default';
        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: '操作',
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
            调度配置
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

    return `${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`;
  };

  // 处理删除巡检的函数
  const handleDeleteInspection = (repo: any, tabKey: string) => {
    const getInspectionTypeName = () => {
      return tabKey === 'basic' ? '基础' : '性能';
    };

    Modal.confirm({
      title: `确定要删除${getInspectionTypeName()}巡检吗？`,
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
        <h4>调度时间</h4>
        <Form.Item
          name={['scheduleTime']}
          rules={[
            {
              required: true,
              message: '请选择调度时间',
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
              删除
            </Button>
          </Form.Item>
        )}
      </Form>
    );
  };

  // 巡检类型配置
  const inspectionTypes = [
    { key: 'basic', label: '基础巡检' },
    { key: 'performance', label: '性能巡检' },
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
        title={`${inspectionPolicies?.obCluster?.namespace}/${inspectionPolicies?.obCluster?.name} 巡检调度配置`}
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
              取消
            </Button>
            <Button
              type="primary"
              loading={saveLoading}
              onClick={() => {
                // 收集所有tab的配置数据
                const allScheduleConfigs = [];

                // 保存当前tab的数据
                const currentFormData = form.getFieldsValue();
                if (
                  currentFormData.scheduleTime &&
                  currentFormData.scheduleDates
                ) {
                  const currentCronExpression = generateCronFromFormData(
                    currentFormData.scheduleTime,
                    currentFormData.scheduleDates,
                  );

                  if (currentCronExpression) {
                    allScheduleConfigs.push({
                      scenario: activeTab,
                      schedule: currentCronExpression,
                    });
                  }
                }

                // 添加其他已保存的tab数据
                Object.keys(allTabData).forEach((tabKey) => {
                  if (tabKey !== activeTab) {
                    const tabData = allTabData[tabKey];
                    if (tabData.scheduleTime && tabData.scheduleDates) {
                      const cronExpression = generateCronFromFormData(
                        tabData.scheduleTime,
                        tabData.scheduleDates,
                      );

                      if (cronExpression) {
                        allScheduleConfigs.push({
                          scenario: tabKey,
                          schedule: cronExpression,
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
                      allScheduleConfigs.push({
                        scenario: config.scenario,
                        schedule: config.schedule,
                      });
                    }
                  });
                }

                // 检查是否有配置数据
                if (allScheduleConfigs.length === 0) {
                  message.error('请至少配置一个巡检类型');
                  return;
                }

                // 构建请求体
                const body: InspectionPolicy = {
                  obCluster: inspectionPolicies?.obCluster,
                  status: inspectionPolicies?.status || 'enabled',
                  scheduleConfig: allScheduleConfigs.map((config) => ({
                    scenario: config.scenario as InspectionInspectionScenario,
                    schedule: config.schedule,
                  })),
                };
                createOrUpdateInspectionPolicy(body);
              }}
            >
              {saveLoading ? '保存中...' : '确定'}
            </Button>
          </Space>
        }
      >
        <Tabs activeKey={activeTab} items={items} onChange={onChange} />
      </Drawer>
    </>
  );
}
