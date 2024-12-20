import { POINT_NUMBER } from '@/constants';
import { useRequestOfMonitor } from '@/hook/useRequestOfMonitor';
import { queryMetricsReq } from '@/services';
import { Line } from '@antv/g2plot';
import { useInViewport, useUpdateEffect } from 'ahooks';
import { Empty, Spin } from 'antd';
import dayjs from 'dayjs';
import { useRef, useState } from 'react';

type MetricType = {
  description: string;
  name: string;
  key: string;
  unit: string;
};

interface LineGraphProps {
  id: string;
  metrics: MetricType[];
  labels: API.MetricsLabels;
  queryRange: Monitor.QueryRangeType;
  groupLabels: API.LableKeys[];
  height?: number;
  isRefresh?: boolean;
  type?: API.MonitorUseTarget;
  useFor: API.MonitorUseFor;
  filterData?: API.ClusterItem[] | API.TenantDetail[];
  filterQueryMetric?: API.MetricsLabels;
}

export default function LineGraph({
  id,
  metrics,
  labels,
  queryRange,
  groupLabels,
  height = 186,
  isRefresh = false,
  type = 'DETAIL',
  useFor,
  filterData,
  filterQueryMetric,
}: LineGraphProps) {
  const [isEmpty, setIsEmpty] = useState<boolean>(true);
  const [isloading, setIsloading] = useState<boolean>(true);
  const lineGraphRef = useRef(null);
  const lineInstanceRef = useRef<Line | null>(null);
  const [inViewport] = useInViewport(lineGraphRef);
  // The number of times to enter the visible area,
  // only initiate a network request when entering the visible area for the first time
  const [inViewportCount, setInViewportCount] = useState<number>(0);

  /**
   * The overview page only displays the first metrics
   * and passes empty labels to obtain all clusters/tenants.
   *
   * The details page displays all metrics and filters by labels.
   *
   * The tenant page in the cluster details needs labels to filter out which cluster the tenant is in.
   */
  const getQueryParms = () => {
    return {
      groupLabels,
      labels: type === 'OVERVIEW' ? [] : labels, // If empty, query all clusters
      metrics:
        type === 'DETAIL'
          ? metrics.map((metric: MetricType) => metric.key)
          : [metrics[0].key],
      queryRange,
      type,
      useFor,
      filterData,
      filterQueryMetric,
    };
  };

  const lineInstanceRender = (metricsData: any) => {
    const values: number[] = [];
    for (const metric of metricsData) {
      values.push(metric.value);
    }

    const config = {
      data: metricsData,
      xField: 'date',
      yField: 'value',
      height: height,
      seriesField: 'name',
      xAxis: {
        nice: false,
        tickCount: POINT_NUMBER,
        label: {
          formatter: (text: number) => {
            const time = dayjs.unix(Math.ceil(text / 1000)).format('HH:mm');
            return time;
          },
        },
      },
      tooltip: {
        title: (value: number) => {
          return dayjs
            .unix(Math.ceil(value / 1000))
            .format('YYYY-MM-DD HH:mm:ss');
        },
      },
      interactions: [{ type: 'marker-active' }, { type: 'brush' }],
    };
    if (lineInstanceRef.current) {
      // lineInstanceRef.current.update({ ...config });
      lineInstanceRef.current.changeData(metricsData);
    } else {
      lineInstanceRef.current = new Line(id, { ...config });
      lineInstanceRef.current.render();
    }
  };

  const lineInstanceDestroy = () => {
    lineInstanceRef.current?.destroy();
    lineInstanceRef.current = null;
    if (isloading) setIsloading(false);
    if (!isEmpty) setIsEmpty(true);
  };

  // filter metricsData
  const {
    data: metricsData,
    run: queryMetrics,
    // cancel: stopQueryMetrics,
  } = useRequestOfMonitor(queryMetricsReq, {
    isRealTime: isRefresh,
    manual: true,
    onSuccess: (metricsData) => {
      if (!metricsData || !metricsData.length) {
        lineInstanceDestroy();
        return;
      }

      if (metricsData && metricsData.length > 0) {
        if (isEmpty) {
          setIsEmpty(false);
          setIsloading(false);
          return;
        }
        lineInstanceRender(metricsData);
      }
      setIsloading(false);
    },
    onError: () => {
      lineInstanceDestroy();
    },
  });

  useUpdateEffect(() => {
    if (!isEmpty) {
      lineInstanceRender(metricsData);
    }
  }, [isEmpty]);

  useUpdateEffect(() => {
    if (inViewport) {
      setInViewportCount(inViewportCount + 1);
    }
  }, [inViewport]);

  useUpdateEffect(() => {
    if (inViewportCount === 1 && !isRefresh) {
      queryMetrics(getQueryParms());
    }
  }, [inViewportCount]);

  // Process after turning on real-time mode
  useUpdateEffect(() => {
    if (!isRefresh) {
      if (inViewport) {
        queryMetrics(getQueryParms());
      } else if (inViewportCount >= 1) {
        setInViewportCount(0);
      }
    } else {
      // if(timerRef.current){
      //   clearTimeout(timerRef.current)
      // }
      queryMetrics(getQueryParms());
    }
  }, [labels, queryRange]);

  return (
    <div style={{ height: `${height}px` }}>
      <Spin spinning={isloading}>
        {isEmpty ? (
          <div ref={lineGraphRef}>
            <Empty />
          </div>
        ) : (
          <div id={id} ref={lineGraphRef}></div>
        )}
      </Spin>
    </div>
  );
}
