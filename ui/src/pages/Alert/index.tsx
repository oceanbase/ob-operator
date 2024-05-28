import { PageContainer } from '@ant-design/pro-components';
import { Outlet, history } from '@umijs/max';
import type { TabsProps } from 'antd';
import { Divider, Tabs } from 'antd';
import { useEffect, useState } from 'react';
import styles from './index.less';

const TAB_KEYS = ['event', 'shield', 'rules', 'channel', 'subscriptions'];

const TAB_ITEMS: TabsProps['items'] = [
  {
    key: 'event',
    label: '告警事件',
  },
  {
    key: 'shield',
    label: '告警屏蔽',
  },
  {
    key: 'divider-1',
    label: <Divider type="vertical" />,
  },
  {
    key: 'rules',
    label: '告警规则',
  },
  {
    key: 'divider-2',
    label: <Divider type="vertical" />,
  },
  {
    key: 'channel',
    label: '告警通道',
  },
  {
    key: 'subscriptions',
    label: '告警推送',
  },
];

const getInitialActiveKey = () => {
  const hashFrags = location.hash.split('/');
  const tailKey = hashFrags[hashFrags.length - 1];
  if (TAB_KEYS.includes(tailKey)) return tailKey;
  return 'event';
};

export default function Alert() {
  const [activeKey, setActiveKey] = useState<string>(getInitialActiveKey());
  const onChange = (key: string) => {
    setActiveKey(key);
    history.push(`/alert/${key}`);
  };

  useEffect(() => {
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
