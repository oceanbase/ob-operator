import { MODE_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Pie } from '@antv/g2plot';
import { Link } from '@umijs/max';
import { Button,Card,Col,Table,Tag,Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';

import { COLOR_MAP } from '@/constants';
import { useEffect } from 'react';
import styles from './index.less';
interface DataType {
  namespace: string;
  name: string;
  status: string;
  statusDetail: string;
  createTime: string;
  image: string;
  cpuPercent: string;
  memoryPercent: string;
  diskPercent: string;
  clusterName: string;
}

interface CanvasPieProps {
  percent: number;
  name: 'cpu' | 'memory' | 'disk';
}

interface ClusterListProps {
  handleAddCluster: () => void;
  clusterList: any;
  loading: boolean;
}

const { Text } = Typography;

const CanvasPie = ({ percent, name }: CanvasPieProps) => {
  const data = [
    {
      type: intl.formatMessage({
        id: 'OBDashboard.pages.Cluster.ClusterList.Used',
        defaultMessage: '已使用',
      }),
      value: percent,
    },
    {
      type: intl.formatMessage({
        id: 'OBDashboard.pages.Cluster.ClusterList.Unused',
        defaultMessage: '未使用',
      }),
      value: 100 - percent,
    },
  ];

  useEffect(() => {
    let pie = new Pie(`${name}-container`, {
      data,
      angleField: 'value',
      colorField: 'type',
      width: 100,
      height: 100,
      legend: false,
      label: {
        type: 'inner',
        offset: '-60%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 10,
          textAlign: 'center',
        },
      },
      color: ({ type }) => {
        if (
          type ===
          intl.formatMessage({
            id: 'OBDashboard.pages.Cluster.ClusterList.Used',
            defaultMessage: '已使用',
          })
        ) {
          if (name === 'cpu') {
            return 'rgb(111,143,243)';
          } else if (name === 'memory') {
            return 'rgb(125,213,169)';
          } else {
            return 'rgb(97,111,143)';
          }
        } else {
          return 'rgb(228,231,236)';
        }
      },
    });
    pie.render();
  }, []);

  return <div id={`${name}-container`}></div>;
};

const columns: ColumnsType<DataType> = [
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.ResourceName',
      defaultMessage: '资源名',
    }),
    dataIndex: 'name',
    key: 'name',
    render: (value, record) => (
      <Link to={`${record.namespace}/${record.name}/${record.clusterName}`}>
        {value}
      </Link>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Namespace',
      defaultMessage: '命名空间',
    }),
    dataIndex: 'namespace',
    key: 'namespace',
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.ClusterName',
      defaultMessage: '集群名',
    }),
    dataIndex: 'clusterName',
    key: 'clusterName',
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.ClusterMode',
      defaultMessage: '集群模式',
    }),
    dataIndex: 'mode',
    key: 'mode',
    render: (value) => <span>{MODE_MAP.get(value)?.text}</span>,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.pages.Cluster.ClusterList.NumberOfZones',
      defaultMessage: 'Zone 数量',
    }),
    dataIndex: 'zoneCount',
    key: 'zoneCount',
    render: (_, record) => <span>{record?.topology?.length}</span>,
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Image',
      defaultMessage: '镜像',
    }),
    dataIndex: 'image',
    key: 'image',
    render: (value) => (
      <Text style={{ width: 216 }} ellipsis={{ tooltip: value }}>
        {value}
      </Text>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.Status',
      defaultMessage: '状态',
    }),
    dataIndex: 'status',
    key: 'status',
    render: (value, record) => (
      <Tag color={COLOR_MAP.get(value)}>
        {' '}
        {value === 'operating' ? (
          <Text
            style={{ maxWidth: 110, color: '#d48806',fontSize:12 }}
            ellipsis={{ tooltip: `${value}/${record.statusDetail}` }}
          >
            {value}/{record.statusDetail}
          </Text>
        ) : (
          value
        )}{' '}
      </Tag>
    ),
  },
  {
    title: intl.formatMessage({
      id: 'OBDashboard.pages.Cluster.ClusterList.CreationTime',
      defaultMessage: '创建时间',
    }),
    dataIndex: 'createTime',
    key: 'createTime',
  },

  //监控还未返回
  // {
  //   title: '监控',
  //   children: [
  //     {
  //       title: 'CPU',
  //       dataIndex: 'cpuPercent',
  //       key: 'cpuPercent',
  //       render: (value) => <CanvasPie percent={Number(value)} name="cpu" />,
  //     },
  //     {
  //       title: '内存',
  //       dataIndex: 'memoryPercent',
  //       key: 'memoryPercent',
  //       render: (value) => <CanvasPie percent={Number(value)} name="memory" />,
  //     },
  //     {
  //       title: '磁盘',
  //       dataIndex: 'diskPercent',
  //       key: 'diskPercent',
  //       render: (value) => <CanvasPie percent={Number(value)} name="disk" />,
  //     },
  //   ],
  // },
];

export default function ClusterList({
  handleAddCluster,
  clusterList,
  loading,
}: ClusterListProps) {
  return (
    <Col span={24}>
      <Card>
        <div className={styles.clusterHeader}>
          <h2>
            {intl.formatMessage({
              id: 'dashboard.pages.Cluster.ClusterList.ClusterList',
              defaultMessage: '集群列表',
            })}
          </h2>
          <Button onClick={handleAddCluster} type="primary">
            {intl.formatMessage({
              id: 'OBDashboard.pages.Cluster.ClusterList.CreateACluster',
              defaultMessage: '创建集群',
            })}
          </Button>
        </div>
        <Table
          loading={loading}
          columns={columns}
          dataSource={clusterList}
          scroll={{ x: 1200 }}
          pagination={{ simple: true }}
          rowKey="name"
          bordered
          sticky
        />
      </Card>
    </Col>
  );
}
