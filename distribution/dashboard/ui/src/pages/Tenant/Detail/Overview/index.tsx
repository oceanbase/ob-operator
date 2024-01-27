import showDeleteConfirm from '@/components/customModal/DeleteModal';
import OperateModal from '@/components/customModal/OperateModal';
import { REFRESH_CLUSTER_TIME } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { deleteTenent, getTenant } from '@/services/tenant';
import { EllipsisOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Row, Tooltip, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import BasicInfo from './BasicInfo';
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
  //当前运维弹窗类型
  const modalType = useRef<API.ModalType>('modifyUnit');

  const [[ns, name]] = useState(getNSName());

  const openOperateModal = (type: API.ModalType) => {
    modalType.current = type;
    setOperateModalVisible(true);
  };

  const handleDelete = async () => {
    const res = await deleteTenent({ ns, name });
    if (res.successful) {
      message.success('删除成功');
      history.push('/tenant');
    }
  };

  const { data: tenantDetailResponse, run: getTenantDetail } = useRequest(
    getTenant,
    {
      manual: true,
      onSuccess: ({ data }) => {
        if (data.info.status === 'operating') {
          setTimeout(() => {
            getTenantDetail({ ns, name });
          }, REFRESH_CLUSTER_TIME);
        }
      },
    },
  );

  const tenantDetail = tenantDetailResponse?.data;
  const operateListConfig: OperateItemConfigType[] = [
    {
      text: 'Unit规格管理',
      onClick: () => openOperateModal('modifyUnit'),
      show: tenantDetail?.info.tenantRole === 'Primary',
      isMore: false,
    },
    {
      text: '修改密码',
      onClick: () => openOperateModal('changePassword'),
      show: tenantDetail?.info.tenantRole === 'Primary',
      isMore: false,
    },
    {
      text: '删除租户',
      onClick: () =>
        showDeleteConfirm({
          onOk: handleDelete,
          title: '你确定要删除该租户吗？',
        }),
      show: true,
      isMore: false,
      danger: true,
    },
    {
      text: '激活备租户',
      onClick: () => openOperateModal('activateTenant'),
      show: tenantDetail?.info.tenantRole === 'Standby',
      isMore: true,
    },
    {
      text: '主备切换',
      onClick: () => openOperateModal('switchTenant'),
      show: Boolean(tenantDetail?.source?.primaryTenant),
      isMore: true,
    },
    {
      text: '备租户回放日志',
      onClick: () => openOperateModal('logReplay'),
      show: tenantDetail?.info.tenantRole === 'Standby',
      isMore: true,
    },
    {
      text: '租户版本升级',
      onClick: () => openOperateModal('upgradeTenant'),
      show:tenantDetail?.info.tenantRole === 'Primary',
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
      title: '概览',
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
  }, []);

  return (
    <div id="tenant-detail-container" className={styles.tenantContainer}>
      <PageContainer header={header()}>
        <Row gutter={[16, 16]}>
          {tenantDetail && (
            <BasicInfo info={tenantDetail.info} source={tenantDetail.source} />
          )}
        </Row>
        <OperateModal
          type={modalType.current}
          visible={operateModalVisible}
          setVisible={setOperateModalVisible}
          successCallback={operateSuccess}
        />
      </PageContainer>
    </div>
  );
}
