import { inspection } from '@/api';
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
import { useRef, useState } from 'react';
import SchduleSelectFormItem from '../Tenant/Detail/NewBackup/SchduleSelectFormItem';

export default function InspectionList() {
  const { token } = theme.useToken();
  const [searchText, setSearchText] = useState('');
  const [searchedColumn, setSearchedColumn] = useState('');
  const searchInput = useRef<InputRef>(null);
  const [open, setOpen] = useState(false);
  const [inspectionPolicies, setInspectionPolicies] = useState({});

  const { data: listInspectionPolicies, loading } = useRequest(
    inspection.listInspectionPolicies,
    {
      defaultParams: [{}],
    },
  );

  const { run: triggerInspection } = useRequest(inspection.triggerInspection, {
    manual: true,
  });

  // 手写巡检调度配置请求
  const { run: createOrUpdateInspectionPolicy, loading: saveLoading } =
    useRequest(
      async (body) => {
        const response = await fetch('/api/v1/inspection/policies', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(body),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(
            errorData.message ||
              `HTTP ${response.status}: ${response.statusText}`,
          );
        }

        return response.json();
      },
      {
        manual: true,
        onSuccess: (data) => {
          console.log('保存成功:', data);
          message.success('调度配置保存成功');
          setOpen(false);
          form.resetFields();
          setActiveTab('basic');
          setInspectionPolicies({});
        },
        onError: (error) => {
          console.error('保存失败:', error);
          message.error(error.message);
        },
      },
    );

  // 手写删除巡检请求
  const { run: deleteInspectionPolicy } = useRequest(
    async (params) => {
      const { namespace, name, scenario } = params;
      const response = await fetch(
        `/api/v1/inspection/policies?namespace=${namespace}&name=${name}&scenario=${scenario}`,
        {
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
          },
        },
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            `HTTP ${response.status}: ${response.statusText}`,
        );
      }

      return response.json();
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('删除成功:', data);
        message.success('巡检配置删除成功');
        // 刷新列表
        listInspectionPolicies?.refresh?.();
      },
      onError: (error) => {
        console.error('删除失败:', error);
        message.error(error.message);
      },
    },
  );

  const dataSource = listInspectionPolicies?.data || [];

  // 手动增加clusterName和namespace，便于搜索
  const realData = dataSource?.map((item) => {
    return {
      ...item,
      clusterName: item?.obCluster?.clusterName || item?.obCluster?.name,
      namespace: `${item?.obCluster?.namespace}/${item?.obCluster?.name}`,
    };
  });
  console.log('real', realData);
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
        arraySearch: true,
        symbol: '/',
      }),

      render: (text, record) => {
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
        dataIndex: 'name',
        searchInput: searchInput,
        setSearchText: setSearchText,
        setSearchedColumn: setSearchedColumn,
        searchText: searchText,
        searchedColumn: searchedColumn,
      }),
      render: (text) => {
        return <CustomTooltip text={text} width={60} />;
      },
    },
    {
      title: '基础巡检',
      dataIndex: 'latestReports',
      sorter: true,
      render: (text) => {
        const repo = text?.find((item) => item?.scenario === 'basic');
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        const id = `${repo?.namespace}/${repo?.name}`;

        return (
          <div>
            {repo ? (
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
                  <span style={{ color: token.colorError }}>{`低${
                    failedCount || 0
                  }`}</span>
                  <a
                    disabled={!repo}
                    onClick={() => {
                      history.push(`/inspection/report/${id}`);
                    }}
                  >
                    查看报告
                  </a>
                  <a
                    disabled={!repo}
                    onClick={() =>
                      Modal.confirm({
                        title: '确定要发起基础巡检吗？',
                        onOk: () => {
                          triggerInspection(
                            repo?.obCluster?.namespace,
                            repo?.obCluster?.name,
                            repo?.scenario,
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
      render: (text) => {
        const repo = text?.find((item) => item?.scenario === 'performance');
        const { failedCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};
        console.log('repo', repo);
        const id = `${repo?.namespace}/${repo?.name}`;
        return (
          <div>
            {repo ? (
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
                  <span style={{ color: token.colorError }}>{`低${
                    failedCount || 0
                  }`}</span>
                  <a
                    disabled={!repo}
                    onClick={() => {
                      history.push(`/inspection/report/${id}`);
                    }}
                  >
                    查看报告
                  </a>
                  <a
                    disabled={!repo}
                    onClick={() => {
                      Modal.confirm({
                        title: '确定要发起性能巡检吗？',
                        onOk: () => {
                          triggerInspection({
                            namespace: repo?.obCluster?.namespace,
                            name: repo?.obCluster?.name,
                            scenario: repo?.scenario,
                          });
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

      render: (text) => {
        const content = text === 'enabled' ? '已启用' : '未启用';
        const color = text === 'enabled' ? 'success' : 'default';
        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: '操作',
      dataIndex: 'opeation',
      render: (_, record) => {
        return (
          <a
            onClick={() => {
              setOpen(true);
              setInspectionPolicies(record);
            }}
          >
            调度配置
          </a>
        );
      },
    },
  ];

  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('basic');

  // 获取指定tab的初始值
  const getInitialValues = (tabKey) => {
    const repo = inspectionPolicies?.scheduleConfig?.find(
      (item) => item?.scenario === tabKey,
    );

    const schedule = parseCronExpression(repo?.schedule).data;

    const getScheduleMode = () => {
      if (schedule?.dayOfMonth) return 'Monthly';
      if (schedule?.dayOfWeek) return 'Weekly';
      return 'Daily';
    };

    const getScheduleDays = () => {
      if (schedule?.dayOfMonth) return [toNumber(schedule.dayOfMonth)];
      if (schedule?.dayOfWeek) return [schedule.dayOfWeek];
      return [];
    };

    const getScheduleTime = () => {
      if (schedule?.hour) {
        return dayjs(
          `${schedule.hour}:${schedule.minute}`,
          TIME_FORMAT_WITHOUT_SECOND,
        );
      }
      return null;
    };

    return {
      scheduleDates: {
        mode: getScheduleMode(),
        days: getScheduleDays(),
      },
      scheduleTime: getScheduleTime(),
    };
  };

  const onChange = (activeKey) => {
    setActiveTab(activeKey);
    const initialValues = getInitialValues(activeKey);
    form.setFieldsValue(initialValues);
  };

  // 从表单数据生成cron表达式
  const generateCronFromFormData = (scheduleTime, scheduleDates) => {
    if (!scheduleTime || !scheduleDates) {
      return null;
    }

    const hour = scheduleTime.hour();
    const minute = scheduleTime.minute();

    let dayOfMonth = '*';
    let dayOfWeek = '*';
    const month = '*';

    switch (scheduleDates.mode) {
      case 'Daily':
        // 每天执行: 0 2 * * *
        break;
      case 'Weekly':
        // 每周执行: 0 2 * * 1 (1=周一)
        if (scheduleDates.days && scheduleDates.days.length > 0) {
          dayOfWeek = scheduleDates.days.join(',');
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
  const handleDeleteInspection = (repo, tabKey) => {
    const getInspectionTypeName = () => {
      return tabKey === 'basic' ? '基础' : '性能';
    };

    Modal.confirm({
      title: `确定要删除${getInspectionTypeName()}巡检吗？`,
      onOk: () => {
        deleteInspectionPolicy({
          namespace: repo?.obCluster?.namespace,
          name: repo?.obCluster?.name,
          scenario: repo?.scenario,
        });
      },
    });
  };

  const scheduleValue = Form.useWatch(['scheduleDates'], form);
  const content = (tabKey) => {
    // 根据tabKey获取对应的调度配置
    const repo = inspectionPolicies?.scheduleConfig?.find(
      (item) => item?.scenario === tabKey,
    );

    return (
      <Form form={form} initialValues={getInitialValues(tabKey)}>
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
          (item) => item.scenario === key,
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
                form.validateFields().then((values) => {
                  const { scheduleTime, scheduleDates } = values;

                  // 生成cron表达式
                  const cronExpression = generateCronFromFormData(
                    scheduleTime,
                    scheduleDates,
                  );

                  if (!cronExpression) {
                    message.error('无法生成有效的cron表达式');
                    return;
                  }

                  // 构建调度配置
                  const scheduleConfig = {
                    scenario: activeTab,
                    crontab: cronExpression,
                  };

                  // 构建请求体
                  const body = {
                    ...inspectionPolicies,
                    scheduleConfig: [scheduleConfig],
                  };

                  console.log('cron表达式:', cronExpression);
                  console.log('请求体:', body);
                  console.log('inspectionPolicies:', inspectionPolicies);

                  // 调用API
                  createOrUpdateInspectionPolicy(body);
                });
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
