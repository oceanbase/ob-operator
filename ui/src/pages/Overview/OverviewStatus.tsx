import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Badge, Card, Col, Row } from 'antd';

import clusterImg from '@/assets/cluster/running.svg';
import tenantImg from '@/assets/tenant.svg';
import { getClusterStatisticReq } from '@/services';
import { getTenantStatisticReq } from '@/services/tenant';
import styles from './index.less';

export default function OverviewStatus() {
  const { data: clusterStatisticRes } = useRequest(getClusterStatisticReq);
  const { data: tenantStatisticRes } = useRequest(getTenantStatisticReq);
  const clusterHref = '/#/cluster';
  const tenantHref = '/#/tenant';
  const clusterStatistic = clusterStatisticRes?.data;
  const tenantStatistic = tenantStatisticRes?.data;

  const CustomBadge = ({
    text,
    type,
    count,
    status,
  }: {
    text: string;
    type: 'cluster' | 'tenant';
    count: number;
    status: 'processing' | 'error' | 'default' | 'warning';
  }) => {
    return (
      <div>
        <Badge status={status} text={text} />
        <a
          className={count === 0 ? styles.zeroText : ''}
          style={{ marginLeft: '8px' }}
          href={type === 'cluster' ? clusterHref : tenantHref}
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
      {clusterStatistic && tenantStatistic
        ? [clusterStatistic, tenantStatistic].map((statistic, index) => {
            return (
              <Col
                span={8}
                className={styles.overviewStatusContainerx}
                key={index}
              >
                <Card className={styles.cardContent}>
                  <Row>
                    <Col
                      span={4}
                      style={{
                        textAlign: 'center',
                      }}
                    >
                      <img
                        src={
                          statistic.type === 'cluster' ? clusterImg : tenantImg
                        }
                        className={styles.imgContent}
                        alt="svg"
                      />

                      <div className={styles.total}>
                        <div className={styles.totalText}>
                          {statistic.total}
                        </div>
                        <div className={styles.totalTitle}>
                          {intl.formatMessage({
                            id: 'dashboard.pages.Overview.OverviewStatus.TotalQuantity',
                            defaultMessage: '总数量',
                          })}
                        </div>
                      </div>
                    </Col>
                    <Col span={14} offset={6}>
                      <div className={styles.name}>{statistic.name}</div>
                      <div>
                        {Object.keys(getStatisticConfig(statistic)).map(
                          (key) => getStatisticConfig(statistic)[key],
                        )}
                      </div>
                    </Col>
                  </Row>
                </Card>
              </Col>
            );
          })
        : null}
    </>
  );
}
