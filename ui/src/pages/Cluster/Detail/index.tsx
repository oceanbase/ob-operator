import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { getAppInfoFromStorage } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { BasicLayout, IconFont } from '@oceanbase/ui';
import { Outlet, history, useLocation, useModel, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
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
  {
    title: intl.formatMessage({
      id: 'Dashboard.Cluster.Detail.Tenant',
      defaultMessage: '租户',
    }),
    key: 'tenant',
    link: '/tenant',
    icon: <IconFont type="tenant" />,
  },
];

const ClusterDetail: React.FC = () => {
  const params = useParams();
  const user = localStorage.getItem('user');
  const [version, setVersion] = useState<string>('');
  const location = useLocation();
  const { reportDataInterval } = useModel('global');
  const { ns, name, clusterName } = params;
  const { run: logout } = useRequest(logoutReq, {
    manual: true,
    onSuccess: (data) => {
      if (data.successful) {
        clearInterval(reportDataInterval.current);
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
        id: 'Dashboard.Cluster.Detail.Connection',
        defaultMessage: '连接集群',
      }),
      key: 'connection',
      link: `/cluster/${ns}/${name}/${clusterName}/connection`,
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

  useEffect(() => {
    getAppInfoFromStorage().then((appInfo) => {
      setVersion(appInfo.version);
    });
  }, []);

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
            version,
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
