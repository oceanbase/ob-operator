import EventsTable from '@/components/EventsTable';
import IconTip from '@/components/IconTip';
import OperateModal from '@/components/customModal/OperateModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { REFRESH_CLUSTER_TIME } from '@/constants';
import { getClusterDetailReq } from '@/services';
import { deleteClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import { floorToTwoDecimalPlaces } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { DownOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Checkbox } from '@oceanbase/design';
import { history, useAccess, useModel, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Dropdown,
  Form,
  Input,
  MenuProps,
  Row,
  Select,
  Space,
  Table,
  Tooltip,
  message,
} from 'antd';
import { useEffect, useRef, useState } from 'react';
import BasicInfo from './BasicInfo';
import NFSInfoModal from './NFSInfoModal';
import ParametersDrawer from './ParametersDrawer';
import ResourceDrawer from './ResourceDrawer';
import ServerTable from './ServerTable';
import ZoneTable from './ZoneTable';

const ClusterOverview: React.FC = () => {
  const { setChooseClusterName } = useModel('global');
  const access = useAccess();
  const [form] = Form.useForm();
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  const [isDrawerOpen, setIsDrawerOpen] = useState<boolean>(false);
  const [parametersRecord, setParametersRecord] = useState({});
  const [resourceDrawerOpen, setResourceDrawerOpen] = useState<boolean>(false);
  const { ns, name } = useParams();
  const chooseZoneName = useRef<string>('');
  const timerRef = useRef<NodeJS.Timeout>();
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  const [mountNFSModal, setMountNFSModal] = useState<boolean>(false);
  const [removeNFSModal, setRemoveNFSModal] = useState<boolean>(false);
  const modalType = useRef<API.ModalType>('addZone');

  const { setFieldsValue } = form;

  const { data: clusterDetail, run: getClusterDetail } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onSuccess: (data) => {
        setChooseClusterName(data.info.clusterName);
        if (data.status === 'operating') {
          timerRef.current = setTimeout(() => {
            getClusterDetail({ ns: ns!, name: name! });
          }, REFRESH_CLUSTER_TIME);
        } else if (timerRef.current) {
          clearTimeout(timerRef.current);
        }
      },
    },
  );

  const handleDelete = async () => {
    const res = await deleteClusterReportWrap({ ns: ns!, name: name! });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.DeletedSuccessfully',
          defaultMessage: '删除成功',
        }),
      );
      history.replace('/cluster');
    }
  };

  const operateSuccess = () => {
    setTimeout(() => {
      getClusterDetail({ ns: ns!, name: name! });
    }, 1000);
  };
  const handleAddZone = () => {
    modalType.current = 'addZone';
    setOperateModalVisible(true);
  };
  const handleUpgrade = () => {
    modalType.current = 'upgradeCluster';
    setOperateModalVisible(true);
  };

  // TODO 判断当前 nfs 状态
  const removeNFS = false;

  const items: MenuProps['items'] = [
    {
      key: '1',
      label: (
        <Button
          onClick={handleAddZone}
          disabled={
            clusterDetail?.status === 'operating' ||
            clusterDetail?.status === 'failed'
          }
          type="text"
        >
          {intl.formatMessage({
            id: 'dashboard.Detail.Overview.AddZone',
            defaultMessage: '新增Zone',
          })}
        </Button>
      ),
    },
    {
      key: '2',
      label: (
        <Button
          type="text"
          disabled={
            clusterDetail?.status === 'operating' ||
            clusterDetail?.status === 'failed'
          }
          onClick={handleUpgrade}
        >
          {intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.Upgrade',
            defaultMessage: '升级',
          })}
        </Button>
      ),
    },
    {
      key: '3',
      label: (
        <Button
          type="text"
          disabled={clusterDetail?.status === 'operating'}
          onClick={() =>
            showDeleteConfirm({
              onOk: handleDelete,
              title: intl.formatMessage({
                id: 'OBDashboard.Detail.Overview.AreYouSureYouWant',
                defaultMessage: '你确定要删除该集群吗？',
              }),
            })
          }
          danger
        >
          {intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.Delete',
            defaultMessage: '删除',
          })}
        </Button>
      ),
    },
    {
      key: '4',
      label: (
        <Button
          type="text"
          onClick={() => {
            if (removeNFS) {
              setRemoveNFSModal(true);
            } else {
              setMountNFSModal(true);
            }
          }}
        >
          {removeNFS ? '移除 NFS 资源' : '挂载 NFS 资源'}
        </Button>
      ),
    },
  ];

  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ClusterOverview',
        defaultMessage: '集群概览',
      }),
      extra: access.obclusterwrite
        ? [
            <Dropdown menu={{ items }} placement="bottomRight">
              <Button>
                集群管理
                <DownOutlined />
              </Button>
            </Dropdown>,
          ]
        : [],
    };
  };

  useEffect(() => {
    getClusterDetail({ ns: ns!, name: name! });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  const { parameters, storage, resource } = clusterDetail?.info || {};

  const resourceinit = [
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.DatafileStorageClass',
        defaultMessage: 'Datafile 存储类',
      }),
      type: 'data',
      label: 'storageClass',
      value: storage?.dataStorage?.storageClass,
    },
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.DatafileStorageSize',
        defaultMessage: 'Datafile 存储大小',
      }),
      type: 'data',
      label: 'size',
      value:
        floorToTwoDecimalPlaces(storage?.dataStorage?.size / (1 << 30)) + 'Gi',
    },
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.RedologStorageClass',
        defaultMessage: 'RedoLog 存储类',
      }),
      type: 'redoLog',
      label: 'storageClass',
      value: storage?.redoLogStorage?.storageClass,
    },
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.RedologSize',
        defaultMessage: 'RedoLog 大小',
      }),
      type: 'redoLog',
      label: 'size',
      value:
        floorToTwoDecimalPlaces(storage?.redoLogStorage?.size / (1 << 30)) +
        'Gi',
    },
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.SystemLogStorageClass',
        defaultMessage: '系统日志存储类',
      }),
      type: 'log',
      label: 'storageClass',
      value: storage?.sysLogStorage?.storageClass,
    },
    {
      key: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.BasicInfo.SystemLogStorageSize',
        defaultMessage: '系统日志存储大小',
      }),
      type: 'log',
      label: 'size',
      value:
        floorToTwoDecimalPlaces(storage?.sysLogStorage?.size / (1 << 30)) +
        'Gi',
    },
  ];

  const columns = [
    {
      title: '参数名',
      dataIndex: 'key',
    },
    {
      title: '参数值',
      dataIndex: 'value',
    },
    {
      title: '参数说明',
      dataIndex: 'name',
    },
    {
      title: '托管 operator',
      dataIndex: 'name',
      filters: [
        {
          label: '是',
          value: 'yes',
        },
        {
          label: '否',
          value: 'no',
        },
      ].map(({ label, value }) => ({
        text: label,
        value,
      })),
    },
    {
      title: <IconTip tip="只有托管 operator 的参数才有状态" content="状态" />,
      dataIndex: 'name',
      render: (text) => {
        return <span>{text || '-'}</span>;
      },
    },
    {
      title: '操作',
      dataIndex: 'operation',
      render: (text, record) => {
        return (
          <a
            onClick={() => {
              setIsDrawerOpen(true);
              setParametersRecord(record);
            }}
          >
            编辑
          </a>
        );
      },
    },
  ];

  console.log('cl', clusterDetail);
  return (
    <PageContainer header={header()}>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <BasicInfo {...(clusterDetail?.info as API.ClusterInfo)} />
          </Col>
        )}
        {/* <Col span={24}>
          <Card
            title={
              <h2 style={{ marginBottom: 0 }}>
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.9F880AEF',
                  defaultMessage: '参数设置',
                })}
              </h2>
            }
            extra={
              <Button onClick={() => setIsDrawerOpen(true)} type="primary">
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.533B34EA',
                  defaultMessage: '编辑',
                })}
              </Button>
            }
          >
            {parameters?.length > 0 ? (
              <Descriptions
                title={intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.7F3B8DF8',
                  defaultMessage: '集群参数',
                })}
              >
                {parameters.map((parameter, index) => (
                  <Descriptions.Item label={parameter.key} key={index}>
                    {parameter.value}
                  </Descriptions.Item>
                ))}
              </Descriptions>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Card>
        </Col> */}
        <Col span={24}>
          <Card
            title={<h2 style={{ marginBottom: 0 }}>节点资源配置</h2>}
            extra={
              <Button
                onClick={() => setResourceDrawerOpen(true)}
                type="primary"
              >
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.533B34EA',
                  defaultMessage: '编辑',
                })}
              </Button>
            }
          >
            <Descriptions title={'计算资源'}>
              <Descriptions.Item label={'CPU'}>
                {resource?.cpu}
              </Descriptions.Item>
              <Descriptions.Item label={'Memory'}>
                {floorToTwoDecimalPlaces(resource?.memory / (1 << 30)) + 'Gi'}
              </Descriptions.Item>
            </Descriptions>
            <div
              style={{
                color: '#132039',
                fontWeight: 600,
                fontSize: '16px',
                marginBottom: '16px',
              }}
            >
              存储资源
            </div>
            {/* TODO  */}
            <Space style={{ marginBottom: '16px' }}>
              PVC 独立生命周期
              <Tooltip title={'只能在创建时指定，不支持修改'}>
                <Checkbox disabled />
              </Tooltip>
            </Space>
            <Descriptions>
              {resourceinit?.map((resource) => (
                <Descriptions.Item label={resource.key}>
                  {resource.value}
                </Descriptions.Item>
              ))}
            </Descriptions>
          </Card>
        </Col>
        <Col span={24}>
          <Card title={<h2 style={{ marginBottom: 0 }}>集群参数设置</h2>}>
            <Form form={form}>
              <Row gutter={[24, 16]}>
                <Col span={6}>
                  <Form.Item name="" label="参数名">
                    <Input placeholder="请输入" />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="" label="托管状态">
                    <Select options={[{ value: 'lucy', label: 'Lucy' }]} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="" label="状态">
                    <Select options={[{ value: 'lucy', label: 'Lucy' }]} />
                  </Form.Item>
                </Col>
                <Col>
                  <Space size="middle">
                    <Button type="primary">查询</Button>
                    <Button
                      onClick={() => {
                        setFieldsValue({
                          shieldStatus: '',
                        });
                      }}
                    >
                      重置
                    </Button>
                  </Space>
                </Col>
              </Row>
            </Form>

            <Table
              rowKey="name"
              pagination={{ simple: true }}
              columns={columns}
              dataSource={parameters}
            />
          </Card>
        </Col>
        {clusterDetail && (
          <ZoneTable
            clusterStatus={clusterDetail.status}
            zones={clusterDetail.zones as API.Zone[]}
            chooseZoneRef={chooseZoneName}
            setVisible={setOperateModalVisible}
            typeRef={modalType}
            setChooseServerNum={setChooseServerNum}
          />
        )}
        {clusterDetail && (
          <ServerTable servers={clusterDetail.servers as API.Server[]} />
        )}
        <Col span={24}>
          <EventsTable
            objectType="OBCLUSTER"
            name={clusterDetail?.info?.name}
          />
        </Col>
      </Row>
      <OperateModal
        type={modalType.current}
        visible={operateModalVisible}
        setVisible={setOperateModalVisible}
        successCallback={operateSuccess}
        params={{
          zoneName: chooseZoneName.current,
          defaultValue: chooseServerNum,
        }}
      />

      <ParametersDrawer
        visible={isDrawerOpen}
        onCancel={() => setIsDrawerOpen(false)}
        onSuccess={() => setIsDrawerOpen(false)}
        initialValues={[parametersRecord]}
        {...(clusterDetail?.info as API.ClusterInfo)}
      />

      <ResourceDrawer
        visible={resourceDrawerOpen}
        onCancel={() => setResourceDrawerOpen(false)}
        onSuccess={() => setResourceDrawerOpen(false)}
        initialValues={resourceinit}
        {...(clusterDetail?.info as API.ClusterInfo)}
      />

      <NFSInfoModal
        removeNFS={removeNFS}
        title={removeNFS ? '移除 NFS 备份卷' : '挂载 NFS 备份卷'}
        visible={removeNFS ? removeNFSModal : mountNFSModal}
        onCancel={() =>
          removeNFS ? setRemoveNFSModal(false) : setMountNFSModal(false)
        }
        onSuccess={() =>
          removeNFS ? setRemoveNFSModal(false) : setMountNFSModal(false)
        }
        {...(clusterDetail?.info as API.ClusterInfo)}
      />
    </PageContainer>
  );
};

export default ClusterOverview;
