import { Alert, Button, Card, Col, Empty, Row, Select, Space, Statistic, Typography } from 'antd';
import { request, history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useMemo, useRef, useState } from 'react';
import { Line } from '@antv/g2plot';
import { L } from '@/pages/LogService/common';
import { buildSSChartOptions } from './chartOptions';

type Series = { key: string; name: string; nameEN: string; unit: string; available: boolean; error?: string; points: { time: number; value: number }[] };
function Chart({ series }: { series: Series }) {
  const ref = useRef<HTMLDivElement>(null);
  const options = useMemo(() => buildSSChartOptions(series.points), [series.points]);
  useEffect(() => {
    if (!ref.current || !options.data.length) return;
    const plot = new Line(ref.current, options);
    plot.render(); return () => plot.destroy();
  }, [options]);
  const latest = options.data[options.data.length - 1];
  const available = series.available && !!latest;
  return <Card title={L(series.name, series.nameEN)} style={{ height: 330 }}>
    {series.error && <Alert type="error" message={series.error} />}
    {!available && <Typography.Text type="warning">{L('当前无有效样本；可能尚未采集、采集失败或没有请求。不是 0。', 'No current valid sample: collection may be pending, failing, or traffic may be absent. This is not zero.')}</Typography.Text>}
    {available && <Statistic value={latest.value} precision={2} suffix={series.unit} valueStyle={{ fontSize: 20 }} />}
    {options.data.length ? <div ref={ref} /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
  </Card>;
}
export default function SSMonitor({ kind, namespace, name }: { kind: 'logservice' | 'obcluster'; namespace: string; name: string }) {
  const [minutes, setMinutes] = useState(60);
  const path = kind === 'logservice' ? `/logservices/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/metrics` : `/obclusters/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/ss-metrics`;
  const { data, error, loading, refresh } = useRequest(async () => {
    const r = await request<{ data: Series[]; successful: boolean; message: string }>(`/api/v1${path}`, { params: { minutes } });
    if (!r.successful) throw new Error(r.message); return r.data;
  }, { refreshDeps: [kind, namespace, name, minutes], pollingInterval: 30000 });
  const failed = data?.filter(s => (s.key === 'collection' || s.key === 'registry') && (!s.available || s.points[s.points.length - 1]?.value !== 1));
  return <Space direction="vertical" size="middle" style={{ width: '100%' }}>
    <Space><Select aria-label={L('监控时间范围', 'Monitoring time range')} value={minutes} onChange={setMinutes} options={[{ value: 15, label: L('最近 15 分钟', 'Last 15 minutes') }, { value: 60, label: L('最近 1 小时', 'Last hour') }, { value: 360, label: L('最近 6 小时', 'Last 6 hours') }, { value: 1440, label: L('最近 24 小时', 'Last 24 hours') }]} /><Button loading={loading} onClick={refresh}>{L('刷新', 'Refresh')}</Button><Button onClick={() => history.push('/alert/rules')}>{L('告警规则', 'Alert rules')}</Button><Button onClick={() => history.push('/alert/event')}>{L('告警事件', 'Alert events')}</Button></Space>
    <Alert type="info" showIcon message={kind === 'logservice' ? L('数据来源：Kubernetes Pod Ready、容器重启计数、LS /manager/ln/show 注册表。30 秒采集一次。', 'Sources: Kubernetes Pod Ready, container restart counters and LS /manager/ln/show registry. Collected every 30 seconds.') : L('数据来源：GV$SYSSTAT 的 SS 专项计数器；仅统计湖库模式。命中率需要实际缓存请求，空闲时无比值。', 'Source: SS-specific GV$SYSSTAT counters, for shared-storage clusters only. Cache hit ratio requires traffic; idle periods have no ratio.')} />
    {error && <Alert type="error" message={error.message} showIcon />}
    {!!failed?.length && <Alert type="warning" showIcon message={L('采集链路异常或尚未就绪，请检查下方采集状态和告警。', 'Collection is unhealthy or not ready. Check collection health and alerts below.')} />}
    <Row gutter={[16, 16]} style={{ width: '100%' }}>{data?.map(s => <Col xs={24} xl={12} key={s.key}><Chart series={s} /></Col>)}</Row>
  </Space>;
}
