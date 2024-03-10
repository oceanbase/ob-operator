import EventsTable from '@/components/EventsTable';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import OperateModal from '@/components/customModal/OperateModal';
import { REFRESH_TENANT_TIME, RESULT_STATUS } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import {
  deleteTenent,
  getBackupJobs,
  getBackupPolicy,
  getTenant,
} from '@/services/tenant';
import { intl } from '@/utils/intl';
import { EllipsisOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Row, Tooltip, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import Backups from './Backups';
import BasicInfo from './BasicInfo';
import Replicas from './Replicas';
import styles from './index.less';

type OperateItemConfigType = {
  text: string;
  onClick: () => void;
  show: boolean;
  isMore: boolean;
  danger?: boolean;
};

export default function TenantOverview() {
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  //Current operation and maintenance modal type
  const modalType = useRef<API.ModalType>('changeUnitCount');
  const timerRef = useRef<NodeJS.Timeout>();
  const [defaultUnitCount, setDefaultUnitCount] = useState<number>(1);
  const [[ns, name]] = useState(getNSName());

  const openOperateModal = (type: API.ModalType) => {
    modalType.current = type;
    setOperateModalVisible(true);
  };

  const handleDelete = async () => {
    const res = await deleteTenent({ ns, name });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'Dashboard.Detail.Overview.DeletedSuccessfully',
          defaultMessage: '删除成功',
        }),
      );
      history.push('/tenant');
    }
  };

  const {
    data: tenantDetailResponse,
    run: getTenantDetail,
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
    defaultParams: [{ name, ns }],
  });
  const { data: backupJobsResponse } = useRequest(getBackupJobs, {
    defaultParams: [{ name, ns, type: 'FULL' }],
  });

  const tenantDetail = tenantDetailResponse?.data;
  const backupPolicy = backupPolicyResponse?.data;
  const backupJobs = backupJobsResponse?.data;

  const operateListConfig: OperateItemConfigType[] = [
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.UnitSpecificationManagement',
        defaultMessage: 'Unit规格管理',
      }),
      onClick: () => openOperateModal('modifyUnitSpecification'),
      show: tenantDetail?.info.tenantRole === 'PRIMARY',
      isMore: false,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ChangePassword',
        defaultMessage: '修改密码',
      }),
      onClick: () => openOperateModal('changePassword'),
      show: tenantDetail?.info.tenantRole === 'PRIMARY',
      isMore: false,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.DeleteTenant',
        defaultMessage: '删除租户',
      }),
      onClick: () =>
        showDeleteConfirm({
          onOk: handleDelete,
          title: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.AreYouSureYouWant',
            defaultMessage: '你确定要删除该租户吗？',
          }),
        }),
      show: true,
      isMore: false,
      danger: true,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ActivateASecondaryTenant',
        defaultMessage: '激活备租户',
      }),
      onClick: () => openOperateModal('activateTenant'),
      show: tenantDetail?.info.tenantRole === 'STANDBY',
      isMore: true,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ActiveStandbySwitchover',
        defaultMessage: '主备切换',
      }),
      onClick: () => openOperateModal('switchTenant'),
      show: Boolean(tenantDetail?.source?.primaryTenant),
      isMore: true,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.StandbyTenantPlaybackLog',
        defaultMessage: '备租户回放日志',
      }),
      onClick: () => openOperateModal('logReplay'),
      show: tenantDetail?.info.tenantRole === 'STANDBY',
      isMore: true,
    },
    {
      text: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.AdjustTheNumberOfUnits',
        defaultMessage: '调整 Unit 数量',
      }),
      onClick: () => openOperateModal('changeUnitCount'),
      show: tenantDetail?.info.tenantRole === 'PRIMARY',
      isMore: true,
    },
  ];

  const OperateListModal = () => (
    <ul>
      {operateListConfig
        .filter((item) => item.isMore && item.show)
        .map((operateItem, index) => (
          <li key={index} onClick={operateItem.onClick}>
            {operateItem.text}
          </li>
        ))}
    </ul>
  );

  const operateSuccess = () => {
    setTimeout(() => {
      getTenantDetail({ ns, name });
    }, 1000);
  };
  const header = () => {
    let container = document.getElementById('tenant-detail-container');
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.Overview',
        defaultMessage: '概览',
      }),
      extra: [
        ...operateListConfig
          .filter((item) => item.show && !item.isMore)
          .map((item, index) => (
            <Button onClick={item.onClick} danger={item.danger} key={index}>
              {item.text}
            </Button>
          )),
        <Tooltip
          getPopupContainer={() => container}
          title={<OperateListModal />}
          placement="bottomLeft"
          key={4}
        >
          <Button>
            <EllipsisOutlined />
          </Button>
        </Tooltip>,
      ],
    };
  };

  useEffect(() => {
    getTenantDetail({ ns, name });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  return (
    <div id="tenant-detail-container" className={styles.tenantContainer}>
      <PageContainer header={header()}>
        <Row justify="start" gutter={[16, 16]}>
          {tenantDetail && (
            <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
          )}

          {tenantDetail && tenantDetail.replicas && (
            <Replicas replicaList={tenantDetail.replicas} />
          )}

          <EventsTable defaultExpand={true} objectType="OBTENANT" collapsible={true} />
          <Backups backupJobs={backupJobs} backupPolicy={backupPolicy} />
        </Row>

        <OperateModal
          type={modalType.current}
          visible={operateModalVisible}
          setVisible={setOperateModalVisible}
          successCallback={operateSuccess}
          defaultValue={defaultUnitCount}
        />
      </PageContainer>
    </div>
  );
}
