import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { intl } from '@/utils/intl';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { BasicLayout, IconFont } from '@oceanbase/ui';
import { Outlet, history, useLocation, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
const subSideMenus: MenuItem[] = [
  {
    title: intl.formatMessage({
      id: 'dashboard.Cluster.Detail.Overview',
      defaultMessage: '概览',
    }),
    key: 'overview',
    link: '/overview',
    icon: <IconFont type="overview" />,
  },
  {
    title: intl.formatMessage({
      id: 'dashboard.Cluster.Detail.Cluster',
      defaultMessage: '集群',
    }),
    key: 'cluster',
    link: '/cluster',
    icon: <IconFont type="cluster" />,
  },
];

const ClusterDetail: React.FC = () => {
  const params = useParams();
  const user = localStorage.getItem('user');
  const location = useLocation();
  const { clusterId } = params;
  const { run: logout } = useRequest(logoutReq, {
    manual: true,
    onSuccess: (data) => {
      if (data.successful) {
        history.push('/login');
      }
    },
  });
  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'dashboard.Cluster.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/cluster/${clusterId}`,
    },
    {
      title: intl.formatMessage({
        id: 'dashboard.Cluster.Detail.TopologyDiagram',
        defaultMessage: '拓扑图',
      }),
      link: `/cluster/${clusterId}/topo`,
    },
    {
      title: intl.formatMessage({
        id: 'OBDashboard.Cluster.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/cluster/${clusterId}/monitor`,
    },
  ];
  const userMenu = (
    <Menu
      onClick={() => {
        logout();
      }}
    >
      <Menu.Item key="logout">
        {intl.formatMessage({
          id: 'dashboard.Layouts.BasicLayout.LogOut',
          defaultMessage: '退出登录',
        })}
      </Menu.Item>
    </Menu>
  );
  return (
    <div>
      <BasicLayout
        logoUrl={logoImg}
        simpleLogoUrl={logoImg}
        topHeader={{
          username: user || '',
          userMenu,
          showLocale: true,
          locales: ['zh-CN', 'en-US'],
          appData: {
            shortName: 'ob dashboard',
            version: '1.0.0',
          },
        }}
        menus={menus}
        location={location}
        subSideMenus={subSideMenus}
        subSideMenuProps={{ selectedKeys: ['/cluster'] }}
      >
        <Outlet />
      </BasicLayout>
    </div>
  );
};

export default ClusterDetail;
