import { intl } from '@/utils/intl';
import { ProCard } from '@ant-design/pro-components';
import { Col, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';

import { COLOR_MAP } from '@/constants';

const getServerColums = () => {
  const serverColums: ColumnsType<API.Server> = [
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.ServerName',
        defaultMessage: 'Server名',
      }),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Namespace',
        defaultMessage: '命名空间',
      }),
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Address',
        defaultMessage: '地址',
      }),
      dataIndex: 'address',
      key: 'address',
    },
    // {
    //   title: '资源水位',
    //   dataIndex: 'metrics',
    //   key: 'metrics',
    //   width: 300,
    //   render: (value) => {
    //     let resources: { percent: string; text: string }[] = [];
    //     Object.keys(value).forEach((val) => {
    //       let text: string = '';
    //       if (val === 'cpuPercent') text = 'CPU';
    //       if (val === 'memoryPercent') text = '内存';
    //       if (val === 'diskPercent') text = '磁盘';
    //       resources.push({
    //         percent: value[val],
    //         text,
    //       });
    //     });
    //     return (
    //       <div>
    //         {resources.map((resource, idx: number) => (
    //           <div key={idx} className={styles.resourceContainer}>
    //             <span className={styles.resourceText}>{resource.text}</span>{' '}
    //             <Progress
    //               className={styles.resourceContent}
    //               strokeLinecap="butt"
    //               percent={Number(resource.percent)}
    //             />
    //           </div>
    //         ))}
    //       </div>
    //     );
    //   },
    // },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.ServerTable.Status',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      key: 'status',
      render: (value) => <Tag color={COLOR_MAP.get(value)}>{value} </Tag>,
    },
    // 目前不支持删除指定server
    // {
    //   title: '操作',
    //   key: 'action',
    //   render: (_, record) => (
    //     <a
    //       onClick={() => {
    //         showDeleteConfirm({
    //           onOk: () => remove(record.name),
    //           title: '你确定要删除该server吗？',
    //         });
    //       }}
    //     >
    //       删除
    //     </a>
    //   ),
    // },
  ];
  return serverColums;
};

export default function ServerTable({ servers }: { servers: API.Server[] }) {
  return (
    <Col span={24}>
      <ProCard>
        <Table
          columns={getServerColums()}
          rowKey="name"
          dataSource={servers}
          pagination={{simple:true}}
          sticky
        />
      </ProCard>
    </Col>
  );
}
