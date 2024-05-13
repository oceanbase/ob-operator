import { PageContainer } from '@ant-design/pro-components';
import { Outlet, history } from '@umijs/max';
import type { TabsProps } from 'antd';
import { Divider, Tabs } from 'antd';
import styles from './index.less'

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

const onChange = (key: string) => {
  history.push(`/alert/${key}`);
};

export default function Alert() {
  return (
    <PageContainer>
      <Tabs className={styles.tabContent} defaultActiveKey="1" items={items} onChange={onChange} />
      <Outlet />
    </PageContainer>
  );
}
