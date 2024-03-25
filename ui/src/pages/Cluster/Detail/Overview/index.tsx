import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { history,useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button,Row,message } from 'antd';
import { useEffect,useRef,useState } from 'react';

import EventsTable from '@/components/EventsTable';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import OperateModal from '@/components/customModal/OperateModal';
import { REFRESH_CLUSTER_TIME } from '@/constants';
import { deleteObcluster,getClusterDetailReq } from '@/services';
import BasicInfo from './BasicInfo';
import ServerTable from './ServerTable';
import ZoneTable from './ZoneTable';
import { getNSName } from './helper';

//集群详情概览页
const ClusterOverview: React.FC = () => {
  const { setChooseClusterName } = useModel('global');
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  const [[ns, name]] = useState(getNSName());
  const chooseZoneName = useRef<string>('');
  const timerRef = useRef<NodeJS.Timeout>();
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  //当前运维弹窗类型
  const modalType = useRef<API.ModalType>('addZone');
  const { data: clusterDetail, run: getClusterDetail } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
      onSuccess: (data) => {
        setChooseClusterName(data.info.clusterName);
        if (data.status === 'operating') {
          timerRef.current = setTimeout(() => {
            getClusterDetail({ ns, name });
          }, REFRESH_CLUSTER_TIME);
        } else if (timerRef.current) {
          clearTimeout(timerRef.current);
        }
      },
    },
  );
  const handleDelete = async () => {
    const res = await deleteObcluster({ ns, name });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.DeletedSuccessfully',
          defaultMessage: '删除成功',
        }),
      );
      history.push('/cluster');
    }
  };

  const operateSuccess = () => {
    setTimeout(() => {
      getClusterDetail({ ns, name });
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

  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ClusterOverview',
        defaultMessage: '集群概览',
      }),
      extra: [
        <Button
          onClick={handleAddZone}
          disabled={
            clusterDetail?.status === 'operating' ||
            clusterDetail?.status === 'failed'
          }
          key="1"
        >
          {intl.formatMessage({
            id: 'dashboard.Detail.Overview.AddZone',
            defaultMessage: '新增Zone',
          })}
        </Button>,
        <Button
          key="2"
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
        </Button>,
        <Button
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
          key="3"
          type="primary"
          danger
        >
          {intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.Delete',
            defaultMessage: '删除',
          })}
        </Button>,
      ],
    };
  };

  useEffect(() => {
    getClusterDetail({ ns, name });

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  return (
    <PageContainer header={header()}>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
        )}

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
          <EventsTable
            objectType="OBCLUSTER"
            name={clusterDetail?.info?.name}
          />
        )}

        {clusterDetail && (
          <ServerTable servers={clusterDetail.servers as API.Server[]} />
        )}
      </Row>
      <OperateModal
        type={modalType.current}
        visible={operateModalVisible}
        setVisible={setOperateModalVisible}
        successCallback={operateSuccess}
        params={{
          zoneName:chooseZoneName.current,
          defaultValue:chooseServerNum
        }}
      />
    </PageContainer>
  );
};

export default ClusterOverview;
