import { useRequest } from 'ahooks';
import { Card, Col, Row } from 'antd';

import { getAllMetrics } from '@/services';
import IconTip from '../IconTip';
import LineGraph from './LineGraph';
import styles from './index.less';

/**
 * Queryable label:
 * ob_cluster_name
 * ob_cluster_id
 * tenant_name
 * tenant_id
 * svr_ip
 * obzone
 */

interface MonitorCompProps {
  filterLabel: API.MetricsLabels;
  queryRange: Monitor.QueryRangeType;
  isRefresh?: boolean;
  queryScope: API.MetricScope;
  type: API.MonitorUseTarget;
  groupLabels: API.LableKeys[];
  useFor?: API.MonitorUseFor;
  filterData?: API.ClusterItem[] | API.TenantDetail[];
  filterQueryMetric?: API.MetricsLabels;
}

export default function MonitorComp({
  filterLabel,
  queryRange,
  isRefresh = false,
  type,
  queryScope,
  groupLabels,
  useFor = 'cluster',
  filterData,
  filterQueryMetric
}: MonitorCompProps) {
  const { data: allMetrics } = useRequest(getAllMetrics, {
    defaultParams: [queryScope],
  });
  // const [visible, setVisible] = useState(false);
  // const [modalProps, setModalProps] = useState<LineGraphProps>({});
  // const Title = ({
  //   metrics,
  //   name,
  // }: {
  //   metrics: MetricType[];
  //   name: string;
  // }) => {
  //   return (
  //     <div>
  //       {name}
  //       <Tooltip
  //         title={
  //           <ul>
  //             {metrics.map((metric, idx) => (
  //               <li key={idx}>
  //                 {metric.name}:{metric.description}
  //               </li>
  //             ))}
  //           </ul>
  //         }
  //       >
  //         <QuestionCircleOutlined
  //           style={{
  //             color: 'rgba(0, 0, 0, 0.45)',
  //             cursor: 'help',
  //             marginLeft: '4px',
  //           }}
  //         />
  //       </Tooltip>
  //     </div>
  //   );
  // };
  return (
    <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
      {allMetrics &&
        allMetrics.map((container: any, index: number) => (
          <Col span={24} key={index}>
            <Card
              bodyStyle={{ padding: 0 }}
              title={
                <h2 style={{marginBottom: 0}}>{container.name}</h2>
              }
            >
              <div>
                <div className={styles.monitorContainer}>
                  {container.metricGroups.map(
                    (graphContainer: any, graphIdx: number) => (
                      <Card className={styles.monitorItem} key={graphIdx}>
                        <div className={styles.graphHeader}>
                          <IconTip
                            tip={graphContainer.description}
                            style={{ fontSize: 16 }}
                            content={
                              <span className={styles.graphHeaderText}>
                                {graphContainer.name}
                                {graphContainer.metrics[0].unit &&
                                  `(
                                ${graphContainer.metrics[0].unit}
                                ${
                                  (graphContainer.metrics[0].unit && type) ===
                                  'OVERVIEW'
                                    ? ','
                                    : ''
                                }
                                ${
                                  type === 'OVERVIEW'
                                    ? graphContainer.metrics[0].key
                                    : ''
                                }
                                )`}
                              </span>
                            }
                          />
                          {/* <Tooltip title="放大查看">
                            <FullscreenOutlined
                              className={styles.fullscreen}
                              onClick={() => {
                                console.log("graphContainer.name",graphContainer.name)
                                setVisible(true);
                                setModalProps({
                                  id: `monitor-${graphContainer.name.replace(
                                    /\s+/g,
                                    '',
                                  )}-detail`,
                                  metrics: graphContainer.metrics,
                                  labels: [clusterName],
                                  height: 300,
                                  name: graphContainer.name,
                                });
                              }}
                            />
                          </Tooltip> */}
                        </div>
                        <LineGraph
                          id={`monitor-${graphContainer.name.replace(
                            /\s+/g,
                            '',
                          )}`}
                          isRefresh={isRefresh}
                          queryRange={queryRange}
                          metrics={graphContainer.metrics}
                          labels={filterLabel}
                          groupLabels={groupLabels}
                          type={type}
                          useFor={useFor}
                          filterData={filterData}
                          filterQueryMetric={filterQueryMetric}
                        />
                      </Card>
                    ),
                  )}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      {/* <LineGraphModal
        title={<Title metrics={modalProps.metrics} name={modalProps.name} />}
        width={960}
        visible={visible}
        setVisible={setVisible}
      >
        <LineGraph {...modalProps} />
      </LineGraphModal> */}
    </Row>
  );
}
