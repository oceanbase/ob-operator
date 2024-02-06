import { POINT_NUMBER } from '@/constants';
import { useRequestOfMonitor } from '@/hook/useRequestOfMonitor';
import type { QueryRangeType } from '@/pages/Cluster/Detail/Monitor';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { queryMetricsReq } from '@/services';
import { Line } from '@antv/g2plot';
import { useInViewport, useUpdateEffect } from 'ahooks';
import { Empty, Spin } from 'antd';
import _ from 'lodash';
import moment from 'moment';
import { useRef, useState } from 'react';

export type MetricType = {
  description: string;
  name: string;
  key: string;
  unit: string;
};

export interface LineGraphProps {
  id: string;
  metrics: MetricType[];
  labels: API.MetricsLabels;
  queryRange: QueryRangeType;
  height?: number;
  isRefresh?: boolean;
  type?: 'detail' | 'overview';
}

export default function LineGraph({
  id,
  metrics,
  labels,
  queryRange,
  height = 186,
  isRefresh = false,
  type,
}: LineGraphProps) {
  const [, chooseClusterName] = getNSName();
  const [isEmpty, setIsEmpty] = useState<boolean>(true);
  const [isloading, setIsloading] = useState<boolean>(true);
  const lineGraphRef = useRef(null);
  const lineInstanceRef = useRef<Line | null>(null);
  const [inViewport] = useInViewport(lineGraphRef);
  // 进入可见区域次数,只在第一次进入可见区域发起网络请求
  const [inViewportCount, setInViewportCount] = useState<number>(0);
  const groupLabels = _.uniq(labels.map((label) => label.key));

  const getQueryParms = () => {
    let metricsKeys: string[] = [metrics[0].key],
      realLabels = labels;
    if (chooseClusterName) {
      metricsKeys = metrics.map((metric: MetricType) => metric.key);
    }
    if (type === 'overview') realLabels = [];
    return {
      groupLabels,
      labels: realLabels, //为空则查询全部集群
      metrics: metricsKeys,
      queryRange,
    };
  };

  const lineInstanceRender = (metricsData: any) => {
    let values: number[] = [];
    for (let metric of metricsData) {
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
            let time = moment.unix(Math.ceil(text / 1000)).format('HH:mm');
            return time;
          },
        },
      },
      tooltip: {
        title: (value: number) => {
          return moment
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

  //开启实时模式后处理
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
