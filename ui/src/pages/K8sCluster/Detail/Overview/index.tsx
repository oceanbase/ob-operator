import { K8sClusterApi } from '@/api';
import EventsTable from '@/components/EventsTable';
import NodesTable from '@/pages/Overview/NodesTable';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { formatTime } from '@oceanbase/util';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Card, Col, Descriptions, Row } from 'antd';
import dayjs from 'dayjs';

const K8sClusterOverview: React.FC = () => {
  const params = useParams();
  const { k8sclusterName } = params;
  const { data: K8sClustersData, loading } = useRequest(
    K8sClusterApi.getRemoteK8sCluster,
    {
      defaultParams: [k8sclusterName],
    },
  );

  const {
    data: K8sClustersNodeData,
    loading: nodeLoading,
    refresh: K8sNodeRefresh,
  } = useRequest(K8sClusterApi.listRemoteK8sNodes, {
    defaultParams: [k8sclusterName],
  });

  const formatNode = () => {
    const res = [];
    for (const node of K8sClustersNodeData?.data || []) {
      const obj = {};
      Object.assign(obj, node.info, node.resource);
      obj.cpu = ((obj.cpuUsed / obj.cpuTotal) * 100).toFixed(1);
      obj.memory = ((obj.memoryUsed / obj.memoryTotal) * 100).toFixed(1);
      obj.uptime = dayjs.unix(obj.uptime).format('YYYY-MM-DD HH:mm:ss');
      res.push(obj);
    }

    return res;
  };

  const clusterDetail = K8sClustersData?.data;

  const { data: K8sClustersEvents, loading: getK8sEventsReqLoading } =
    useRequest(K8sClusterApi.listRemoteK8sEvents, {
      defaultParams: [k8sclusterName],
      onSuccess: (r) => {
        if (r.successful) {
          let count = 0;
          r.data.sort((pre, next) => next.lastSeen - pre.lastSeen);
          for (const event of r.data) {
            event.id = ++count;
            event.firstOccur = dayjs
              .unix(event.firstOccur)
              .format('YYYY-MM-DD HH:mm:ss');
            event.lastSeen = dayjs
              .unix(event.lastSeen)
              .format('YYYY-MM-DD HH:mm:ss');
          }
        }
        return r.data;
      },
    });

  return (
    <PageContainer
      header={{
        title: 'k8s 集群概览',
      }}
      loading={loading}
    >
      <Row gutter={[16, 16]}>
        {clusterDetail && (
          <Col span={24}>
            <Card
              title={intl.formatMessage({
                id: 'Dashboard.Detail.Overview.BasicInfo.BasicClusterInformation',
                defaultMessage: '集群基本信息',
              })}
            >
              <Descriptions column={3}>
                <Descriptions.Item
                  label={intl.formatMessage({
                    id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterName',
                    defaultMessage: '集群名',
                  })}
                >
                  {clusterDetail?.name}
                </Descriptions.Item>
                <Descriptions.Item label={'描述信息'}>
                  {clusterDetail?.description || '-'}
                </Descriptions.Item>
                <Descriptions.Item label={'创建日期'}>
                  {formatTime(clusterDetail?.createdAt)}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
        )}
        <NodesTable
          loading={nodeLoading}
          type="k8s"
          k8sClusterName={k8sclusterName}
          K8sClustersNodeList={formatNode()}
          onSuccess={() => {
            K8sNodeRefresh();
          }}
        />
        <Col span={24}>
          <EventsTable
            externaLoading={getK8sEventsReqLoading}
            externalData={K8sClustersEvents?.data}
            type={'k8s'}
          />
        </Col>
      </Row>
    </PageContainer>
  );
};

export default K8sClusterOverview;
