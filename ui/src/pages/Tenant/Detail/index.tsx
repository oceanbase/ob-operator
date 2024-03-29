import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { getAppInfoFromStorage } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { BasicLayout, IconFont } from '@oceanbase/ui';
import { Outlet, history, useLocation, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
const subSideMenus: MenuItem[] = [
  {
    title: intl.formatMessage({
      id: 'Dashboard.Tenant.Detail.Overview',
      defaultMessage: '概览',
    }),
    key: 'overview',
    link: '/overview',
    icon: <IconFont type="overview" />,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Tenant.Detail.Cluster',
      defaultMessage: '集群',
    }),
    key: 'cluster',
    link: '/cluster',
    icon: <IconFont type="cluster" />,
  },
  {
    title: intl.formatMessage({
      id: 'Dashboard.Tenant.Detail.Tenant',
      defaultMessage: '租户',
    }),
    key: 'tenant',
    link: '/tenant',
    icon: <IconFont type="tenant" />,
  },
];

const TenantDetail: React.FC = () => {
  const params = useParams();
  const user = localStorage.getItem('user');
  const [version, setVersion] = useState<string>('');
  const location = useLocation();
  const { tenantId } = params;

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
        id: 'Dashboard.Tenant.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/tenant/${tenantId}`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.TopologyDiagram',
        defaultMessage: '拓扑图',
      }),
      link: `/tenant/${tenantId}/topo`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Backup',
        defaultMessage: '备份',
      }),
      link: `/tenant/${tenantId}/backup`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/tenant/${tenantId}/monitor`,
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
        subSideMenuProps={{ selectedKeys: ['/tenant'] }}
      >
        <Outlet />
      </BasicLayout>
    </div>
  );
};

export default TenantDetail;
