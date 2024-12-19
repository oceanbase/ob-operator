import { obcluster } from '@/api';
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
  Tag,
  Tooltip,
  message,
} from 'antd';
import { isEmpty } from 'lodash';
import { useEffect, useRef, useState } from 'react';
import BasicInfo from './BasicInfo';
import NFSInfoModal from './NFSInfoModal';
import ParametersModal from './ParametersModal';
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
  const [parametersData, setParametersData] = useState([]);

  const { setFieldsValue, validateFields } = form;

  const {
    data: listOBClusterParameters,
    loading,
    refresh,
  } = useRequest(obcluster.listOBClusterParameters, {
    defaultParams: [ns, name],
    onSuccess: (res) => {
      const newData = getNewData(res?.data);
      setParametersData(newData);
    },
  });

  const getNewData = (data) => {
    const obt = data
      ?.map((element) => {
        // obcluster 的 parameters 里面加了个 specValue 的字段，
        // 如果 specValue 不等于 value，状态写 "不匹配" (黄色tag)，如果两个值相等，写"已匹配"(绿色tag)
        const findSpec = parameters?.find(
          (item: any) => element.value === item.specValue,
        );
        if (!isEmpty(findSpec)) {
          return { ...element, accordance: true };
        } else if (isEmpty(findSpec)) {
          return { ...element, accordance: false };
        }
      })
      ?.map((element: any) => {
        // 在 obcluster 的 parameters  里面的就是托管给 operator
        const findName = parameters?.find(
          (item: any) => element.name === item.name,
        );
        if (!isEmpty(findName)) {
          return { ...element, controlParameter: true };
        } else if (isEmpty(findName)) {
          return { ...element, controlParameter: false };
        }
      });
    return obt;
  };

  const {
    data: clusterDetail,
    run: getClusterDetail,
    refresh: clusterDetailRefresh,
    loading: clusterDetailLoading,
  } = useRequest(getClusterDetailReq, {
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
  });

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

  // 不为空即为绑定了NFS
  const removeNFS = !!clusterDetail?.info?.backupVolume;

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
          disabled={
            clusterDetail?.status !== 'running' ||
            !clusterDetail?.supportStaticIP
          }
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
      value: floorToTwoDecimalPlaces(storage?.dataStorage?.size / (1 << 30)),
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
      value: floorToTwoDecimalPlaces(storage?.redoLogStorage?.size / (1 << 30)),
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
      value: floorToTwoDecimalPlaces(storage?.sysLogStorage?.size / (1 << 30)),
    },
  ];

  const controlParameters = [
    {
      label: '已托管',
      value: true,
    },
    {
      label: '未托管',
      value: false,
    },
  ];

  const accordanceList = [
    {
      label: <Tag color={'green'}>{'已匹配'}</Tag>,
      value: true,
    },
    {
      label: <Tag color={'gold'}>{'不匹配'}</Tag>,
      value: false,
    },
  ];

  useEffect(() => {
    getClusterDetail({ ns: ns!, name: name! });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  const columns = [
    {
      title: '参数名',
      dataIndex: 'name',
    },
    {
      title: '参数值',
      dataIndex: 'value',
    },
    {
      title: '参数说明',
      dataIndex: 'info',
    },
    {
      title: '托管 operator',
      dataIndex: 'controlParameter',
      filters: controlParameters.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.controlParameter === value;
      },
      render: (text: boolean) => {
        return <span>{text ? '是' : '否'}</span>;
      },
    },
    {
      title: <IconTip tip="只有托管 operator 的参数才有状态" content="状态" />,
      dataIndex: 'accordance',
      width: 100,
      render: (text: boolean) => {
        const tagColor = text ? 'green' : 'gold';
        const tagContent = text ? '已匹配' : '不匹配';

        return <Tag color={tagColor}>{tagContent}</Tag>;
      },
    },
    {
      title: '操作',
      dataIndex: 'controlParameter',
      render: (text, record) => {
        return (
          <Space size={1}>
            <Button
              type="link"
              onClick={() => {
                setIsDrawerOpen(true);
                setParametersRecord(record);
              }}
            >
              编辑
            </Button>
            {text && (
              <Button type="link" onClick={() => {}}>
                解除托管
              </Button>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <PageContainer header={header()} loading={clusterDetailLoading}>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <BasicInfo {...(clusterDetail?.info as API.ClusterInfo)} />
          </Col>
        )}
        <Col span={24}>
          <Card
            title={<h2 style={{ marginBottom: 0 }}>节点资源配置</h2>}
            extra={
              <Button
                onClick={() => setResourceDrawerOpen(true)}
                type="primary"
                disabled={!clusterDetail?.supportStaticIP}
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
            <Space style={{ marginBottom: '16px' }}>
              PVC 独立生命周期
              <Tooltip title={'只能在创建时指定，不支持修改'}>
                <Checkbox disabled />
              </Tooltip>
            </Space>
            <Descriptions>
              {resourceinit?.map((resource) => (
                <Descriptions.Item label={resource.key}>
                  {resource.label === 'size'
                    ? `${resource.value}Gi`
                    : resource.value}
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
                  <Form.Item label="参数名" name={'name'}>
                    <Input placeholder="请输入" allowClear />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="托管状态" name={'controlParameter'}>
                    <Select options={controlParameters} allowClear={true} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="状态" name={'accordance'}>
                    <Select options={accordanceList} allowClear={true} />
                  </Form.Item>
                </Col>
                <Col>
                  <Space size="middle">
                    <Button
                      type="primary"
                      onClick={() => {
                        validateFields().then((values) => {
                          const { name, controlParameter, accordance } = values;
                          const newParametersData = getNewData(
                            listOBClusterParameters?.data,
                          );

                          if (name) {
                            setParametersData(
                              newParametersData?.filter((item) =>
                                item.name?.includes(name),
                              ),
                            );
                          }
                          if (controlParameter) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.controlParameter === controlParameter,
                              ),
                            );
                          }
                          if (accordance) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) => item.accordance === accordance,
                              ),
                            );
                          }
                          if (!!name && !!controlParameter) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.name?.includes(name) &&
                                  item.controlParameter === controlParameter,
                              ),
                            );
                          }
                          if (!!name && !!accordance) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.name?.includes(name) &&
                                  item.accordance === accordance,
                              ),
                            );
                          }
                          if (!!name && !!controlParameter && !!accordance) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.name?.includes(name) &&
                                  item.controlParameter === controlParameter &&
                                  item.accordance === accordance,
                              ),
                            );
                          }
                        });
                      }}
                    >
                      查询
                    </Button>
                    <Button
                      onClick={() => {
                        setFieldsValue({
                          name: '',
                          controlParameter: '',
                          accordance: '',
                        });
                        refresh();
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
              loading={loading}
              dataSource={parametersData}
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
          <ServerTable
            clusterDetail={clusterDetail}
            clusterDetailRefresh={() => {
              clusterDetailRefresh();
            }}
          />
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

      <ParametersModal
        visible={isDrawerOpen}
        onCancel={() => setIsDrawerOpen(false)}
        onSuccess={() => {
          setIsDrawerOpen(false);
          clusterDetailRefresh();
        }}
        initialValues={parametersRecord}
        {...(clusterDetail?.info as API.ClusterInfo)}
      />

      <ResourceDrawer
        visible={resourceDrawerOpen}
        onCancel={() => setResourceDrawerOpen(false)}
        onSuccess={() => {
          setResourceDrawerOpen(false);
          clusterDetailRefresh();
        }}
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
        onSuccess={() => {
          removeNFS ? setRemoveNFSModal(false) : setMountNFSModal(false);
          clusterDetailRefresh();
        }}
        {...(clusterDetail?.info as API.ClusterInfo)}
      />
    </PageContainer>
  );
};

export default ClusterOverview;
