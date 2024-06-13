import { Tooltip } from 'antd';
import { ReactNode } from 'react';
import styles from './index.less';

interface TooltipPrettyProps {
  title: ReactNode;
  children: ReactNode;
}

export default function TooltipPretty({ title, children }: TooltipPrettyProps) {
  return (
    <Tooltip
      color="#fff"
      overlayInnerStyle={{ color: 'rgba(0,0,0,.85)' }}
      overlayClassName={styles.toolTipContent}
      placement="topLeft"
      title={title}
    >
      {children}
    </Tooltip>
  );
}
