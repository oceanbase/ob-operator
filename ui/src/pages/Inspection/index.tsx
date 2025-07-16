import { PageContainer } from '@ant-design/pro-components';
import { Card } from 'antd';
import { useState } from 'react';
import HistoryList from './HistoryList';
import InspectionList from './InspectionList';

export default function Inspection() {
  const tabList = [
    {
      key: 'list',
      tab: '巡检列表',
    },
    {
      key: 'history',
      tab: '巡检历史',
    },
  ];

  const contentList: Record<string, React.ReactNode> = {
    list: <InspectionList />,
    history: <HistoryList />,
  };
  const [activeTabKey, setActiveTabKey] = useState<string>('list');
  const onTabChange = (key: string) => {
    setActiveTabKey(key);
  };
  return (
    <PageContainer title="巡检" ghost={true}>
      <Card
        style={{ width: '100%' }}
        tabList={tabList}
        activeTabKey={activeTabKey}
        onTabChange={onTabChange}
      >
        {contentList[activeTabKey]}
      </Card>
    </PageContainer>
  );
}
