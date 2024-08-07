import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { getAppInfoFromStorage } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { AlertFilled, TeamOutlined } from '@ant-design/icons';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { IconFont, BasicLayout as OBLayout } from '@oceanbase/ui';
import { Outlet, history, useAccess, useLocation, useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';

const BasicLayout: React.FC = () => {
  const location = useLocation();
  const { initialState } = useModel('@@initialState');
  const [version, setVersion] = useState<string>('');
  const { reportDataInterval } = useModel('global');
  const access = useAccess();
  const { run: logout } = useRequest(logoutReq, {
    manual: true,
    onSuccess: (data) => {
      if (data.successful) {
        history.push('/login');
        clearInterval(reportDataInterval.current);
      }
    },
  });

  useEffect(() => {
    getAppInfoFromStorage().then((appInfo) => {
      setVersion(appInfo?.version);
    });
  }, []);

  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'dashboard.Layouts.BasicLayout.Overview',
        defaultMessage: '概览',
      }),
      link: '/overview',
      icon: <IconFont type="overview" />,
      accessible: access.systemread,
    },
    {
      title: intl.formatMessage({
        id: 'dashboard.Layouts.BasicLayout.Cluster',
        defaultMessage: '集群',
      }),
      link: '/cluster',
      icon: <IconFont type="cluster" />,
      accessible: access.obclusterread,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Layouts.BasicLayout.Tenant',
        defaultMessage: '租户',
      }),
      link: '/tenant',
      icon: <IconFont type="tenant" />,
      accessible: access.obclusterread,
    },
    {
      title: 'OBProxy',
      link: '/obproxy',
      icon: <IconFont type="obproxy" />,
      accessible: access.obproxyread,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Layouts.BasicLayout.EDA4D7D9',
        defaultMessage: '告警',
      }),
      link: '/alert',
      icon: <AlertFilled style={{ color: 'rgb(109,120,147)' }} />,
      accessible: access.alarmread,
    },
    {
      title: '权限控制',
      link: '/access',
      icon: <TeamOutlined style={{ color: 'rgb(109,120,147)' }} />,
      accessible: access.acread,
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
    history.replace(menus.find((item) => item.accessible)?.link || '/overview');
  }, []);

  return (
    <div>
      <OBLayout
        logoUrl={logoImg}
        simpleLogoUrl={logoImg}
        menus={menus}
        defaultSelectedKeys={[menus.find((item) => item.accessible)?.link || '/overview']}
        location={location}
        topHeader={{
          username: initialState?.accountInfo?.nickname || 'admin',
          userMenu,
          showLocale: true,
          locales: ['zh-CN', 'en-US'],
          appData: {
            shortName: 'ob dashboard',
            version,
          },
        }}
      >
        <Outlet />
      </OBLayout>
    </div>
  );
};

export default BasicLayout;
