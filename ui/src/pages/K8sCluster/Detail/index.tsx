import DetailLayout from '@/pages/Layouts/DetailLayout';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { useParams } from '@umijs/max';

export default () => {
  const params = useParams();
  const { k8sclusterName } = params;
  const menus: MenuItem[] = [
    {
      title: '概览',
      link: `/k8scluster/${k8sclusterName}`,
    },
  ];
  return <DetailLayout menus={menus} subSideSelectKey="k8scluster" />;
};
