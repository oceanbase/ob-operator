import { PageContainer } from '@ant-design/pro-components';
import { Outlet, history } from '@umijs/max';
import type { TabsProps } from 'antd';
import { Divider, Tabs } from 'antd';
import { useEffect, useState } from 'react';
import styles from './index.less';

const items: TabsProps['items'] = [
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
export default function Alert() {
  const [activeKey, setActiveKey] = useState<string>('event');
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
        items={items}
        onChange={onChange}
      />
      <Outlet />
    </PageContainer>
  );
}
