import { history, useAccess, useModel, useParams } from '@umijs/max';
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
import { useEffect, useRef, useState } from 'react';

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
import { useRequest } from 'ahooks';
import { isEmpty } from 'lodash';
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

  const getNewData = (data) => {
    const obt = data?.map((element: any) => {
      // 在 obcluster 的 parameters  里面的就是托管给 operator
      const findName = parameters?.find(
        (item: any) => element.name === item.name,
      );

      if (!isEmpty(findName)) {
        return {
          ...element,
          controlParameter: true,
          accordance: findName?.value === findName?.specValue,
        };
      } else if (isEmpty(findName)) {
        return { ...element, controlParameter: false, accordance: undefined };
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
          refresh();
        }, REFRESH_CLUSTER_TIME);
      } else if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    },
  });

  const {
    data: listOBClusterParameters,
    loading,
    refresh,
  } = useRequest(obcluster.listOBClusterParameters, {
    defaultParams: [ns, name],
    refreshDeps: [clusterDetail?.status],
    onSuccess: (res) => {
      const newData = getNewData(res?.data);
      setParametersData(newData);
    },
  });

  const { runAsync: patchOBCluster, loading: patchOBClusterloading } =
    useRequest(obcluster.patchOBCluster, {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.FF85D01F',
              defaultMessage: '解除托管已成功',
            }),
          );
          setFieldsValue({
            name: undefined,
            controlParameter: undefined,
            accordance: undefined,
          });
          refresh();
          clusterDetailRefresh();
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

  const {
    parameters,
    storage,
    resource,
    deletionProtection,
    backupVolume,
    pvcIndependent,
  } = clusterDetail?.info || {};

  // 不为空即为绑定了NFS
  const removeNFS = !!backupVolume;

  const menuChange = ({ key }) => {
    if (key === 'AddZone') {
      return handleAddZone();
    } else if (key === 'Upgrade') {
      return handleUpgrade();
    } else if (key === 'delete') {
      return showDeleteConfirm({
        onOk: handleDelete,
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.AreYouSureYouWant',
          defaultMessage: '你确定要删除该集群吗？',
        }),
      });
    } else if (key === 'nfs') {
      if (removeNFS) {
        setRemoveNFSModal(true);
      } else {
        setMountNFSModal(true);
      }
    }
  };

  const items: MenuProps['items'] = [
    {
      key: 'AddZone',
      label: intl.formatMessage({
        id: 'dashboard.Detail.Overview.AddZone',
        defaultMessage: '新增Zone',
      }),

      disabled: !isEmpty(clusterDetail) && clusterDetail?.status !== 'running',
    },
    {
      key: 'Upgrade',
      label: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.Upgrade',
        defaultMessage: '升级',
      }),
      disabled: !isEmpty(clusterDetail) && clusterDetail?.status !== 'running',
    },
    {
      key: 'delete',
      label: intl.formatMessage({
        id: 'OBDashboard.Detail.Overview.Delete',
        defaultMessage: '删除',
      }),
      danger: true,
      disabled:
        !isEmpty(clusterDetail) &&
        (clusterDetail?.status === 'deleting' || deletionProtection),
    },
    {
      key: 'nfs',
      label: (
        <span>
          {removeNFS
            ? intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.C47B9DA4',
                defaultMessage: '移除 NFS 资源',
              })
            : intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.6B97ABB6',
                defaultMessage: '挂载 NFS 资源',
              })}
        </span>
      ),
      disabled:
        isEmpty(clusterDetail) ||
        clusterDetail?.status !== 'running' ||
        !clusterDetail?.supportStaticIP,
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
            <Dropdown menu={{ items, onClick: menuChange }}>
              <Button>
                <Space>
                  {intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.A0A43F50',
                    defaultMessage: '集群管理',
                  })}
                  <DownOutlined />
                </Space>
              </Button>
            </Dropdown>,
          ]
        : [],
    };
  };

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
      label: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.403B7E1C',
        defaultMessage: '已托管',
      }),
      value: true,
    },
    {
      label: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.46B66B3E',
        defaultMessage: '未托管',
      }),
      value: false,
    },
  ];

  const accordanceList = [
    {
      label: (
        <Tag color={'green'}>
          {intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.D5CCD27D',
            defaultMessage: '已匹配',
          })}
        </Tag>
      ),

      value: true,
    },
    {
      label: (
        <Tag color={'gold'}>
          {intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.DF83C06D',
            defaultMessage: '不匹配',
          })}
        </Tag>
      ),

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
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.E5342F26',
        defaultMessage: '参数名',
      }),
      dataIndex: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.FA0D096B',
        defaultMessage: '参数值',
      }),
      dataIndex: 'value',
      render: (text: string, record) => {
        const content =
          parameters?.find((item) => item.name === record.name)?.value || text;

        return <span>{content}</span>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.93A9D19D',
        defaultMessage: '参数说明',
      }),
      dataIndex: 'info',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.4FCF90AF',
        defaultMessage: '托管 operator',
      }),
      width: 140,
      dataIndex: 'controlParameter',
      filters: controlParameters.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.controlParameter === value;
      },
      render: (text: boolean) => {
        return (
          <span>
            {text
              ? intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.319FA0DB',
                  defaultMessage: '是',
                })
              : intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.5DD958C7',
                  defaultMessage: '否',
                })}
          </span>
        );
      },
    },
    {
      title: (
        <IconTip
          tip={intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.0B4A3E74',
            defaultMessage: '只有托管 operator 的参数才有状态',
          })}
          content={intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.6AD01A82',
            defaultMessage: '状态',
          })}
        />
      ),

      dataIndex: 'accordance',
      width: 100,
      render: (text) => {
        const tagColor = text ? 'green' : 'gold';
        const tagContent = text
          ? intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.9A3A4407',
              defaultMessage: '已匹配',
            })
          : intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.D6588C55',
              defaultMessage: '不匹配',
            });

        return text === undefined ? (
          '/'
        ) : (
          <Tag color={tagColor}>{tagContent}</Tag>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.1B9EA477',
        defaultMessage: '操作',
      }),
      dataIndex: 'controlParameter',
      align: 'center',
      render: (text, record) => {
        const disableUnescrow = [
          'memory_limit',
          'datafile_maxsize',
          'datafile_next',
          'enable_syslog_recycle',
          'max_syslog_file_count',
        ];

        const valueContent =
          parameters?.find((item) => item.name === record.name)?.value ||
          record?.value;

        return (
          <Space size={1}>
            <Button
              type="link"
              onClick={() => {
                setIsDrawerOpen(true);
                setParametersRecord({
                  ...record,
                  value: valueContent,
                });
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.F5A088FB',
                defaultMessage: '编辑',
              })}
            </Button>
            {text && (
              <Button
                type="link"
                disabled={disableUnescrow.some((item) => item === record.name)}
                loading={patchOBClusterloading}
                onClick={() => {
                  patchOBCluster(ns, name, {
                    deletedParameters: [record.name],
                  });
                }}
              >
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.5FACF7C0',
                  defaultMessage: '解除托管',
                })}
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
            <BasicInfo
              {...(clusterDetail?.info as API.ClusterInfo)}
              clusterDetailRefresh={() => {
                clusterDetailRefresh();
              }}
            />
          </Col>
        )}
        <Col span={24}>
          <Card
            title={
              <h2 style={{ marginBottom: 0 }}>
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.43C45255',
                  defaultMessage: '节点资源配置',
                })}
              </h2>
            }
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
            <Descriptions
              title={intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.C5E0380E',
                defaultMessage: '计算资源',
              })}
            >
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
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.05F3B008',
                defaultMessage: '存储资源',
              })}
            </div>
            <Space style={{ marginBottom: '16px' }}>
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.F4D80804',
                defaultMessage: 'PVC 独立生命周期',
              })}

              <Tooltip
                title={intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.21732CE3',
                  defaultMessage: '只能在创建时指定，不支持修改',
                })}
              >
                <Checkbox defaultChecked={pvcIndependent} disabled />
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
          <Card
            title={
              <h2 style={{ marginBottom: 0 }}>
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.BFE7CA02',
                  defaultMessage: '集群参数设置',
                })}
              </h2>
            }
          >
            <Form form={form}>
              <Row gutter={[24, 16]}>
                <Col span={6}>
                  <Form.Item
                    label={intl.formatMessage({
                      id: 'src.pages.Cluster.Detail.Overview.BF489BCE',
                      defaultMessage: '参数名',
                    })}
                    name={'name'}
                  >
                    <Input
                      placeholder={intl.formatMessage({
                        id: 'src.pages.Cluster.Detail.Overview.E5E4E6B5',
                        defaultMessage: '请输入',
                      })}
                      allowClear
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item
                    label={intl.formatMessage({
                      id: 'src.pages.Cluster.Detail.Overview.4F7F81B0',
                      defaultMessage: '托管状态',
                    })}
                    name={'controlParameter'}
                  >
                    <Select
                      options={controlParameters}
                      allowClear={true}
                      placeholder={intl.formatMessage({
                        id: 'src.pages.Cluster.Detail.Overview.A89209B9',
                        defaultMessage: '请选择',
                      })}
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item
                    label={intl.formatMessage({
                      id: 'src.pages.Cluster.Detail.Overview.2873685C',
                      defaultMessage: '状态',
                    })}
                    name={'accordance'}
                  >
                    <Select
                      options={accordanceList}
                      allowClear={true}
                      placeholder={intl.formatMessage({
                        id: 'src.pages.Cluster.Detail.Overview.5C55A022',
                        defaultMessage: '请选择',
                      })}
                    />
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

                          if (name !== undefined) {
                            setParametersData(
                              newParametersData?.filter((item) =>
                                item.name?.includes(name.trim()),
                              ),
                            );
                          }
                          if (controlParameter !== undefined) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.controlParameter === controlParameter,
                              ),
                            );
                          }
                          // 已托管的参数才有托管状态
                          const controlParameterContent =
                            newParametersData?.filter(
                              (item) => item.controlParameter === true,
                            );
                          if (accordance !== undefined) {
                            setParametersData(
                              controlParameterContent?.filter(
                                (item) => item.accordance === accordance,
                              ),
                            );
                          }
                          if (
                            name !== undefined &&
                            controlParameter !== undefined
                          ) {
                            setParametersData(
                              newParametersData?.filter(
                                (item) =>
                                  item.name?.includes(name.trim()) &&
                                  item.controlParameter === controlParameter,
                              ),
                            );
                          }
                          if (name !== undefined && accordance !== undefined) {
                            setParametersData(
                              controlParameterContent?.filter(
                                (item) =>
                                  item.name?.includes(name.trim()) &&
                                  item.accordance === accordance,
                              ),
                            );
                          }
                          if (
                            controlParameter !== undefined &&
                            accordance !== undefined
                          ) {
                            if (controlParameter === true) {
                              setParametersData(
                                controlParameterContent?.filter(
                                  (item) => item.accordance === accordance,
                                ),
                              );
                            } else {
                              setParametersData([]);
                            }
                          }

                          if (
                            name !== undefined &&
                            controlParameter !== undefined &&
                            accordance !== undefined
                          ) {
                            if (controlParameter === true) {
                              setParametersData(
                                controlParameterContent?.filter(
                                  (item) =>
                                    item.name?.includes(name.trim()) &&
                                    item.accordance === accordance,
                                ),
                              );
                            } else {
                              // 未托管的，没有状态，因此为空
                              setParametersData([]);
                            }
                          }
                        });
                      }}
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Cluster.Detail.Overview.E3D520F9',
                        defaultMessage: '查询',
                      })}
                    </Button>
                    <Button
                      onClick={() => {
                        setFieldsValue({
                          name: undefined,
                          controlParameter: undefined,
                          accordance: undefined,
                        });
                        refresh();
                      }}
                    >
                      {intl.formatMessage({
                        id: 'src.pages.Cluster.Detail.Overview.96EEA6EE',
                        defaultMessage: '重置',
                      })}
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
          // 编辑成功后，清空搜索条件，刷新参数列表
          setFieldsValue({
            name: undefined,
            controlParameter: undefined,
            accordance: undefined,
          });
          refresh();
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
        title={
          removeNFS
            ? intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.24BBEBC2',
                defaultMessage: '移除 NFS 备份卷',
              })
            : intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.44A8C98B',
                defaultMessage: '挂载 NFS 备份卷',
              })
        }
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
