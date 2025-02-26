import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { useRequest } from 'ahooks';
import { Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import CollapsibleCard from '../CollapsibleCard';

import { EVENTSTABLE_STATUS_LIST } from '@/constants';
import { getEventsReq } from '@/services';
import { findByValue } from '@oceanbase/util';
import CustomTooltip from '../CustomTooltip';

interface DataType {
  key: React.Key;
  namespace: string;
  type: string;
  reason: string;
  object: string;
  firstOccur: string;
  lastSeen: string;
  count: number;
  message: string;
}

interface EventsTableProps {
  objectType?: API.EventObjectType;
  cardType?: 'card' | 'proCard';
  collapsible?: boolean;
  defaultExpand?: boolean;
  overView?: boolean; // 多个资源需要加上fiilter
  name?: string;
}

export default function EventsTable({
  objectType,
  cardType,
  collapsible = false,
  defaultExpand = false,
  name,
  overView,
}: EventsTableProps) {
  const defaultParams: API.EventParams = {};
  if (objectType) {
    defaultParams.objectType = objectType;
  }
  if (name) {
    defaultParams.name = name;
  }

  const { data, loading } = useRequest(getEventsReq, {
    defaultParams: [defaultParams],
  });

  const CustomCard = (props) => {
    const { title, loading } = props;

    return (
      <>
        {cardType === 'proCard' ? (
          <ProCard loading={loading} title={title} collapsible={collapsible}>
            {props.children}
          </ProCard>
        ) : (
          <CollapsibleCard
            loading={loading}
            defaultExpand={defaultExpand}
            title={title}
            collapsible={collapsible}
          >
            {props.children}
          </CollapsibleCard>
        )}
      </>
    );
  };

  const namespaceList = data?.map((item) => ({
    text: item.namespace,
    value: item.namespace,
  }));

  const namespaceFilterList = namespaceList?.reduce((prev, curr) => {
    const found = prev.find(
      (item) => item.text === curr.text && item.value === curr.value,
    );
    return found ? prev : [...prev, curr];
  }, []);

  const columns: ColumnsType<DataType> = [
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.Namespace',
        defaultMessage: '命名空间',
      }),
      dataIndex: 'namespace',
      key: 'namespace',
      width: 120,

      ...(overView
        ? {
            filters: namespaceFilterList,
            onFilter: (value: any, record) => {
              return record.namespace === value;
            },
          }
        : {}),

      // filters: namespaceFilterList,
      // onFilter: (value: any, record) => {
      //   return record.namespace === value;
      // },
      render: (text) => <span>{text}</span>,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.Type',
        defaultMessage: '类型',
      }),
      dataIndex: 'type',
      key: 'type',
      filters: [
        {
          text: intl.formatMessage({
            id: 'OBDashboard.components.EventsTable.Normal',
            defaultMessage: '正常',
          }),
          value: 'Normal',
        },
        {
          text: intl.formatMessage({
            id: 'OBDashboard.components.EventsTable.Warning',
            defaultMessage: '警告',
          }),
          value: 'Warning',
        },
      ],

      onFilter: (value: any, record) => {
        return record.type === value;
      },
      render: (text) => {
        const value = findByValue(EVENTSTABLE_STATUS_LIST, text);
        return <Tag color={value.badgeStatus}>{value.label}</Tag>;
      },
      width: 120,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.NumberOfOccurrences',
        defaultMessage: '发生次数',
      }),
      dataIndex: 'count',
      key: 'count',
      width: 130,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.FirstOccurrenceTime',
        defaultMessage: '第一次发生时间',
      }),
      dataIndex: 'firstOccur',
      key: 'firstOccur',
      width: 210,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.RecentOccurrenceTime',
        defaultMessage: '最近发生时间',
      }),
      dataIndex: 'lastSeen',
      key: 'lastSeen',
      width: 210,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.Cause',
        defaultMessage: '原因',
      }),
      dataIndex: 'reason',
      key: 'reason',
      width: 160,
      render: (text) => <span>{text || '-'}</span>,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.AssociatedObjects',
        defaultMessage: '关联对象',
      }),
      dataIndex: 'object',
      key: 'object',
      width: 150,
      render: (val) => <CustomTooltip text={val} width={120} />,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.components.EventsTable.Information',
        defaultMessage: '信息',
      }),
      dataIndex: 'message',
      key: 'message',
      width: 220,
      render: (val) => <CustomTooltip text={val} width={200} />,
    },
  ];

  return (
    <CustomCard
      loading={loading}
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'OBDashboard.components.EventsTable.Event',
            defaultMessage: '事件',
          })}
        </h2>
      }
    >
      <Table
        rowKey="id"
        pagination={{ simple: true }}
        columns={columns}
        dataSource={data}
      />
    </CustomCard>
  );
}
