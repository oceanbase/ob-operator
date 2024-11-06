import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { history, useAccess, useModel, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Row, message } from 'antd';
import { useEffect, useRef, useState } from 'react';

import EventsTable from '@/components/EventsTable';
import OperateModal from '@/components/customModal/OperateModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { REFRESH_CLUSTER_TIME } from '@/constants';
import { getClusterDetailReq } from '@/services';
import { deleteClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import BasicInfo from './BasicInfo';
import ServerTable from './ServerTable';
import ZoneTable from './ZoneTable';

const ClusterOverview: React.FC = () => {
  const { setChooseClusterName } = useModel('global');
  const access = useAccess();
  const [operateModalVisible, setOperateModalVisible] =
    useState<boolean>(false);
  const { ns, name } = useParams();
  const chooseZoneName = useRef<string>('');
  const timerRef = useRef<NodeJS.Timeout>();
  const [chooseServerNum, setChooseServerNum] = useState<number>(1);
  const modalType = useRef<API.ModalType>('addZone');
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

  const header = () => {
    return {
      title: intl.formatMessage({
        id: 'Dashboard.Detail.Overview.ClusterOverview',
        defaultMessage: '集群概览',
      }),
      extra: access.obclusterwrite
        ? [
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

  return (
    <PageContainer header={header()}>
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <BasicInfo {...(clusterDetail.info as API.ClusterInfo)} />
          </Col>
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
        {clusterDetail && <ServerTable clusterDetail={clusterDetail} />}
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
    </PageContainer>
  );
};

export default ClusterOverview;
