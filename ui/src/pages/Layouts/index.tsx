import { ConfigProvider } from '@oceanbase/design';
import { getAppInfoFromStorage } from '@/utils/helper';
import enUS from '@oceanbase/ui/es/locale/en-US';
import zhCN from '@oceanbase/ui/es/locale/zh-CN';
import { Outlet, getLocale, useNavigate, useModel,history } from '@umijs/max';
import { Layout } from 'antd';

import { useEffect } from 'react';
import styles from './index.less';

const PreLayout: React.FC = () => {
  const navigate = useNavigate();
  const { initialState } = useModel('@@initialState');
  const locale = getLocale() || 'zh-CN';
  const localeMap = {
    'zh-CN': zhCN,
    'en-US': enUS,
  };
  
  useEffect(() => {
    if(initialState?.accountInfo?.needReset){
      history.replace('/reset')
    }
    getAppInfoFromStorage().then((appInfo) => {
      sessionStorage.setItem('appInfo', JSON.stringify(appInfo));
    });
  }, []);

  return (
    <ConfigProvider locale={localeMap[locale]} navigate={navigate}>
      <div className={styles.rootContainer}>
        <Layout className={styles.layout}>
          <Outlet />
        </Layout>
      </div>
    </ConfigProvider>
  );
};

export default PreLayout;
