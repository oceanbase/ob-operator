import logoImg from '@/assets/logo1.svg';
import MyInfoModal from '@/components/customModal/MyInfoModal';
import ResetPwdModal from '@/components/customModal/ResetPwdModal';
import { logoutReq } from '@/services';
import { getAppInfoFromStorage } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { AlertFilled, TeamOutlined } from '@ant-design/icons';
import { Menu } from '@oceanbase/design';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { BasicLayout, IconFont } from '@oceanbase/ui';
import { Outlet, history, useAccess, useLocation, useModel } from '@umijs/max';
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
  const { initialState } = useModel('@@initialState');
  const access = useAccess();
  const [version, setVersion] = useState<string>('');
  const location = useLocation();
  const { reportDataInterval } = useModel('global');
  const [resetModalVisible, setResetModalVisible] = useState<boolean>(false);
  const [infoModalVisible, setInfoModalVisible] = useState<boolean>(false);
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
      onClick={({ key }) => {
        if (key === 'logout') logout();
        if (key === 'myinfo') {
          setInfoModalVisible(true);
        }
        if (key === 'reset') {
          setResetModalVisible(true);
        }
      }}
    >
      <Menu.Item key="reset">修改密码</Menu.Item>
      <Menu.Item key="myinfo">我的信息</Menu.Item>
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
          username: initialState?.accountInfo?.nickname || 'admin',
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
      <ResetPwdModal
        visible={resetModalVisible}
        setVisible={setResetModalVisible}
      />
      <MyInfoModal
        visible={infoModalVisible}
        setVisible={setInfoModalVisible}
      />
    </div>
  );
};

export default DetailLayout;
