import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { intl } from '@/utils/intl';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { IconFont, BasicLayout as OBLayout } from '@oceanbase/ui';
import { Outlet, history, useLocation } from '@umijs/max';
import { useRequest } from 'ahooks';

const BasicLayout: React.FC = () => {
  const location = useLocation();
  const user = localStorage.getItem('user');
  const { run: logout } = useRequest(logoutReq, {
    manual: true,
    onSuccess: (data) => {
      if (data.successful) {
        history.push('/login');
      }
    },
  });

  // const Title = () => (
  //   <img
  //     style={{ height: 20, marginLeft: -16, paddingLeft: 6,cursor:'pointer' }}
  //     onClick={()=>history.push('/')}
  //     src="https://www.gartner.com/pi/vendorimages/oceanbase_1640501555454.png"
  //   />
  // );

  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'dashboard.Layouts.BasicLayout.Overview',
        defaultMessage: '概览',
      }),
      link: '/overview',
      icon: <IconFont type="overview" />,
    },
    {
      title: intl.formatMessage({
        id: 'dashboard.Layouts.BasicLayout.Cluster',
        defaultMessage: '集群',
      }),
      link: '/cluster',
      icon: <IconFont type="cluster" />,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Layouts.BasicLayout.Tenant',
        defaultMessage: '租户',
      }),
      link: '/tenant',
      icon: <IconFont type="tenant" />,
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
      <OBLayout
        logoUrl={logoImg}
        simpleLogoUrl={logoImg}
        menus={menus}
        defaultSelectedKeys={['/overview']}
        location={location}
        topHeader={{
          username: user || 'admin',
          userMenu,
          showLocale: true,
          locales: ['zh-CN', 'en-US'],
          appData: {
            shortName: 'ob dashboard',
            version: '1.0.0',
          },
        }}
      >
        <Outlet />
      </OBLayout>
    </div>
  );
};

export default BasicLayout;
