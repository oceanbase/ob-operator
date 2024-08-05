import logoImg from '@/assets/logo1.svg';
import { logoutReq } from '@/services';
import { getAppInfoFromStorage } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { AlertFilled, TeamOutlined } from '@ant-design/icons';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { BasicLayout, IconFont } from '@oceanbase/ui';
import { Outlet, history, useLocation, useModel, useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';

interface DetailLayoutProps {
  subSideSelectKey: string;
  menus: MenuItem[];
}



const DetailLayout: React.FC<DetailLayoutProps> = ({
  subSideSelectKey,
  menus,
}) => {
  const user = localStorage.getItem('user');
  const access = useAccess();
  const [version, setVersion] = useState<string>('');
  const location = useLocation();
  const { reportDataInterval } = useModel('global');
  const { run: logout } = useRequest(logoutReq, {
    manual: true,
    onSuccess: (data) => {
      if (data.successful) {
        clearInterval(reportDataInterval.current);
        history.push('/login');
      }
    },
  });

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
      accessible: access.obclusterread,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Cluster.Detail.Tenant',
        defaultMessage: '租户',
      }),
      key: 'tenant',
      link: '/tenant',
      icon: <IconFont type="tenant" />,
      accessible: access.obclusterread,
    },
    {
      title: 'OBProxy',
      key: 'obproxy',
      link: '/obproxy',
      icon: <IconFont type="obproxy" />,
      accessible: access.obproxyread,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Layouts.DetailLayout.BC23B91F',
        defaultMessage: '告警',
      }),
      key: 'alert',
      link: '/alert',
      icon: <AlertFilled style={{ color: 'rgb(109,120,147)' }} />,
      accessible: access.alarmread,
    },
    {
      title: '权限控制',
      key: 'access',
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
    getAppInfoFromStorage().then((appInfo) => {
      setVersion(appInfo?.version);
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
        subSideMenuProps={{ selectedKeys: [`/${subSideSelectKey}`] }}
      >
        <Outlet />
      </BasicLayout>
    </div>
  );
};

export default DetailLayout;
