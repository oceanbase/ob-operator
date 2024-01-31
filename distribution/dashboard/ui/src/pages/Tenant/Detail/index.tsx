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
    title: '概览',
    key: 'overview',
    link: '/overview',
    icon: <IconFont type="overview" />,
  },
  {
    title: '集群',
    key: 'cluster',
    link: '/cluster',
    icon: <IconFont type="cluster" />,
  },
  {
    title: '租户',
    key: 'tenant',
    link: '/tenant',
    icon: <IconFont type="tenant" />,
  },
];

const TenantDetail: React.FC = () => {
  const params = useParams();
  const user = localStorage.getItem('user');
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
      title: '概览',
      link: `/tenant/${tenantId}`,
    },
    {
      title:'拓扑图',
      link: `/tenant/${tenantId}/topo`,
    },
    {
      title:'备份',
      link:`/tenant/${tenantId}/backup`
    }
    // {
    //   title: '性能监控',
    //   key: 'monitor',
    //   link: `/tenant/${clusterId}/monitor`,
    // },
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
        subSideMenuProps={{ selectedKeys: ['/tenant'] }}
      >
        <Outlet />
      </BasicLayout>
    </div>
  );
};

export default TenantDetail;
