import { inspection } from '@/api';
import CustomTooltip from '@/components/CustomTooltip';
import { getColumnSearchProps } from '@/utils/component';
import { formatTime } from '@/utils/datetime';
import { CheckCircleFilled } from '@ant-design/icons';
import { theme } from '@oceanbase/design';
import { Link } from '@umijs/max';
import { useRequest } from 'ahooks';
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
} from 'antd';
import { isEmpty } from 'lodash';
import { useRef, useState } from 'react';
import SchduleSelectFormItem from '../Tenant/Detail/NewBackup/SchduleSelectFormItem';

export default function InspectionList() {
  const { token } = theme.useToken();
  const [searchText, setSearchText] = useState('');
  const [searchedColumn, setSearchedColumn] = useState('');
  const searchInput = useRef<InputRef>(null);
  const [open, setOpen] = useState(false);
  const [inspectionPolicies, setInspectionPolicies] = useState({});

  const { data: listInspectionPolicies } = useRequest(
    inspection.listInspectionPolicies,
    {
      defaultParams: [{}],
    },
  );

  const { run: triggerInspection } = useRequest(inspection.triggerInspection, {
    manual: true,
  });
  const { run: deleteInspectionPolicy } = useRequest(
    inspection.deleteInspectionPolicy,
    {
      manual: true,
    },
  );

  console.log('listInspectionPolicies', listInspectionPolicies);

  const dataSource = listInspectionPolicies?.data || [];

  // const dataSource = [
  //   {
  //     key: '1',
  //     name: '胡彦斌',
  //     age: 32,
  //     obCluster: {
  //       namespace: 'testnamespace',
  //       name: 'testname',
  //       clusterName: 'testclusterName',
  //     },
  //   },
  //   {
  //     key: '2',
  //     name: '胡彦祖',
  //     age: 42,
  //     address: '西湖区湖底公园1号',
  //   },
  // ];

  const columns = [
    {
      title: '资源名',
      dataIndex: 'obCluster',
      ...getColumnSearchProps({
        dataIndex: 'obCluster',
        searchInput: searchInput,
        setSearchText: setSearchText,
        setSearchedColumn: setSearchedColumn,
        searchText: searchText,
        searchedColumn: searchedColumn,
        arraySearch: true,
        symbol: '=',
      }),

      render: (text) => {
        return (
          <Link
            to={`/cluster/${text?.namespace}/${text?.name}/${text?.clusterName}`}
          >
            <CustomTooltip
              text={`${text?.namespace}/${text?.name}`}
              width={100}
            />
          </Link>
        );
      },
    },
    {
      title: '集群名',
      dataIndex: 'obCluster',
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
        return <CustomTooltip text={text?.clusterName} width={60} />;
      },
    },
    {
      title: '基础巡检',
      dataIndex: 'latestReports',
      sorter: true,
      render: (text) => {
        const repo = text?.find((item) => item?.scenario === 'basic');
        const { negligibleCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        return (
          <div>
            <div>{`巡检时间：${formatTime(repo?.finishTime)}`}</div>
            <Space size={6}>
              <span>巡检结果：</span>
              <span style={{ color: 'red' }}>{`高${criticalCount || 0}`}</span>
              <span style={{ color: 'orange' }}>{`中${
                moderateCount || 0
              }`}</span>
              <span style={{ color: 'orange' }}>{`低${
                negligibleCount || 0
              }`}</span>

              <Link to={`/inspection/report/${1}`} target="_blank">
                查看报告
              </Link>
              <a
                onClick={() =>
                  Modal.confirm({
                    title: '确定要发起基础巡检吗？',
                    onOk: () => {
                      // triggerInspection(repo?.obCluster?.namespace);
                      triggerInspection({
                        namespace: repo?.obCluster?.namespace,
                        name: repo?.obCluster?.name,
                        scenario: repo?.scenario,
                      });
                    },
                  })
                }
              >
                立即巡检
              </a>
            </Space>
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
        const { negligibleCount, criticalCount, moderateCount } =
          repo?.resultStatistics || {};

        return (
          <div>
            <div>{`巡检时间：${formatTime(repo?.finishTime)}`}</div>
            <Space size={6}>
              <span>巡检结果：</span>
              <span style={{ color: 'red' }}>{`高${criticalCount || 0}`}</span>
              <span style={{ color: 'orange' }}>{`中${
                moderateCount || 0
              }`}</span>
              <span style={{ color: 'orange' }}>{`低${
                negligibleCount || 0
              }`}</span>
              <Link to={`/inspection/report/${1}`} target="_blank">
                查看报告
              </Link>
              <a
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

  const onChange = () => {
    form.resetFields();
  };

  const [form] = Form.useForm();

  const initialValues = {
    scheduleDates: {
      mode: 'Monthly',
      days: [],
    },
  };
  const scheduleValue = Form.useWatch(['scheduleDates'], form);
  const content = () => {
    const performanceRepo = inspectionPolicies?.latestReports?.find(
      (item) => item?.scenario === 'performance',
    );
    const basicRepo = inspectionPolicies?.latestReports?.find(
      (item) => item?.scenario === 'basic',
    );

    return (
      <Form form={form} initialValues={initialValues}>
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
          <TimePicker format={'HH:mm'} />
        </Form.Item>
        {!isEmpty(performanceRepo) || !isEmpty(basicRepo) ? (
          <Form.Item>
            <Button
              style={{ color: 'red' }}
              onClick={() => {
                Modal.confirm({
                  title: `确定要删除${
                    isEmpty(performanceRepo) && !isEmpty(basicRepo)
                      ? '基础'
                      : '性能'
                  }巡检吗？`,
                  onOk: () => {
                    deleteInspectionPolicy({
                      namespace:
                        isEmpty(performanceRepo) && !isEmpty(basicRepo)
                          ? basicRepo?.obCluster?.namespace
                          : performanceRepo?.obCluster?.namespace,
                      name:
                        isEmpty(performanceRepo) && !isEmpty(basicRepo)
                          ? basicRepo?.obCluster?.name
                          : performanceRepo?.obCluster?.name,
                      scenario:
                        isEmpty(performanceRepo) && !isEmpty(basicRepo)
                          ? basicRepo?.scenario
                          : performanceRepo?.scenario,
                    });
                  },
                });
              }}
            >
              删除
            </Button>
          </Form.Item>
        ) : null}
      </Form>
    );
  };

  const items = [
    {
      key: 'basic',
      label: (
        <Space>
          <span>基础巡检</span>
          {inspectionPolicies?.scheduleConfig?.find(
            (item) => item.scenario === 'basic',
          ) ? (
            <CheckCircleFilled style={{ color: token.colorSuccess }} />
          ) : null}
        </Space>
      ),
      children: content('basic'),
    },
    {
      key: 'performance',
      label: (
        <Space>
          <span>性能巡检</span>
          {inspectionPolicies?.scheduleConfig?.find(
            (item) => item.scenario === 'performance',
          ) ? (
            <CheckCircleFilled style={{ color: token.colorSuccess }} />
          ) : null}
        </Space>
      ),
      children: content('performance'),
    },
  ];
  console.log('inspectionPolicies', inspectionPolicies);
  return (
    <>
      <Table dataSource={dataSource} columns={columns} />
      <Drawer
        title={`${inspectionPolicies?.obCluster?.namespace}/${inspectionPolicies?.obCluster?.name} 巡检调度配置`}
        onClose={() => {
          setOpen(false);
          form.resetFields();
        }}
        open={open}
        footer={
          <Space>
            <Button
              onClick={() => {
                setOpen(false);
                form.resetFields();
              }}
            >
              取消
            </Button>
            <Button
              type="primary"
              onClick={() => {
                setOpen(false);
                form.resetFields();
              }}
            >
              确定
            </Button>
          </Space>
        }
      >
        <Tabs defaultActiveKey="basic" items={items} onChange={onChange} />
      </Drawer>
    </>
  );
}
