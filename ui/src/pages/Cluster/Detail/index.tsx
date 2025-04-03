import DetailLayout from '@/pages/Layouts/DetailLayout';
import { intl } from '@/utils/intl';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { useAccess, useParams } from '@umijs/max';

export default () => {
  const params = useParams();
  const { ns, name, clusterName } = params;
  const access = useAccess();
  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'dashboard.Cluster.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/cluster/${ns}/${name}/${clusterName}`,
    },
    {
      title: intl.formatMessage({
        id: 'dashboard.Cluster.Detail.TopologyDiagram',
        defaultMessage: '拓扑图',
      }),
      link: `/cluster/${ns}/${name}/${clusterName}/topo`,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Cluster.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/cluster/${ns}/${name}/${clusterName}/monitor`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Cluster.Detail.Tenant',
        defaultMessage: '租户',
      }),
      key: 'tenant',
      link: `/cluster/${ns}/${name}/${clusterName}/tenant`,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.C05D24A0',
        defaultMessage: '集群参数',
      }),
      key: 'parameters',
      link: `/cluster/${ns}/${name}/${clusterName}/parameters`,
      accessible: access.obclusterwrite,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Cluster.Detail.Connection1',
        defaultMessage: '连接集群',
      }),
      key: 'connection',
      link: `/cluster/${ns}/${name}/${clusterName}/connection`,
      accessible: access.obclusterwrite,
    },
  ];

  return <DetailLayout menus={menus} subSideSelectKey="cluster" />;
};
