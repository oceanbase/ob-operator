import DetailLayout from '@/pages/Layouts/DetailLayout';
import { intl } from '@/utils/intl';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { useParams } from '@umijs/max';

export default () => {
  const params = useParams();
  const { ns, name, tenantName } = params;

  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.TopologyDiagram',
        defaultMessage: '拓扑图',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}/topo`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Backup',
        defaultMessage: '备份',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}/backup`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/tenant/${ns}/${name}/${tenantName}/monitor`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Connection1',
        defaultMessage: '连接租户',
      }),
      key: 'connection',
      link: `/tenant/${ns}/${name}/${tenantName}/connection`,
    },
  ];
  return <DetailLayout menus={menus} subSideSelectKey="tenant" />;
};
