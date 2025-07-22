import { history, useAccess, useModel, useParams } from '@umijs/max';
import {
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
  Dropdown,
  Form,
  MenuProps,
  Modal,
  Row,
  Space,
  Tooltip,
  message,
} from 'antd';

import { job, obcluster } from '@/api';
import DownloadModal from '@/components/DownloadModal';
import EventsTable from '@/components/EventsTable';
import OperateModal from '@/components/customModal/OperateModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { REFRESH_CLUSTER_TIME } from '@/constants';
import { DATE_TIME_FORMAT, TIME_FORMAT } from '@/constants/datetime';
import { getClusterDetailReq } from '@/services';
import { deleteClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import { floorToTwoDecimalPlaces } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { DownOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Checkbox } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import dayjs from 'dayjs';
import { isEmpty } from 'lodash';
import { useEffect, useRef, useState } from 'react';
import BasicInfo from './BasicInfo';
import NFSInfoModal from './NFSInfoModal';
import ResourceDrawer from './ResourceDrawer';
import ServerTable from './ServerTable';
import ZoneTable from './ZoneTable';
const { RangePicker } = DatePicker;

const ClusterOverview: React.FC = () => {
  const { setChooseClusterName } = useModel('global');
  const access = useAccess();
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  const [resourceDrawerOpen, setResourceDrawerOpen] = useState<boolean>(false);
  const { ns, name } = useParams();
  const chooseZoneName = useRef<string>('');
  const timerRef = useRef<NodeJS.Timeout>();
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  const [mountNFSModal, setMountNFSModal] = useState<boolean>(false);
  const [removeNFSModal, setRemoveNFSModal] = useState<boolean>(false);
  const [downloadModal, setDownloadModal] = useState(false);
  const [diagnoseModal, setDiagnoseModal] = useState(false);
  const [diagnoseStatus, setDiagnoseStatus] = useState('');
  const [jobValue, setjobValue] = useState({});
  const [attachmentValue, setAttachmentValue] = useState('');
  // const [originalTimeRange, setOriginalTimeRange] = useState<[string, string]>([
  //   '',
  //   '',
  // ]);
  const modalType = useRef<API.ModalType>('addZone');

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

  const { run: getJob } = useRequest(job.getJob, {
    manual: true,
    onSuccess: ({ data }) => {
      console.log('data', data);
      setAttachmentValue(data?.result?.attachmentId || '');
      setDiagnoseStatus(data?.status);

      // 如果状态不是successful或failed，继续轮询
      if (data?.status !== 'successful' && data?.status !== 'failed') {
        setTimeout(() => {
          // 使用保存的原始时间参数进行轮询
          getJob(jobValue?.namespace, jobValue?.name);
        }, 2000); // 每2秒轮询一次
      }
    },
  });

  const { run: downloadOBClusterLog } = useRequest(
    obcluster.downloadOBClusterLog,
    {
      manual: true,
      onSuccess: ({ data }) => {
        console.log('data data', data, data?.namespace);
        getJob(data?.namespace, data?.name);
        setDiagnoseStatus(data?.status);
        setjobValue(data);
        setDiagnoseModal(false);
        setDownloadModal(true);
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

  const {
    storage,
    resource,
    deletionProtection,
    backupVolume,
    pvcIndependent,
  } = clusterDetail?.info || {};

  // 不为空即为绑定了NFS
  const removeNFS = !!backupVolume;
  const menuChange = ({ key }: { key: string }) => {
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
    } else if (key === 'download') {
      setDiagnoseModal(true);
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
    {
      key: 'download',
      label: '日志下载',
      disabled: isEmpty(clusterDetail),
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

  useEffect(() => {
    getClusterDetail({ ns: ns!, name: name! });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);
  const [form] = Form.useForm();

  // 设置默认时间为当前时间
  const defaultTime = dayjs();

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
          obVersion: clusterDetail?.info?.version,
        }}
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

      <DownloadModal
        visible={downloadModal}
        onCancel={() => setDownloadModal(false)}
        onOk={() => setDownloadModal(false)}
        title={'日志下载'}
        diagnoseStatus={diagnoseStatus}
        attachmentValue={attachmentValue}
        jobValue={jobValue}
      />
      <Modal
        title="日志下载"
        open={diagnoseModal}
        onCancel={() => setDiagnoseModal(false)}
        onOk={() => {
          form.validateFields().then((values) => {
            // 转换时间格式为 YYYY-MM-DD HH:mm:ss
            const formattedValues = {
              ...values,
              range: values.range?.map((time: any) => {
                if (time && typeof time.format === 'function') {
                  return time.format('YYYY-MM-DD HH:mm:ss');
                }
                return time;
              }),
            };

            // 调用下载日志API
            if (formattedValues.range && formattedValues.range.length === 2) {
              // 保存原始时间范围用于轮询
              // setOriginalTimeRange([
              //   formattedValues.range[0],
              //   formattedValues.range[1],
              // ]);
              downloadOBClusterLog(
                ns!,
                name!,
                formattedValues.range[0],
                formattedValues.range[1],
              );
            }
          });
        }}
      >
        <Form
          form={form}
          initialValues={{
            range: [defaultTime, defaultTime],
          }}
        >
          <Form.Item name="range">
            <RangePicker
              showTime={{
                format: TIME_FORMAT,
                defaultValue: [defaultTime, defaultTime],
              }}
              format={DATE_TIME_FORMAT}
              placeholder={['开始时间', '结束时间']}
              style={{ width: '100%' }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default ClusterOverview;
