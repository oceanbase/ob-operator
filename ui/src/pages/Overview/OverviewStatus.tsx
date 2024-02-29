import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Badge, Card, Col, Row } from 'antd';

import clusterImg from '@/assets/cluster/running.svg';
import tenantImg from '@/assets/tenant.svg';
import { getOBStatisticReq } from '@/services';
import styles from './index.less';

export default function OverviewStatus() {
  //每次请求完成后，useRequest 钩子函数会根据请求结果更新 data 的值 可以直接使用data值
  const { data } = useRequest(getOBStatisticReq);
  const clusterHref = '/#/cluster';
  const tenantHref = '/#/tenant';
  return (
    <>
      {data &&
        data.map((item, index) => {
          return (
            <Col
              span={8}
              className={styles.overviewStatusContainerx}
              key={index}
            >
              <Card
                className={`${styles.cardContent} ${
                  item.type === 'tenant' && styles.banContent
                }`}
              >
                <Row>
                  <Col
                    span={4}
                    style={{
                      textAlign: 'center',
                    }}
                  >
                    <img
                      src={item.type === 'cluster' ? clusterImg : tenantImg}
                      className={styles.imgContent}
                      alt="svg"
                    />

                    <div className={styles.total}>
                      <div className={styles.totalText}>{item.total}</div>
                      <div className={styles.totalTitle}>
                        {intl.formatMessage({
                          id: 'dashboard.pages.Overview.OverviewStatus.TotalQuantity',
                          defaultMessage: '总数量',
                        })}
                      </div>
                    </div>
                  </Col>
                  <Col span={14} offset={6}>
                    <div className={styles.name}>{item.name}</div>
                    <div>
                      <div>
                        <Badge
                          status="processing"
                          text={intl.formatMessage({
                            id: 'dashboard.pages.Overview.OverviewStatus.Running',
                            defaultMessage: '运行中',
                          })}
                        />

                        <a
                          className={item.running === 0 ? styles.zeroText : ''}
                          style={{ marginLeft: '8px' }}
                          href={
                            item.type === 'cluster' ? clusterHref : tenantHref
                          }
                        >
                          {item.running}
                        </a>
                      </div>
                      <div>
                        <Badge
                          status="error"
                          text={intl.formatMessage({
                            id: 'OBDashboard.pages.Overview.OverviewStatus.Deleting',
                            defaultMessage: '删除中',
                          })}
                        />
                        <a
                          className={item.deleting === 0 ? styles.zeroText : ''}
                          style={{ marginLeft: '8px' }}
                          href={
                            item.type === 'cluster' ? clusterHref : tenantHref
                          }
                        >
                          {item.deleting}
                        </a>
                      </div>
                      <div>
                        <Badge
                          status="warning"
                          text={intl.formatMessage({
                            id: 'OBDashboard.pages.Overview.OverviewStatus.InOperation',
                            defaultMessage: '操作中',
                          })}
                        />
                        <a
                          className={
                            item.operating === 0 ? styles.zeroText : ''
                          }
                          style={{ marginLeft: '8px' }}
                          href={
                            item.type === 'cluster' ? clusterHref : tenantHref
                          }
                        >
                          {item.operating}
                        </a>
                      </div>
                    </div>
                  </Col>
                </Row>
              </Card>
            </Col>
          );
        })}
    </>
  );
}
