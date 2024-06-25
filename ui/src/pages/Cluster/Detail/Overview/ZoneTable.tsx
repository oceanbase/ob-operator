import { intl } from '@/utils/intl';
import { Button, Card, Col, Table, Tag, message } from 'antd';
import type { ColumnType } from 'antd/es/table';

import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { COLOR_MAP } from '@/constants';
import { deleteObzoneReportWrap } from '@/services/reportRequest/clusterReportReq';
import { useParams } from '@umijs/max';
interface ZoneTableProps {
  zones: API.Zone[];
  setVisible: React.Dispatch<React.SetStateAction<boolean>>;
  chooseZoneRef: React.MutableRefObject<string>;
  typeRef: React.MutableRefObject<API.ModalType>;
  setChooseServerNum: React.Dispatch<React.SetStateAction<number>>;
  clusterStatus: 'running' | 'failed' | 'operating';
}

export default function ZoneTable({
  zones,
  setVisible,
  chooseZoneRef,
  typeRef,
  setChooseServerNum,
  clusterStatus,
}: ZoneTableProps) {
  const { ns, name } = useParams();
  const getZoneColumns = (remove, clickScale) => {
    const columns: ColumnType<API.Zone> = [
      {
        title: intl.formatMessage({
          id: 'Dashboard.Detail.Overview.ZoneTable.ZoneResourceName',
          defaultMessage: 'Zone 资源名',
        }),
        dataIndex: 'name',
        key: 'name',
        width: 190,
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Namespace',
          defaultMessage: '命名空间',
        }),
        dataIndex: 'namespace',
        key: 'namespace',
      },
      {
        title: intl.formatMessage({
          id: 'Dashboard.Detail.Overview.ZoneTable.ZoneName',
          defaultMessage: 'Zone 名',
        }),
        dataIndex: 'zone',
        key: 'zone',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.NumberOfMachines',
          defaultMessage: '机器数量',
        }),
        dataIndex: 'replicas',
        key: 'replicas',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.RootServiceIp',
          defaultMessage: '根服务IP',
        }),
        dataIndex: 'rootService',
        key: 'rootService',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Status',
          defaultMessage: '状态',
        }),
        dataIndex: 'status',
        key: 'status',
        render: (value) => <Tag color={COLOR_MAP.get(value)}>{value} </Tag>,
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Operation',
          defaultMessage: '操作',
        }),
        key: 'action',
        render: (_, record) => {
          return (
            <>
              <Button
                style={{ paddingLeft: 0 }}
                onClick={() => {
                  clickScale(record.zone);
                  setChooseServerNum(record.replicas);
                }}
                disabled={clusterStatus !== 'running'}
                type="link"
              >
                {intl.formatMessage({
                  id: 'OBDashboard.Detail.Overview.ZoneTable.Scale',
                  defaultMessage: '扩缩容',
                })}
              </Button>
              <Button
                style={
                  clusterStatus === 'running' && zones.length > 2
                    ? { color: '#ff4b4b' }
                    : {}
                }
                onClick={() => {
                  showDeleteConfirm({
                    onOk: () => remove(record.zone),
                    title: intl.formatMessage({
                      id: 'OBDashboard.Detail.Overview.ZoneTable.AreYouSureYouWant',
                      defaultMessage: '你确定要删除该zone吗？',
                    }),
                  });
                }}
                disabled={clusterStatus !== 'running' || zones.length <= 2}
                type="link"
              >
                {intl.formatMessage({
                  id: 'OBDashboard.Detail.Overview.ZoneTable.Delete',
                  defaultMessage: '删除',
                })}
              </Button>
            </>
          );
        },
      },
    ];

    return columns;
  };
  const clickScale = (zoneName: string) => {
    chooseZoneRef.current = zoneName;
    typeRef.current = 'scaleServer';
    setVisible(true);
  };
  const handleDelete = async (zoneName: string) => {
    const res = await deleteObzoneReportWrap({
      ns: ns!,
      name: name!,
      zoneName,
    });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.OperationSucceeded',
          defaultMessage: '操作成功！',
        }),
      );
    }
  };
  return (
    <Col span={24}>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.ZoneTable.ZoneList',
              defaultMessage: 'Zone 列表',
            })}
          </h2>
        }
      >
        <Table
          rowKey="name"
          pagination={{ simple: true }}
          columns={getZoneColumns(handleDelete, clickScale)}
          dataSource={zones}
        />
      </Card>
    </Col>
  );
}
