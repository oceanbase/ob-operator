import { useRequest } from 'ahooks';
import { Card, Row } from 'antd';

import { getAllMetrics } from '@/services';
import { useState } from 'react';
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
  filterQueryMetric,
}: MonitorCompProps) {
  const [activeTabKey, setActiveTabKey] = useState<string>('0');

  const { data: allMetrics } = useRequest(getAllMetrics, {
    defaultParams: [queryScope],
  });
  const { data: allMetrics2 } = useRequest(getAllMetrics, {
    defaultParams: ['OBTENANT_OVERVIEW'],
  });

  const tabList =
    allMetrics2?.concat(allMetrics)?.map((container: any, index: number) => ({
      key: index.toString(),
      label: container?.name,
    })) || [];

  // 动态生成内容列表
  const contentList: Record<string, React.ReactNode> = {};

  allMetrics?.concat(allMetrics2)?.forEach((container: any, index: number) => {
    contentList[index.toString()] = (
      <div className={styles.monitorContainer}>
        {container?.metricGroups?.map(
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
                        (graphContainer.metrics[0].unit && type) === 'OVERVIEW'
                          ? ','
                          : ''
                      }
                      ${
                        type === 'OVERVIEW' ? graphContainer.metrics[0].key : ''
                      }
                      )`}
                    </span>
                  }
                />
              </div>
              <LineGraph
                id={`monitor-${graphContainer.name.replace(/\s+/g, '')}`}
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
    );
  });

  return (
    <Row style={{ marginTop: 16 }}>
      {allMetrics && tabList.length > 0 && (
        <Card
          tabList={tabList}
          bodyStyle={{ padding: 0 }}
          activeTabKey={activeTabKey}
          onTabChange={(key) => setActiveTabKey(key)}
        >
          {contentList[activeTabKey] || (
            <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>
              暂无数据
            </div>
          )}
        </Card>
      )}
    </Row>
  );
}
