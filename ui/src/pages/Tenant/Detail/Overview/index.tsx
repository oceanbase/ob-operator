import EventsTable from '@/components/EventsTable';
import OperateModal from '@/components/customModal/OperateModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { REFRESH_TENANT_TIME, RESULT_STATUS } from '@/constants';
import {
  getEssentialParameters as getEssentialParametersReq,
  getSimpleClusterList,
} from '@/services';
import { deleteTenantReportWrap } from '@/services/reportRequest/tenantReportReq';
import { getBackupJobs, getBackupPolicy, getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { DownOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history, useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Dropdown, MenuProps, Row, Space, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import {
  getClusterFromTenant,
  getOriginResourceUsages,
  getZonesOptions,
} from '../../helper';
import Backups from './Backups';
import BasicInfo from './BasicInfo';
import Replicas from './Replicas';
import styles from './index.less';

export default function TenantOverview() {
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  const access = useAccess();
  //Current operation and maintenance modal type
  const modalType = useRef<API.ModalType>('changeUnitCount');
  const operateTypeRef = useRef<OBTenant.OperateType>();
  const timerRef = useRef<NodeJS.Timeout>();
  const [defaultUnitCount, setDefaultUnitCount] = useState<number>(1);
  const { ns, name } = useParams();
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>([]);
  const [editZone, setEditZone] = useState<string>('');
  useRequest(getSimpleClusterList, {
    onSuccess: ({ successful, data }) => {
      if (successful) {
        data.forEach((cluster) => {
          cluster.topology.forEach((zone) => {
            zone.checked = false;
          });
        });
        setClusterList(data);
      }
    },
  });
  const { data: essentialParameterRes, run: getEssentialParameters } =
    useRequest(getEssentialParametersReq, {
      manual: true,
    });

  const openOperateModal = (
    type: API.ModalType,
    operateType?: OBTenant.OperateType,
  ) => {
    if (operateType) {
      operateTypeRef.current = operateType;
    }
    modalType.current = type;
    setOperateModalVisible(true);
  };

  const handleDelete = async () => {
    const res = await deleteTenantReportWrap({ ns: ns!, name: name! });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'Dashboard.Detail.Overview.DeletedSuccessfully',
          defaultMessage: '删除成功',
        }),
      );
      history.replace('/tenant');
    }
  };

  const {
    data: tenantDetailResponse,
    run: getTenantDetail,
    loading,
    refresh: reGetTenantDetail,
  } = useRequest(getTenant, {
    manual: true,
    onSuccess: ({ data, successful }) => {
      if (successful) {
        if (data.info.unitNumber) {
          setDefaultUnitCount(data.info.unitNumber);
        }
        if (!RESULT_STATUS.includes(data.info.status)) {
          timerRef.current = setTimeout(() => {
            reGetTenantDetail();
          }, REFRESH_TENANT_TIME);
        } else if (timerRef.current) {
          clearTimeout(timerRef.current);
        }
      }
    },
  });
  const { data: backupPolicyResponse } = useRequest(getBackupPolicy, {
    defaultParams: [{ name: name!, ns: ns! }],
  });
  const { data: backupJobsResponse } = useRequest(getBackupJobs, {
    defaultParams: [{ name: name!, ns: ns!, type: 'FULL' }],
  });

  const tenantDetail = tenantDetailResponse?.data;
  const backupPolicy = backupPolicyResponse?.data;
  const backupJobs = backupJobsResponse?.data;
  const essentialParameter = essentialParameterRes?.data;

  const operateSuccess = () => {
    setTimeout(() => {
      getTenantDetail({ ns: ns!, name: name! });
    }, 1000);
  };
 
  const items: MenuProps['items'] = [
    ...(tenantDetail?.info.tenantRole === 'PRIMARY'
      ? [
          {
            label: intl.formatMessage({
              id: 'Dashboard.Detail.Overview.ChangePassword',
              defaultMessage: '修改密码',
            }),
            key: 'changePassword',
          },
          {
            label: intl.formatMessage({
              id: 'Dashboard.Detail.Overview.AdjustTheNumberOfUnits',
              defaultMessage: '调整 Unit 数量',
            }),
            key: 'changeUnitCount',
          },
          {
            label: '创建租户恢复策略',
            key: 'createBackupPolicy',
          },
        ]
      : []),
    ...(tenantDetail?.info.tenantRole === 'STANDBY'
      ? [
          {
            label: intl.formatMessage({
              id: 'Dashboard.Detail.Overview.ActivateASecondaryTenant',
              defaultMessage: '激活备租户',
            }),
            key: 'activateTenant',
            onClick: () => openOperateModal('activateTenant'),
          },
          {
            label: intl.formatMessage({
              id: 'Dashboard.Detail.Overview.StandbyTenantPlaybackLog',
              defaultMessage: '备租户回放日志',
            }),
            key: 'logReplay',
          },
        ]
      : []),
    ...(tenantDetail?.source?.primaryTenant
      ? [
          {
            label: intl.formatMessage({
              id: 'Dashboard.Detail.Overview.ActiveStandbySwitchover',
              defaultMessage: '主备切换',
            }),
            key: 'switchTenant',
          },
        ]
      : []),
    {
      label: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.DeleteTenant',
        defaultMessage: '删除租户',
      }),
      key: 'deleteTenant',
      danger: true,
    },
  ];
  const menuChange = ({ key }) => {
    if (key === 'createBackupPolicy') {
      history.push(
        `/tenant/${ns}/${name}/${tenantDetail?.info?.name}/backup/new?overview=true`,
      );
    } else if (key === 'deleteTenant') {
      showDeleteConfirm({
        onOk: handleDelete,
        title: intl.formatMessage({
          id: 'Dashboard.Detail.Overview.AreYouSureYouWant',
          defaultMessage: '你确定要删除该租户吗？',
        }),
      });
    } else if (key) {
      openOperateModal(key);
    }
  };
  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.TenantOverview',
        defaultMessage: '租户概览',
      }),
      extra: access.obclusterwrite
        ? [
            <Dropdown menu={{ items, onClick: menuChange }}>
              <Button>
                <Space>
                  租户管理
                  <DownOutlined />
                </Space>
              </Button>
            </Dropdown>,
          ]
        : [],
    };
  };

  useEffect(() => {
    getTenantDetail({ name: name!, ns: ns! });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (tenantDetail && clusterList) {
      const cluster = getClusterFromTenant(
        clusterList,
        tenantDetail.info.clusterResourceName,
      );
      if (cluster) {
        const { name, namespace } = cluster;
        getEssentialParameters({
          ns: namespace,
          name,
        });
      }
    }
  }, [clusterList, tenantDetail]);

  const isCreateResourcePool = operateTypeRef.current === 'create';

  return (
    <div id="tenant-detail-container" className={styles.tenantContainer}>
      <PageContainer header={header()}>
        <Row justify="start" gutter={[16, 16]}>
          <Col span={24}>
            <BasicInfo
              loading={loading}
              info={tenantDetail?.info}
              source={tenantDetail?.source}
              name={name}
              ns={ns}
            />
          </Col>

          {tenantDetail && tenantDetail.replicas && (
            <Replicas
              refreshTenant={reGetTenantDetail}
              replicaList={tenantDetail.replicas}
              openOperateModal={openOperateModal}
              cluster={getClusterFromTenant(
                clusterList,
                tenantDetail.info.clusterResourceName,
              )}
              tenantStatus={tenantDetail?.info?.status}
              setEditZone={setEditZone}
              editZone={editZone}
              operateType={operateTypeRef}
            />
          )}
          {tenantDetail && (access.systemread || access.systemwrite) ? (
            <Col span={24}>
              <EventsTable
                defaultExpand={true}
                objectType={['OBTENANT', 'OBBACKUPPOLICY']}
                collapsible={true}
                name={tenantDetail?.info.name}
              />
            </Col>
          ) : null}

          <Backups
            loading={loading}
            backupJobs={backupJobs}
            backupPolicy={backupPolicy}
          />
        </Row>

        <OperateModal
          type={modalType.current}
          visible={operateModalVisible}
          setVisible={setOperateModalVisible}
          successCallback={operateSuccess}
          params={{
            defaultUnitCount,
            clusterList: clusterList,
            editZone,
            essentialParameter: isCreateResourcePool
              ? essentialParameter
              : getOriginResourceUsages(
                  essentialParameter,
                  tenantDetail?.replicas?.find(
                    (replica) => replica.zone === editZone,
                  ),
                ),
            clusterResourceName: tenantDetail?.info.clusterResourceName,
            setClusterList,
            setEditZone,
            replicaList: tenantDetail?.replicas,
            newResourcePool: isCreateResourcePool,
            zonesOptions: isCreateResourcePool
              ? getZonesOptions(
                  getClusterFromTenant(
                    clusterList,
                    tenantDetail?.info.clusterResourceName,
                  ),
                  tenantDetail?.replicas,
                )
              : undefined,
            obVersion: tenantDetail?.info?.version,
          }}
        />
      </PageContainer>
    </div>
  );
}
