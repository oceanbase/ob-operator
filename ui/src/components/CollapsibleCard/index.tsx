import { Card, Space, Switch } from 'antd';
import { useState } from 'react';

export default function CollapsibleCard({
  collapsible,
  children,
  title,
  defaultExpand=false,
  ...props
}: any) {
  const [isExpand, setIsExpand] = useState(defaultExpand);
  return (
    <Card
      title={
        <Space>
          {title}
          {collapsible && (
            <Switch
              onChange={() => setIsExpand(!isExpand)}
              checked={isExpand}
            />
          )}
        </Space>
      }
      bodyStyle={((collapsible && isExpand) || !collapsible) ? { padding: 24 } : { padding: 0 }}
      {...props}
    >
      {((collapsible && isExpand) || !collapsible) && children}
    </Card>
  );
}
