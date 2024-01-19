import { ConfigProvider } from '@oceanbase/design';
import { Outlet, useNavigate } from '@umijs/max';
import { Layout } from 'antd';
import type { Locale } from 'antd/es/locale';
import zhCN from 'antd/locale/zh_CN';
import { useState } from 'react';

import styles from './index.less';

//前置处理
const PreLayout: React.FC = () => {
  const [locale] = useState<Locale>(zhCN);
  const navigate = useNavigate();

  return (
    <ConfigProvider navigate={navigate}>
      <div className={styles.rootContainer}>
        <Layout className={styles.layout}>
          <Outlet />
        </Layout>
      </div>
    </ConfigProvider>
  );
};

export default PreLayout;
