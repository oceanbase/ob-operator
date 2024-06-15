import { getSimpleClusterList } from '@/services';
import { getAllTenants } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { Outlet, history, useModel } from '@umijs/max';
import type { TabsProps } from 'antd';
import { Divider, Tabs } from 'antd';
import { useEffect, useState } from 'react';
import styles from './index.less';

const TAB_KEYS = ['event', 'shield', 'rules', 'channel', 'subscriptions'];

const TAB_ITEMS: TabsProps['items'] = [
  {
    key: 'event',
    label: intl.formatMessage({
      id: 'src.pages.Alert.542DC103',
      defaultMessage: '告警事件',
    }),
  },
  {
    key: 'shield',
    label: intl.formatMessage({
      id: 'src.pages.Alert.0E2168FC',
      defaultMessage: '告警屏蔽',
    }),
  },
  {
    key: 'divider-1',
    label: <Divider type="vertical" />,
  },
  {
    key: 'rules',
    label: intl.formatMessage({
      id: 'src.pages.Alert.6261133D',
      defaultMessage: '告警规则',
    }),
  },
  {
    key: 'divider-2',
    label: <Divider type="vertical" />,
  },
  {
    key: 'channel',
    label: intl.formatMessage({
      id: 'src.pages.Alert.5208FED8',
      defaultMessage: '告警通道',
    }),
  },
  {
    key: 'subscriptions',
    label: intl.formatMessage({
      id: 'src.pages.Alert.8F65A789',
      defaultMessage: '告警推送',
    }),
  },
];

const getInitialActiveKey = () => {
  const hashFrags = location.hash.split('/');
  const tailKey = hashFrags[hashFrags.length - 1];
  if (TAB_KEYS.includes(tailKey)) return tailKey;
  return 'event';
};

export default function Alert() {
  const { setClusterList, setTenantList } = useModel('alarm');
  const [activeKey, setActiveKey] = useState<string>(getInitialActiveKey());
  const onChange = (key: string) => {
    setActiveKey(key);
    history.push(`/alert/${key}`);
  };

  useEffect(() => {
    getSimpleClusterList().then(({ successful, data }) => {
      if (successful) setClusterList(data);
    });
    getAllTenants().then(({ successful, data }) => {
      if (successful) setTenantList(data);
    });
    const unlisten = history.listen(({ location }) => {
      const curKey =
        location.pathname.split('/')[location.pathname.split('/').length - 1];
      if (curKey !== activeKey) {
        setActiveKey(curKey);
      }
    });
    return () => {
      unlisten();
    };
  }, []);
  return (
    <PageContainer>
      <Tabs
        activeKey={activeKey}
        className={styles.tabContent}
        items={TAB_ITEMS}
        onChange={onChange}
      />

      <Outlet />
    </PageContainer>
  );
}
