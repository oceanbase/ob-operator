import { QuestionCircleOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { Card,Col,Row,Tooltip } from 'antd';
import { useState } from 'react';

import type { QueryRangeType } from '@/components/MonitorDetail';
import { getAllMetrics } from '@/services';
import LineGraph,{ LineGraphProps,MetricType } from './LineGraph';
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
  queryRange: QueryRangeType;
  isRefresh?: boolean;
  queryScope:API.EventObjectType;
  type: API.MonitorUseTarget;
  groupLabels:API.LableKeys[];
  useFor?: API.MonitorUseFor;
}

export default function MonitorComp({
  filterLabel,
  queryRange,
  isRefresh = false,
  type,
  queryScope,
  groupLabels,
  useFor='cluster'
}: MonitorCompProps) {
  const { data: allMetrics } = useRequest(getAllMetrics, {
    defaultParams: [queryScope],
  });
  const [visible, setVisible] = useState(false);
  const [modalProps, setModalProps] = useState<LineGraphProps>({});
  const Title = ({
    metrics,
    name,
  }: {
    metrics: MetricType[];
    name: string;
  }) => {
    return (
      <div>
        {name}
        <Tooltip
          title={
            <ul>
              {metrics.map((metric, idx) => (
                <li key={idx}>
                  {metric.name}:{metric.description}
                </li>
              ))}
            </ul>
          }
        >
          <QuestionCircleOutlined
            style={{
              color: 'rgba(0, 0, 0, 0.45)',
              cursor: 'help',
              marginLeft: '4px',
            }}
          />
        </Tooltip>
      </div>
    );
  };
  return (
    <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
      {allMetrics &&
        allMetrics.map((container: any, index: number) => (
          <Col span={24} key={index}>
            <Card bodyStyle={{ padding: 0 }}>
              <div>
                <div className={styles.monitorHeader}>
                  {type === 'OVERVIEW' ? (
                    <h2>{container.name}</h2>
                  ) : (
                    <p className={styles.headerText}>{container.name}</p>
                  )}
                </div>
                <div className={styles.monitorContainer}>
                  {container.metricGroups.map(
                    (graphContainer: any, graphIdx: number) => (
                      <Card className={styles.monitorItem} key={graphIdx}>
                        <div className={styles.graphHeader}>
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
