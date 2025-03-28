import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Badge, Card, Col, Row } from 'antd';

import clusterImg from '@/assets/cluster/running.svg';
import kubernetesImg from '@/assets/kubernetes _logo.png';
import tenantImg from '@/assets/tenant.svg';

import { getClusterStatisticReq, getK8sObclusterListReq } from '@/services';
import { getTenantStatisticReq } from '@/services/tenant';
import { formatStatisticData } from '@/utils/helper';
import styles from './index.less';

export default function OverviewStatus() {
  const { data: clusterStatisticRes } = useRequest(getClusterStatisticReq);
  const { data: tenantStatisticRes } = useRequest(getTenantStatisticReq);
  const { data: K8sClustersList } = useRequest(getK8sObclusterListReq);
  const clusterHref = '/#/cluster';
  const tenantHref = '/#/tenant';
  const k8sClusterHref = '/#/k8scluster';
  const clusterStatistic = clusterStatisticRes?.data;
  const tenantStatistic = tenantStatisticRes?.data;

  const k8sClusterData = [
    {
      status: 'running',
      count: K8sClustersList?.data?.length || 0,
    },
  ];

  const k8sClusterStatistic = formatStatisticData('k8sCluster', k8sClusterData);

  const CustomBadge = ({
    text,
    type,
    count,
    status,
  }: {
    text: string;
    type: 'cluster' | 'tenant' | 'k8sCluster';
    count: number;
    status: 'processing' | 'error' | 'default' | 'warning';
  }) => {
    return (
      <div>
        <Badge status={status} text={text} />
        <a
          className={count === 0 ? styles.zeroText : ''}
          style={{ marginLeft: '8px' }}
          href={
            type === 'cluster'
              ? clusterHref
              : type === 'tenant'
              ? tenantHref
              : k8sClusterHref
          }
        >
          {count}
        </a>
      </div>
    );
  };

  const getStatisticConfig = (statistic: API.StatisticData) => ({
    running: (
      <CustomBadge
        key={'running'}
        text={intl.formatMessage({
          id: 'Dashboard.pages.Overview.OverviewStatus.Running',
          defaultMessage: '运行中',
        })}
        status="processing"
        type={statistic.type}
        count={statistic.running}
      />
    ),

    deleting: (
      <CustomBadge
        key={'deleting'}
        text={intl.formatMessage({
          id: 'Dashboard.pages.Overview.OverviewStatus.Deleting',
          defaultMessage: '删除中',
        })}
        status="error"
        type={statistic.type}
        count={statistic.deleting}
      />
    ),

    operating: (
      <CustomBadge
        key={'operating'}
        text={intl.formatMessage({
          id: 'Dashboard.pages.Overview.OverviewStatus.InOperation',
          defaultMessage: '操作中',
        })}
        status="warning"
        type={statistic.type}
        count={statistic.operating}
      />
    ),

    failed: (
      <CustomBadge
        key={'failed'}
        text={intl.formatMessage({
          id: 'Dashboard.pages.Overview.OverviewStatus.AnErrorHasOccurred',
          defaultMessage: '已出错',
        })}
        status="default"
        type={statistic.type}
        count={statistic.failed}
      />
    ),
  });

  return (
    <>
      {clusterStatistic && tenantStatistic && k8sClusterStatistic
        ? [clusterStatistic, tenantStatistic, k8sClusterStatistic].map(
            (statistic, index) => {
              return (
                <Col
                  span={8}
                  className={styles.overviewStatusContainerx}
                  key={index}
                >
                  <Card className={styles.cardContent}>
                    {statistic.type === 'k8sCluster' ? (
                      <>
                        <div
                          style={{
                            fontSize: '16px',
                            fontWeight: 'bold',
                            marginLeft: '30px',
                            marginTop: '21px',
                          }}
                        >
                          K8S 集群
                        </div>
                      </>
                    ) : null}
                    <Row>
                      <Col
                        span={4}
                        style={{
                          textAlign: 'center',
                        }}
                      >
                        {statistic.type !== 'k8sCluster' ? (
                          <img
                            src={
                              statistic.type === 'cluster'
                                ? clusterImg
                                : tenantImg
                            }
                            className={styles.imgContent}
                            alt="svg"
                          />
                        ) : null}

                        <div
                          className={styles.total}
                          style={
                            statistic.type === 'k8sCluster'
                              ? { width: '88px' }
                              : {}
                          }
                        >
                          <div
                            className={styles.totalText}
                            style={
                              statistic.type === 'k8sCluster'
                                ? { marginLeft: '40px' }
                                : {}
                            }
                          >
                            {statistic.total}
                          </div>
                          <div
                            className={styles.totalTitle}
                            style={
                              statistic.type === 'k8sCluster'
                                ? { marginLeft: '40px' }
                                : {}
                            }
                          >
                            {intl.formatMessage({
                              id: 'dashboard.pages.Overview.OverviewStatus.TotalQuantity',
                              defaultMessage: '总数量',
                            })}
                          </div>
                        </div>
                      </Col>
                      <Col span={14} offset={6}>
                        {statistic?.type !== 'k8sCluster' ? (
                          <>
                            <div className={styles.name}>{statistic.name}</div>
                            <div>
                              {Object.keys(getStatisticConfig(statistic)).map(
                                (key) => getStatisticConfig(statistic)[key],
                              )}
                            </div>
                          </>
                        ) : (
                          <img
                            src={kubernetesImg}
                            alt="png"
                            style={{
                              width: '90px',
                              height: '90px',
                              marginLeft: '10px',
                            }}
                          />
                        )}
                      </Col>
                    </Row>
                  </Card>
                </Col>
              );
            },
          )
        : null}
    </>
  );
}
