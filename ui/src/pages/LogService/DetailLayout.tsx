import DetailLayout from '@/pages/Layouts/DetailLayout';
import type { MenuItem } from '@oceanbase/ui/es/BasicLayout';
import { useParams } from '@umijs/max';
import { Typography } from 'antd';
import { L } from './common';

export default function LogServiceDetailLayout() {
  const { ns = '', name = '' } = useParams();
  const base = `/logservice/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`;
  const menus: MenuItem[] = [
    { title: L('概览与拓扑', 'Overview & topology'), link: `${base}/overview` },
    { title: L('日志对象存储', 'Log object storage'), link: `${base}/storage` },
    { title: L('LS 专项监控', 'LS monitoring'), link: `${base}/monitor` },
    { title: L('启动参数', 'Startup parameters'), link: `${base}/parameters` },
    { title: L('事件', 'Events'), link: `${base}/events` },
  ];

  return (
    <DetailLayout
      menus={menus}
      subSideSelectKey="logservice"
      sideHeader={
        <div style={{ padding: '16px 24px', overflowWrap: 'anywhere' }}>
          <Typography.Text strong>LogService · {name}</Typography.Text>
          <br />
          <Typography.Text type="secondary">{ns}</Typography.Text>
        </div>
      }
    />
  );
}
