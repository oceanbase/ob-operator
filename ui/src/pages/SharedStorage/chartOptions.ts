import type { LineOptions } from '@antv/g2plot';

export type MetricPoint = { time: number; value: number };

export function buildSSChartOptions(points: MetricPoint[]): LineOptions {
  // Prometheus uses Unix seconds; G2's time scale expects milliseconds.
  const data = points
    .filter(p => Number.isFinite(p.time) && p.time > 0 &&
      Number.isFinite(new Date(p.time * 1000).getTime()) && Number.isFinite(p.value))
    .map(p => ({ time: p.time * 1000, value: p.value }));

  return {
    data,
    xField: 'time',
    yField: 'value',
    height: 200,
    // Axis label callbacks receive formatted text, not a Unix timestamp.
    meta: { time: { type: 'time', mask: 'HH:mm:ss', tickCount: 5 } },
    tooltip: {
      title: (_text, datum) => {
        const time = datum?.time;
        return typeof time === 'number' && Number.isFinite(new Date(time).getTime())
          ? new Date(time).toLocaleString()
          : '-';
      },
    },
    yAxis: { min: 0 },
    animation: false,
    connectNulls: false,
  };
}
