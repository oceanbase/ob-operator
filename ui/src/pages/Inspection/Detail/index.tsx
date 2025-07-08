import DetailLayout from '@/pages/Layouts/DetailLayout';
import { intl } from '@/utils/intl';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { useParams } from '@umijs/max';

export default () => {
  const params = useParams();
  const { inspectionName } = params;
  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'src.pages.K8sCluster.Detail.2369D746',
        defaultMessage: '概览',
      }),
      link: `/inspection/${inspectionName}`,
    },
  ];

  return <DetailLayout menus={menus} subSideSelectKey="k8scluster" />;
};
