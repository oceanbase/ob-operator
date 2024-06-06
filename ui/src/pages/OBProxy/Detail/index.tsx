import DetailLayout from '@/pages/Layouts/DetailLayout';
import { intl } from '@/utils/intl';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { useParams } from '@umijs/max';
export default function Detail() {
  const params = useParams();
  const { ns, name } = params;
  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/obproxy/${ns}/${name}`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/obproxy/${ns}/${name}/monitor`,
    },
  ];
  return <DetailLayout menus={menus} subSideSelectKey="obproxy" />;
}
